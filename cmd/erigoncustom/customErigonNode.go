// A copy of turbo/node/node.go, but simplified and added:
//  +  appendCustomSyncStage(ethereum)

package main

import (
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
)

type CustomErigonNode struct {
	stack   *node.Node
	backend *eth.Ethereum
}

func (eri *CustomErigonNode) Serve() error {
	defer eri.stack.Close()
	utils.StartNode(eri.stack)
	eri.stack.Wait()
	return nil
}

func NewCustomErigonNode(
	nodeConfig *nodecfg.Config,
	ethConfig *ethconfig.Config,
	logger log.Logger,
) (*CustomErigonNode, error) {
	node, err := node.New(nodeConfig)
	if err != nil {
		utils.Fatalf("Failed to create Erigon node: %v", err)
	}

	ethereum, err := eth.New(node, ethConfig, logger)
	if err != nil {
		return nil, err
	}

	insertCustomSyncStage(ethereum)

	err = ethereum.Init(node, ethConfig)
	if err != nil {
		return nil, err
	}
	return &CustomErigonNode{stack: node, backend: ethereum}, nil
}
