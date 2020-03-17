package main

import (
	"github.com/spf13/pflag"

	"github.com/dapperlabs/flow-go/cmd"
	"github.com/dapperlabs/flow-go/engine/verification/ingest"
	"github.com/dapperlabs/flow-go/engine/verification/verifier"
	"github.com/dapperlabs/flow-go/module"
	"github.com/dapperlabs/flow-go/module/assignment"
	"github.com/dapperlabs/flow-go/module/mempool/stdmap"
)

func main() {

	var (
		receiptLimit    uint
		collectionLimit uint
		blockLimit      uint
		chunkLimit      uint
		err             error
		receipts        *stdmap.Receipts
		blocks          *stdmap.Blocks
		collections     *stdmap.Collections
		chunkStates     *stdmap.ChunkStates
		chunkDataPacks  *stdmap.ChunkDataPacks
		verifierEng     *verifier.Engine
	)

	cmd.FlowNode("verification").
		ExtraFlags(func(flags *pflag.FlagSet) {
			flags.UintVar(&receiptLimit, "receipt-limit", 100000, "maximum number of execution receipts in the memory pool")
			flags.UintVar(&collectionLimit, "collection-limit", 100000, "maximum number of collections in the memory pool")
			flags.UintVar(&blockLimit, "block-limit", 100000, "maximum number of result blocks in the memory pool")
			flags.UintVar(&chunkLimit, "chunk-limit", 100000, "maximum number of chunk states in the memory pool")
		}).
		Module("execution receipts mempool", func(node *cmd.FlowNodeBuilder) error {
			receipts, err = stdmap.NewReceipts(receiptLimit)
			return err
		}).
		Module("collections mempool", func(node *cmd.FlowNodeBuilder) error {
			collections, err = stdmap.NewCollections(collectionLimit)
			return err
		}).
		Module("blocks mempool", func(node *cmd.FlowNodeBuilder) error {
			blocks, err = stdmap.NewBlocks(blockLimit)
			return err
		}).
		Module("chunk states mempool", func(node *cmd.FlowNodeBuilder) error {
			chunkStates, err = stdmap.NewChunkStates(chunkLimit)
			return err
		}).
		Module("chunk data pack mempool", func(node *cmd.FlowNodeBuilder) error {
			chunkDataPacks, err = stdmap.NewChunkDataPacks(chunkLimit)
			return err
		}).
		Component("verifier engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			verifierEng, err = verifier.New(node.Logger, node.Network, node.State, node.Me)
			return verifierEng, err
		}).
		Component("ingest engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			alpha := 10
			assigner := assignment.NewPublicAssignment(alpha)
			// https://github.com/dapperlabs/flow-go/issues/2703
			// proper place and only referenced here
			// Todo the hardcoded default value should be parameterized as alpha in a
			// should be moved to a configuration class
			// DISCLAIMER: alpha down there is not a production-level value
			eng, err := ingest.New(node.Logger, node.Network, node.State, node.Me, verifierEng, receipts, blocks, collections, chunkStates, chunkDataPacks, assigner)
			return eng, err
		}).
		Run()
}
