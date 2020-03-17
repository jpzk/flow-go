package assignment

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/dapperlabs/flow-go/crypto/random"
	"github.com/dapperlabs/flow-go/model/chunkassignment"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/network/gossip/libp2p/test"
)

// PublicAssignmentTestSuite contains tests against methods of the PublicAssignment scheme
type PublicAssignmentTestSuite struct {
	suite.Suite
}

// TestAssignment invokes all the tests in this test suite
func TestAssignment(t *testing.T) {
	suite.Run(t, new(PublicAssignmentTestSuite))
}

// TestByNodeID evaluates the correctness of ByNodeID method of PublicAssignment
func (a *PublicAssignmentTestSuite) TestByNodeID() {
	size := 5
	// creates ids and twice chunks of the ids
	ids := test.CreateIDs(size)
	chunks := a.CreateChunks(2 * size)
	assignment := chunkassignment.NewAssignment()

	// assigns two chunks to each verifier node
	// j keeps track of chunks
	j := 0
	for i := 0; i < size; i++ {

		c := chunks.ByIndex(uint64(j))
		assignment.Add(c, append(assignment.Verifiers(c), ids[i].NodeID))
		j++
		c = chunks.ByIndex(uint64(j))
		assignment.Add(c, append(assignment.Verifiers(c), ids[i].NodeID))
	}

	// evaluating the chunk assignment
	// each verifier should have two certain chunks based on the assignment
	// j keeps track of chunks
	j = 0
	for i := 0; i < size; i++ {
		assignedChunks := assignment.ByNodeID(ids[i].NodeID)
		require.Len(a.T(), assignedChunks, 2)
		require.Contains(a.T(), assignedChunks, chunks.ByIndex(uint64(j)).Index)
		j++
		require.Contains(a.T(), assignedChunks, chunks.ByIndex(uint64(j)).Index)
	}

}

// TestAssignDuplicate tests assign Add duplicate verifiers
func (a *PublicAssignmentTestSuite) TestAssignDuplicate() {
	size := 5
	// creates ids and twice chunks of the ids
	var ids flow.IdentityList = test.CreateIDs(size)
	chunks := a.CreateChunks(2)
	assignment := chunkassignment.NewAssignment()

	// assigns first chunk to non-duplicate list of verifiers
	c := chunks.ByIndex(uint64(0))
	assignment.Add(c, ids.NodeIDs())
	require.Len(a.T(), assignment.Verifiers(c), size)

	// duplicates first verifier, hence size increases by 1
	ids = append(ids, ids[0])
	require.Len(a.T(), ids, size+1)
	// assigns second chunk to a duplicate list of verifiers
	c = chunks.ByIndex(uint64(1))
	assignment.Add(c, ids.NodeIDs())
	// should be size not size + 1
	require.Len(a.T(), assignment.Verifiers(c), size)
}

// TestPermuteEntirely tests permuting an entire IdentityList against
// randomness and deterministicity
func (a *PublicAssignmentTestSuite) TestPermuteEntirely() {
	// creates random ids
	count := 10
	seed := []byte{62, 53, 41, 97, 80, 21, 64, 58, 62, 53, 41, 97, 80, 21, 64, 58}
	var idList flow.IdentityList = test.CreateIDs(count)
	var ids flow.IdentifierList = idList.NodeIDs()
	original := make(flow.IdentifierList, count)
	copy(original, ids)

	// Randomness:
	rng1, err := random.NewRand(seed)
	require.NoError(a.T(), err)
	err = rng1.Shuffle(len(ids), ids.Swap)
	require.NoError(a.T(), err)

	// permutation should not change length of the list
	require.Len(a.T(), ids, count)

	// list should be permuted
	require.NotEqual(a.T(), ids, original)

	// Deterministiciy:
	// shuffling same list with the same seed should generate the same permutation
	rng2, err := random.NewRand(seed)
	require.NoError(a.T(), err)
	// permutes original list with the same seed
	err = rng2.Shuffle(len(original), original.Swap)
	require.NoError(a.T(), err)
	require.Equal(a.T(), ids, original)
}

// TestPermuteSublist tests permuting an a sublist of an
// IdentityList against randomness and deterministicity
func (a *PublicAssignmentTestSuite) TestPermuteSublist() {
	// creates random ids
	count := 10
	subset := 4

	seed := []byte{62, 53, 41, 97, 80, 21, 64, 58, 62, 53, 41, 97, 80, 21, 64, 58}
	var idList flow.IdentityList = test.CreateIDs(count)
	var ids flow.IdentifierList = idList.NodeIDs()
	original := make([]flow.Identifier, count)
	copy(original, ids)

	// Randomness:
	rng1, err := random.NewRand(seed)
	require.NoError(a.T(), err)
	err = rng1.Samples(len(ids), subset, ids.Swap)
	require.NoError(a.T(), err)

	// permutation should not change length of the list
	require.Len(a.T(), ids, count)

	// the initial subset of the list that is permuted should
	// be different than the original
	require.NotEqual(a.T(), ids[:subset], original[:subset])
}

