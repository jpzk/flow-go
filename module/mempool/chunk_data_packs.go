// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package mempool

import (
	"github.com/dapperlabs/flow-go/model/flow"
)

// Collections represents a concurrency-safe memory pool for chunk data packs.
type ChunkDataPacks interface {

	// Has checks whether the ChunkDataPack with the given hash is currently in
	// the memory pool.
	Has(cdpID flow.Identifier) bool

	// Add will add the given ChunkDataPack to the memory pool; it will error if
	// the ChunkDataPack is already in the memory pool.
	Add(cdp *flow.ChunkDataPack) error

	// Rem will remove the given ChunkDataPack from the memory pool; it will
	// return true if the ChunkDataPack was known and removed.
	Rem(cdpID flow.Identifier) bool

	// Get will retrieve the given ChunkDataPack from the memory pool; it will
	// error if the ChunkDataPack is not in the memory pool.
	ByID(cdpID flow.Identifier) (*flow.ChunkDataPack, error)

	// Size will return the current size of the memory pool.
	Size() uint

	// All will retrieve all ChunkDataPacks that are currently in the memory pool
	// as a slice.
	All() []*flow.ChunkDataPack

	// Hash will return a hash of the contents of the memory pool.
	Hash() flow.Identifier
}
