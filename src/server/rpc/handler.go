package rpc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
)

// RPCErrorCode represents an error code to be used as a part of an RPCError
// which is in turn used in a JSON-RPC Response object.
//
// A specific type is used to help ensure the wrong errors aren't used.
type RPCErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

func handleGetBestBlock(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	//	if hash, ok := model_server.DefaultBlockStore.BlockHashes[int64(model_server.DefaultBlockStore.BestBlockHeight)]; ok {
	//		result := &btcjson.GetBestBlockResult{
	//			Hash:   hash,
	//			Height: model_server.DefaultBlockStore.BestBlockHeight,
	//		}
	//		return result, nil
	//	}
	return nil, fmt.Errorf("Block not found")
}

//
func handleGetBlock(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (retBlock interface{}, err error) {
	//	c := cmd.(*btcjson.GetBlockCmd)
	//	if block, ok := model_server.DefaultBlockStore.HashBlocks[c.Hash]; ok {
	//		//block.NextHash = ""
	//		//if block.Height < int64(defaultBlockStore.BestBlockHeight) {
	//		//	if hash, ok := defaultBlockStore.BlockHashes[block.Height+1]; ok {
	//		//		block.NextHash = hash
	//		//	}
	//		//}
	//
	//		retBlock = block
	//
	return
}

//
//	return nil, &btcjson.RPCError{
//		Code:    btcjson.ErrRPCBlockNotFound,
//		Message: "Block not found",
//	}
//}
//
func handleGetBlockHash(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	//	c := cmd.(*btcjson.GetBlockHashCmd)
	//	if hash, ok := model_server.DefaultBlockStore.BlockHashes[c.Index]; ok {
	//		return hash, nil
	//	}
	//
	return nil, &btcjson.RPCError{
		Code:    btcjson.ErrRPCBlockNotFound,
		Message: "Block not found",
	}
}

//
func handleGetBlockCount(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	//	return int64(model_server.DefaultBlockStore.BestBlockHeight), nil
	return 0, nil
}

//
//
func handleNextDeposit(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	//	var deposits []model_server.Deposit
	//	if cmd != nil {
	//		deposits = cmd.([]model_server.Deposit)
	//		fmt.Printf("Got %v\n", deposits)
	//	}
	//
	//	newBlock, err := model_server.ProcessDeposits(deposits)
	//
	//	if err != nil {
	//		fmt.Printf("processDeposits %v\n", err)
	//		return nil,  fmt.Errorf("Block not found")
	//	}
	//
	return nil, nil
}
