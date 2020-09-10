package epochmgr

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/dapperlabs/flow-go/engine"
	"github.com/dapperlabs/flow-go/engine/collection/proposal"
	chainsync "github.com/dapperlabs/flow-go/engine/collection/synchronization"
	"github.com/dapperlabs/flow-go/module"
	"github.com/dapperlabs/flow-go/module/mempool"
	"github.com/dapperlabs/flow-go/state/cluster"
	"github.com/dapperlabs/flow-go/state/protocol"
	"github.com/dapperlabs/flow-go/storage"
)

// all requirements for participating in the cluster chain for one epoch
type epochreqs struct {
	state    cluster.State
	proposal *proposal.Engine
	sync     *chainsync.Engine
	hotstuff module.HotStuff
	// TODO: ingest/txpool should also be epoch-dependent, possibly managed by this engine
}

// Engine is the epoch manager, which coordinates the lifecycle of other modules
// that are epoch-dependent. The manager is responsible for spinning up engines
// when a new epoch is about to start and spinning down engines for an epoch that
// has ended.
type Engine struct {
	unit  *engine.Unit
	epoch *epochreqs // requirements for the current epoch

	log   zerolog.Logger
	me    module.Local
	state protocol.State

	// TODO should be per-epoch eventually, cache here for now
	pool mempool.Transactions

	// factories for building new engines for a new epoch
	clusterStateFactory *ClusterStateFactory
	builderFactory      *BuilderFactory
	proposalFactory     *ProposalEngineFactory
	syncFactory         *SyncEngineFactory
	hotstuffFactory     *HotStuffFactory
}

func New(
	log zerolog.Logger,
	me module.Local,
	state protocol.State,
	pool mempool.Transactions,
	clusterStateFactory *ClusterStateFactory,
	builderFactory *BuilderFactory,
	proposalFactory *ProposalEngineFactory,
	syncFactory *SyncEngineFactory,
	hotstuffFactory *HotStuffFactory,
) (*Engine, error) {

	e := &Engine{
		unit:                engine.NewUnit(),
		log:                 log,
		me:                  me,
		state:               state,
		pool:                pool,
		clusterStateFactory: clusterStateFactory,
		builderFactory:      builderFactory,
		proposalFactory:     proposalFactory,
		syncFactory:         syncFactory,
		hotstuffFactory:     hotstuffFactory,
	}

	// get the current epoch
	epoch, err := e.state.Final().Epoch()
	if err != nil {
		return nil, fmt.Errorf("could not get current epoch number: %w", err)
	}

	reqs, err := e.setupEpoch(epoch)
	if err != nil {
		return nil, fmt.Errorf("could not setup requirements for epoch (%d): %w", epoch, err)
	}

	e.epoch = reqs
	_ = e.state       // TODO lint
	_ = e.epoch.state // TODO lint
	return e, nil
}

// Ready returns a ready channel that is closed once the engine has fully
// started. For proposal engine, this is true once the underlying consensus
// algorithm has started.
func (e *Engine) Ready() <-chan struct{} {
	return e.unit.Ready(func() {
		<-e.epoch.hotstuff.Ready()
		<-e.epoch.proposal.Ready()
		<-e.epoch.sync.Ready()
	})
}

// Done returns a done channel that is closed once the engine has fully stopped.
func (e *Engine) Done() <-chan struct{} {
	return e.unit.Done(func() {
		<-e.epoch.hotstuff.Done()
		<-e.epoch.proposal.Done()
		<-e.epoch.sync.Done()
	})
}

// setupEpoch sets up cluster state and HotStuff for a new chain for the given
// epoch. This can be used for in-progress chains (for example, when restarting
// mid-epoch) or to bootstrap the chain for a new epoch.
func (e *Engine) setupEpoch(epoch uint64) (*epochreqs, error) {

	clusterState, headers, payloads, blocks, err := e.createClusterState(epoch)
	if err != nil {
		return nil, fmt.Errorf("could not create cluster state: %w", err)
	}

	// determine this node's cluster for the epoch
	clusters, err := e.state.AtEpoch(epoch).Clusters()
	if err != nil {
		return nil, fmt.Errorf("could not get clusters for epoch: %w", err)
	}
	cluster, _, ok := clusters.ByNodeID(e.me.NodeID())
	if !ok {
		return nil, fmt.Errorf("could not find my cluster")
	}

	// retrieve the root block and QC for the epoch
	root, err := e.state.AtEpoch(epoch).ClusterRootBlock(cluster)
	if err != nil {
		return nil, fmt.Errorf("could not get cluster root block: %w", err)
	}
	qc, err := e.state.AtEpoch(epoch).ClusterRootQC(cluster)
	if err != nil {
		return nil, fmt.Errorf("could not get cluster root qc: %w", err)
	}

	clusterID := root.Header.ChainID

	builder, finalizer, err := e.builderFactory.Create(headers, payloads, e.pool)
	if err != nil {
		return nil, fmt.Errorf("could not create builder/finalizer: %w", err)
	}

	// TODO need a protocol state method for this - for now fake it with root ID
	//seed, err := e.state.AtEpoch(epoch).LeaderSelectionSeed()
	rootID := root.ID()
	seed := rootID[:]

	proposalEngine, err := e.proposalFactory.Create(clusterState, headers, payloads)
	if err != nil {
		return nil, fmt.Errorf("could not create proposal engine: %w", err)
	}
	syncCore, syncEngine, err := e.syncFactory.Create(cluster, clusterState, blocks, proposalEngine)
	if err != nil {
		return nil, fmt.Errorf("could not create sync engine: %w", err)
	}
	hotstuff, err := e.hotstuffFactory.Create(
		clusterID,
		cluster,
		clusterState,
		headers,
		payloads,
		seed,
		builder,
		finalizer,
		proposalEngine,
		root.Header,
		qc,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create hotstuff: %w", err)
	}

	// attach dependencies to the proposal engine
	proposalEngine = proposalEngine.
		WithHotStuff(hotstuff).
		WithSync(syncCore)

	engines := &epochreqs{
		proposal: proposalEngine,
		sync:     syncEngine,
		hotstuff: hotstuff,
	}
	return engines, nil
}

func (e *Engine) createClusterState(epoch uint64) (cluster.State, storage.Headers, storage.ClusterPayloads, storage.ClusterBlocks, error) {

	// determine this node's cluster for the epoch
	clusters, err := e.state.AtEpoch(epoch).Clusters()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not get clusters for epoch: %w", err)
	}
	cluster, _, ok := clusters.ByNodeID(e.me.NodeID())
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("could not find my cluster")
	}

	// retrieve the root block and QC for the epoch
	root, err := e.state.AtEpoch(epoch).ClusterRootBlock(cluster)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not get cluster root block: %w", err)
	}

	clusterID := root.Header.ChainID

	// create the cluster state
	clusterState, headers, payloads, blocks, err := e.clusterStateFactory.Create(clusterID)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not create cluster state: %w", err)
	}
	_, err = clusterState.Final().Head()
	// storage layer error while checking state - fail fast
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, nil, nil, nil, fmt.Errorf("could not check cluster state db: %w", err)
	}
	// the cluster state for this epoch has already been bootstrapped
	if err == nil {
		return clusterState, headers, payloads, blocks, nil
	}

	// no existing cluster state, bootstrap with root block for epoch
	err = clusterState.Mutate().Bootstrap(root)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not bootstrap cluster state: %w", err)
	}

	return clusterState, headers, payloads, blocks, nil
}