// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package badger

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/storage/badger/operation"
	"github.com/dapperlabs/flow-go/storage/badger/procedure"
	"github.com/dapperlabs/flow-go/utils/unittest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var identities = flow.IdentityList{
	{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
	{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
	{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
	{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
}

var genesis = flow.Genesis(identities)

func testWithBootstraped(t *testing.T, f func(t *testing.T, mutator *Mutator, db *badger.DB)) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		err := mutator.Bootstrap(genesis)
		require.Nil(t, err)

		f(t, mutator, db)
	})
}

func TestBootStrapValid(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		var boundary uint64
		err := db.View(operation.RetrieveBoundary(&boundary))
		require.Nil(t, err)

		var storedID flow.Identifier
		err = db.View(operation.RetrieveNumber(0, &storedID))
		require.Nil(t, err)

		var storedHeader flow.Header
		err = db.View(operation.RetrieveHeader(genesis.ID(), &storedHeader))
		require.Nil(t, err)

		var storedSeal flow.Seal
		err = db.View(procedure.LookupSealByBlock(storedHeader.ID(), &storedSeal))
		require.NoError(t, err)
		require.Equal(t, flow.ZeroID, storedSeal.BlockID) //genesis seal is special

		assert.Zero(t, boundary)
		assert.Equal(t, genesis.ID(), storedID)
		assert.Equal(t, genesis.Header, storedHeader)

		for _, identity := range identities {
			var delta int64
			err = db.View(operation.RetrieveDelta(genesis.Header.View, identity.Role, identity.NodeID, &delta))
			require.Nil(t, err)

			assert.Equal(t, int64(identity.Stake), delta)
		}
	})
}

func TestExtendSealedBoundary(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		state := State{db: db}

		seal, err := state.Final().Seal()
		assert.NoError(t, err)
		assert.Equal(t, genesis.Seals[0], &seal)
		assert.Equal(t, flow.ZeroID, seal.BlockID) //genesis seal seals block with ID 0x00

		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Payload.Seals = nil
		block.Header.Height = 1
		block.Header.ParentID = genesis.ID()
		block.Header.PayloadHash = block.Payload.Hash()

		// seal
		newSeal := flow.Seal{
			BlockID:       block.ID(),
			PreviousState: genesis.Seals[0].FinalState,
			FinalState:    unittest.StateCommitmentFixture(),
			Signature:     nil,
		}

		sealingBlock := unittest.BlockFixture()
		sealingBlock.Payload.Identities = nil
		sealingBlock.Payload.Guarantees = nil
		sealingBlock.Payload.Seals = []*flow.Seal{&newSeal}
		sealingBlock.Header.Height = 2
		sealingBlock.Header.ParentID = block.ID()
		sealingBlock.Header.PayloadHash = sealingBlock.Payload.Hash()

		err = db.Update(func(txn *badger.Txn) error {
			err = procedure.InsertBlock(&block)(txn)
			if err != nil {
				return err
			}
			err = procedure.InsertBlock(&sealingBlock)(txn)
			if err != nil {
				return err
			}
			return nil
		})
		assert.NoError(t, err)

		err = mutator.Extend(block.ID())
		assert.NoError(t, err)

		err = mutator.Extend(sealingBlock.ID())
		assert.NoError(t, err)

		sealed, err := state.Final().Seal()
		assert.NoError(t, err)

		// Sealed only moves after a block is finalized
		// so here we still want to check for genesis seal
		assert.Equal(t, flow.ZeroID, sealed.BlockID)

		err = mutator.Finalize(sealingBlock.ID())
		assert.NoError(t, err)

		sealed, err = state.Final().Seal()
		assert.NoError(t, err)

		assert.Equal(t, block.ID(), sealed.BlockID)
	})
}

func TestBootstrapDuplicateID(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: duplicate node identifier "+
			"(0100000000000000000000000000000000000000000000000000000000000000)")
	})
}

func TestBootstrapZeroStake(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 0},
			{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: zero stake identity "+
			"(0100000000000000000000000000000000000000000000000000000000000000)")
	})
}

func TestBootstrapNoCollection(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: need at least one collection node")
	})
}

func TestBootstrapNoConsensus(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: need at least one consensus node")
	})
}

func TestBootstrapNoExecution(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: need at least one execution node")
	})
}

func TestBootstrapNoVerification(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x02}, Address: "a2", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: need at least one verification node")
	})
}

func TestBootstrapExistingAddress(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}

		identities := flow.IdentityList{
			{NodeID: flow.Identifier{0x01}, Address: "a1", Role: flow.RoleCollection, Stake: 1},
			{NodeID: flow.Identifier{0x02}, Address: "a1", Role: flow.RoleConsensus, Stake: 2},
			{NodeID: flow.Identifier{0x03}, Address: "a3", Role: flow.RoleExecution, Stake: 3},
			{NodeID: flow.Identifier{0x04}, Address: "a4", Role: flow.RoleVerification, Stake: 4},
		}
		genesis := flow.Genesis(identities)

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: duplicate node address (6131)")
	})
}

func TestBootstrapNonZeroHeight(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.Height = 42

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis header not valid: invalid initial height (42 != 0)")
	})
}

func TestBootstrapNonZeroParent(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.ParentID = unittest.BlockFixture().ID()

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis header not valid: genesis parent must be zero hash")
	})
}

func TestBootstrapNonEmptyCollections(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.Guarantees = unittest.CollectionGuaranteesFixture(1)
		genesis.PayloadHash = genesis.Payload.Hash()

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: genesis block must have zero guarantees")
	})
}