// TestDeterministicy evaluates deterministic behavior of chunk assignment when
// chunks, random generator, and nodes are the same
func (a *PublicAssignmentTestSuite) TestDeterministicy() {
	c := 10    // keeps number of chunks
	n := 10    // keeps number of verifier nodes
	alpha := 1 // each chunk requires alpha verifiers
	seed := []byte{62, 53, 41, 97, 80, 21, 64, 58, 62, 53, 41, 97, 80, 21, 64, 58}
	chunks := a.CreateChunks(c)

	// making two random generator with the same seed
	// random generator #1
	rng1, err := random.NewRand(seed)
	require.NoError(a.T(), err)

	// random generator #2
	rng2, err := random.NewRand(seed)
	require.NoError(a.T(), err)

	// creates two set of the same nodes
	nodes1 := test.CreateIDs(n)
	nodes2 := make([]*flow.Identity, n)
	require.Equal(a.T(), copy(nodes2, nodes1), n)

	// chunk assignment of the first set
	p1, err := NewPublicAssignment(alpha).Assign(nodes1, chunks, rng1)
	require.NoError(a.T(), err)

	// chunk assignment of the second set
	p2, err := NewPublicAssignment(alpha).Assign(nodes2, chunks, rng2)
	require.NoError(a.T(), err)

	// list of nodes should get shuffled after public assignment
	// but it should contain same elements
	require.Equal(a.T(), p1, p2)
}

// TestChunkAssignmentOneToOne evaluates chunk assignment against
// several single chunk to single node assignment
func (a *PublicAssignmentTestSuite) TestChunkAssignmentOneToOne() {
	// assigning 10 chunks to one node
	a.ChunkAssignmentScenario(10, 1, 1)
	// assigning 10 chunks to 2 nodes
	// each chunk to one verifier
	a.ChunkAssignmentScenario(10, 2, 1)
	// each chunk to 2 verifiers
	a.ChunkAssignmentScenario(10, 2, 2)

	// assigning 10 chunks to 10 nodes
	// each chunk to one verifier
	a.ChunkAssignmentScenario(10, 10, 1)
	// each chunk to 6 verifiers
	a.ChunkAssignmentScenario(10, 10, 6)
	// each chunk to 9 verifiers
	a.ChunkAssignmentScenario(10, 10, 9)
}

// TestChunkAssignmentOneToMay evaluates chunk assignment
func (a *PublicAssignmentTestSuite) TestChunkAssignmentOneToMany() {
	//  against assigning 52 chunks to 7 nodes
	//  each chunk to 5 verifiers
	a.ChunkAssignmentScenario(52, 7, 5)
	//  against assigning 49 chunks to 9 nodes
	//  each chunk to 8 verifiers
	a.ChunkAssignmentScenario(52, 9, 8)
}

// ChunkAssignmentScenario is a test helper that creates chunkNum chunks, verNum verifiers
// and then assign each chunk to alpha randomly chosen verifiers
// it also evaluates that each chuck is assigned to alpha many unique verifier nodes
func (a *PublicAssignmentTestSuite) ChunkAssignmentScenario(chunkNum, verNum, alpha int) {
	rng, err := random.NewRand([]byte{62, 53, 41, 97, 80, 21, 64, 58, 62, 53, 41, 97, 80, 21, 64, 58})
	require.NoError(a.T(), err)
	chunks := a.CreateChunks(chunkNum)

	// creates nodes and keeps a copy of them
	nodes := test.CreateIDs(verNum)
	original := make([]*flow.Identity, verNum)
	require.Equal(a.T(), copy(original, nodes), verNum)

	p, err := NewPublicAssignment(alpha).Assign(nodes, chunks, rng)
	require.NoError(a.T(), err)

	// list of nodes should get shuffled after public assignment
	require.ElementsMatch(a.T(), nodes, original)

	for _, chunk := range chunks {
		// each chunk should be assigned to alpha verifiers
		require.Equal(a.T(), p.Verifiers(chunk).Len(), alpha)
	}
}

// CreateChunk creates and returns num chunks. It only fills the Index part of
// chunks to make them distinct from each other.
func (a *PublicAssignmentTestSuite) CreateChunks(num int) flow.ChunkList {
	list := flow.ChunkList{}
	for i := 0; i < num; i++ {
		c := &flow.Chunk{
			Index: uint64(i),
		}
		list.Insert(c)
	}
	require.Equal(a.T(), num, list.Size())
	return list
}
