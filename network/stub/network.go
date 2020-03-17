package stub

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module"
	"github.com/dapperlabs/flow-go/network"
	"github.com/dapperlabs/flow-go/protocol"
)

// Network is a mocked network layer made for testing Engine's behavior.
// When an Engine is installed on a Network, the mocked network will deliver
// all Engine's events synchronously in memory to another Engine, so that tests can run
// fast and easy to assert errors.
type Network struct {
	sync.Mutex
	state        protocol.State
	me           module.Local
	hub          *Hub
	engines      map[uint8]network.Engine
	seenEventIDs map[string]bool
}

// NewNetwork create a mocked network.
// The committee has the identity of the node already, so only `committee` is needed
// in order for a mock hub to find each other.
func NewNetwork(state protocol.State, me module.Local, hub *Hub) *Network {
	o := &Network{
		state:        state,
		me:           me,
		hub:          hub,
		engines:      make(map[uint8]network.Engine),
		seenEventIDs: make(map[string]bool),
	}
	// Plug the network to a hub so that networks can find each other.
	hub.Plug(o)
	return o
}

// GetID returns the identity of the Node.
func (mn *Network) GetID() flow.Identifier {
	return mn.me.NodeID()
}

// Register implements pkg/module/Network's interface
func (mn *Network) Register(channelID uint8, engine network.Engine) (network.Conduit, error) {
	_, ok := mn.engines[channelID]
	if ok {
		return nil, errors.Errorf("engine code already taken (%d)", channelID)
	}
	conduit := &Conduit{
		channelID: channelID,
		submit:    mn.submit,
	}
	mn.engines[channelID] = engine
	return conduit, nil
}

// submit is called when an Engine is sending an event to an Engine on another node or nodes.
func (mn *Network) submit(channelID uint8, event interface{}, targetIDs ...flow.Identifier) error {
	m := &PendingMessage{
		From:      mn.GetID(),
		ChannelID: channelID,
		Event:     event,
		TargetIDs: targetIDs,
	}

	mn.buffer(m)

	return nil
}

// return a certain node has seen a certain key
func (mn *Network) haveSeen(key string) bool {
	seen, ok := mn.seenEventIDs[key]
	if !ok {
		return false
	}
	return seen
}

// mark a certain node has seen a certain event for a certain engine
func (mn *Network) seen(key string) {
	mn.seenEventIDs[key] = true
}

// buffer saves the request into pending buffer
func (mn *Network) buffer(m *PendingMessage) {
	mn.hub.Buffer.Save(m)
}

// DeliverAll sends all pending messages to the receivers. The receivers
// might be triggered to forward messages to its peers, so this function will
// block until all receivers have done their forwarding.
func (mn *Network) DeliverAll(recursive bool) {
	mn.hub.Buffer.DeliverRecursive(func(m *PendingMessage) {
		_ = mn.sendToAllTargets(m, recursive)
	})
}

// DeliverAllRecursiveExcept flushes all pending messages in the buffer except
// those that satisfy the shouldDrop predicate function. All messages that
// satisfy the shouldDrop predicate are permanently dropped. This function will
// block until all receivers have done their forwarding.
func (mn *Network) DeliverAllExcept(recursive bool, shouldDrop func(*PendingMessage) bool) {
	mn.hub.Buffer.DeliverRecursive(func(m *PendingMessage) {
		if shouldDrop(m) {
			return
		}
		_ = mn.sendToAllTargets(m, recursive)
	})
}

// DeliverSome delivers all messages in the buffer that satisfy the
// shouldDeliver predicate. Any messages that are not delivered remain in the
// buffer.
func (mn *Network) DeliverSome(recursive bool, shouldDeliver func(*PendingMessage) bool) {
	mn.hub.Buffer.Deliver(func(m *PendingMessage) bool {
		if shouldDeliver(m) {
			return mn.sendToAllTargets(m, recursive) != nil
		}
		return false
	})
}

// sendToAllTargets send a message to all its targeted nodes if the targeted
// node has not yet seen it.
func (mn *Network) sendToAllTargets(m *PendingMessage, recursive bool) error {
	mn.Lock()
	defer mn.Unlock()

	key, err := eventKey(m.ChannelID, m.Event)
	if err != nil {
		return err
	}
	for _, nodeID := range m.TargetIDs {
		// Find the network of the targeted node
		receiverNetwork, exist := mn.hub.GetNetwork(nodeID)
		if !exist {
			return errors.Errorf("Network can not find a node by ID %v", nodeID)
		}

		// Check if the given engine already received the event.
		// This prevents a node receiving the same event twice.
		if receiverNetwork.haveSeen(key) {
			continue
		}

		// mark the peer has seen the event
		receiverNetwork.seen(key)

		// Find the engine of the targeted network
		receiverEngine, ok := receiverNetwork.engines[m.ChannelID]
		if !ok {
			return errors.Errorf("Network can not find engine ID: %v for node: %v", m.ChannelID, nodeID)
		}

		if recursive {
			err := receiverEngine.Process(m.From, m.Event)
			if err != nil {
				return errors.Wrapf(err, "senderEngine failed to process event: %v", m.Event)
			}
		} else {
			// Call `Submit` to let receiver engine receive the event directly.
			// Submit is supposed to process event asynchronously, but if it doesn't we are risking
			// deadlock (if it trigger another message sending we might end up calling this very function again)
			// Running it in Go-routine is some cheap form of defense against deadlock in tests
			go receiverEngine.Submit(m.From, m.Event)
		}

	}
	return nil
}
