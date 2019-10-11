package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/evan-forbes/gather/centralized/coinbase"
	"github.com/evan-forbes/gather/db/ethnode"
	"github.com/evan-forbes/gather/decentralized/oasis"
	"github.com/evan-forbes/gather/manage"
	"github.com/evan-forbes/sink"
)

func main() {
	///////////// ------- Management & Monitoring ------- /////////////
	var (
		GlobalCtx, GlobalCancel = context.WithCancel(context.Background())
		GlobalWG                sync.WaitGroup
	)
	// Start the GlobalShutdown (ctrl+C) listening thread
	manage.WaitForInterupt(GlobalCtx, GlobalCancel, &GlobalWG)

	// Local JSON based Logging
	manage.StartLogging(GlobalCtx, &GlobalWG)
	manage.Logger <- sink.Wrap("Boot", time.Now().Unix())

	///////////// Establish Connections /////////////
	////// Decentralized
	// - Ethereum Node
	// using infura for testing and demonstration purposes
	// *highly* reccomend using your own ETH node node
	ethClient, err = ethnode.Connect("wss://mainnet.infura.io/ws")
	if err != nil {
		fmt.Println("--Could not connect to eth node--")
	}
	// - Eth2Dai contract
	err = oasis.Run(ethClient)
	if err != nil {
		fmt.Println("could not connect with oasis", err)
	}
	////// Centralized
	// - Coinbase
	cb := manage.NewController(GlobalCtx, coinbase.Boot)
	err := cb.Start()
	if err != nil {
		fmt.Println("-------  failed to start Coinbase  --------\n", err)
	}
	// - Kraken
	GlobalWG.Wait()
	fmt.Println("Program completed")
}

////////////////////////////
// Program Pipeline Setup:
///////////////////////////
// establish connections

// initiate collection.
// -input clients
// - output a channel of data
// - errc
// v
// sort/label/process stream
// - output a channel of proprietary data structures
// - errc
// v
// standardize
// - output a channel of standardized data structures
// - errc
// v
// influxize
// - output a channel of influx points
// - errc
// v
// Store
// send off to there own handler
// - errc

// Error Handler
// - input all the errc
// - output channels