func TestBootstrapNoSeal(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.Seals = nil
		genesis.PayloadHash = genesis.Payload.Hash()

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: genesis block must have one seal")
	})
}

func TestBootstrapInvalidSeal(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.Seals[0].BlockID = unittest.IdentifierFixture()
		genesis.PayloadHash = genesis.Payload.Hash()

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: initial seal needs zero block ID")
	})
}

func TestBootstrapInvalidSealPrevState(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		mutator := &Mutator{state: &State{db: db}}
		genesis := flow.Genesis(identities)
		genesis.Seals[0].PreviousState = unittest.SealFixture().PreviousState
		genesis.PayloadHash = genesis.Payload.Hash()

		err := mutator.Bootstrap(genesis)
		require.EqualError(t, err, "genesis identities not valid: initial seal needs nil parent commit")
	})
}

func TestExtendValid(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.NoError(t, err)

		seal, err := mutator.state.Final().Seal()
		assert.NoError(t, err)
		assert.Equal(t, genesis.Seals[0], &seal)
		assert.Equal(t, flow.ZeroID, seal.BlockID) //genesis seal seals block with ID 0x00
	})
}

func TestExtendMissingParent(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 2
		block.View = 2
		block.ParentID = unittest.BlockFixture().ID()
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, "could not retrieve parent seal: could not lookup seal ID by block: key not found")

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendViewTooSmall(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.NoError(t, err)

		// create another block with the same height and view, that is coming after
		block.ParentID = block.ID()
		block.Height = 2
		block.View = 1

		err = db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendHeightTooSmall(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.NoError(t, err)

		// create another block with the same height and view, that is coming after
		block.ParentID = block.ID()
		block.Height = 1
		block.View = 2

		err = db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, "extend header not valid: block needs height equal to ancestor height+1 (1 != 1+1)")

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendBlockNotConnected(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		// add 2 blocks, the second finalizing/sealing the state of the first
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.PayloadHash = block.Payload.Hash()

		sealingBlock := unittest.BlockFixture()
		sealingBlock.Identities = nil
		sealingBlock.Guarantees = nil
		sealingBlock.Height = 2
		sealingBlock.View = 2
		sealingBlock.ParentID = block.ID()
		sealingBlock.Seals = []*flow.Seal{&flow.Seal{
			BlockID:       block.ID(),
			PreviousState: genesis.Seals[0].FinalState,
			FinalState:    unittest.StateCommitmentFixture(),
			Signature:     nil,
		}}
		sealingBlock.PayloadHash = sealingBlock.Payload.Hash()

		err := db.Update(func(txn *badger.Txn) error {
			err := procedure.InsertPayload(&block.Payload)(txn)
			if err != nil {
				return err
			}
			err = procedure.InsertBlock(&block)(txn)
			if err != nil {
				return err
			}
			err = operation.InsertSeal(sealingBlock.Seals[0])(txn)
			if err != nil {
				return err
			}
			err = procedure.InsertPayload(&sealingBlock.Payload)(txn)
			if err != nil {
				return err
			}
			err = procedure.InsertBlock(&sealingBlock)(txn)
			if err != nil {
				return err
			}
			return nil
		})
		assert.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.NoError(t, err)

		err = mutator.Finalize(block.ID())
		require.NoError(t, err)

		// create a fork at view/height 1 and try to connect it to genesis
		block.Timestamp = block.Timestamp.Add(time.Second)
		block.ParentID = genesis.ID()

		err = db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, "extend header not valid: block doesn't connect to finalized state")

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendSealNotConnected(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		seal1 := &flow.Seal{
			BlockID:       genesis.ID(),
			PreviousState: genesis.Seals[0].FinalState,
			FinalState:    unittest.StateCommitmentFixture(),
			Signature:     nil,
		}
		seal2 := &flow.Seal{
			BlockID:       genesis.ID(),
			PreviousState: unittest.StateCommitmentFixture(), // not seal1.FinalState, this will cause the error
			FinalState:    unittest.StateCommitmentFixture(),
			Signature:     nil,
		}

		block := unittest.BlockFixture()
		block.Identities = nil
		block.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.Seals = []*flow.Seal{seal1, seal2}
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(func(txn *badger.Txn) error {
			if err := procedure.InsertPayload(&block.Payload)(txn); err != nil {
				return err
			}
			if err := procedure.InsertBlock(&block)(txn); err != nil {
				return err
			}

			return nil
		})
		assert.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, fmt.Sprintf("seals not connected to state chain (%x)", seal1.FinalState))

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendIdentities(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = flow.IdentityList{block.Identities[0]}
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()
		block.PayloadHash = block.Payload.Hash()

		err := db.Update(operation.InsertIdentity(block.Identities[0]))
		require.NoError(t, err)
		err = db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, "extend payload not valid: extend block has identities")

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}

func TestExtendWrongIdentity(t *testing.T) {
	testWithBootstraped(t, func(t *testing.T, mutator *Mutator, db *badger.DB) {
		block := unittest.BlockFixture()
		block.Payload.Identities = nil
		block.Payload.Guarantees = nil
		block.Height = 1
		block.View = 1
		block.ParentID = genesis.ID()

		err := db.Update(procedure.InsertBlock(&block))
		require.NoError(t, err)

		err = mutator.Extend(block.ID())
		require.EqualError(t, err, "block integrity check failed")

		// verify seal not indexed
		var seal flow.Identifier
		err = db.View(operation.LookupSealIDByBlock(block.ID(), &seal))
		assert.EqualError(t, err, "key not found")
	})
}
