package api

import (
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/modeneis/coind/src/providers/sky"
	"github.com/modeneis/coind/src/providers/waves"
	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/server/utils"
)

func init() {

	model_server.UseProviders(
		sky.New(),
		waves.New(),
	)

	//if err := json.Unmarshal([]byte(SkyBlockString), &skyCoinFake.InitialBlock); err != nil {
	//	panic(err)
	//}

}

// ProcessDeposits
func ProcessDeposits(deposits []model_server.Deposit, w http.ResponseWriter) (err error) {
	//// Add new blocks
	var newBlock interface{}

	for _, deposit := range deposits {

		coinType := deposit.CoinType
		provider, err := model_server.GetProvider(coinType)
		if err != nil {
			err = fmt.Errorf("CoinType (%s) not supported for deposit %v", coinType, deposit)
			return err
		}

		// create new block
		newBlock, err = provider.CreateFakeBlock(deposit)
		if err != nil {
			return err
		}

	}

	if err := utils.JSONResponse(w, newBlock); err != nil {
		err = fmt.Errorf("ProcessDeposits got Err when running JSONResponse %v", err)
		return err
	}

	return err
}

// GetBlock
func GetBlock(coinType, hash string, w http.ResponseWriter) (err error) {
	provider, err := model_server.GetProvider(coinType)
	if err != nil {
		err = fmt.Errorf("CoinType (%s) not supported", coinType)
		return err
	}
	block, err := provider.GetBlock(hash)
	if err != nil {
		return &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	if err = utils.JSONResponse(w, block); err != nil {
		err = fmt.Errorf("ProcessDeposits got Err when running JSONResponse %v", err)
		return err
	}

	return nil
}

//GetBestBlock a block by seq or latest block if seq is 0
func GetBestBlock(coinType string, seq int64, w http.ResponseWriter) (err error) {
	provider, err := model_server.GetProvider(coinType)
	if err != nil {
		err = fmt.Errorf("CoinType (%s) not supported", coinType)
		return err
	}

	result, err := provider.GetBestBlock(seq)
	if err = utils.JSONResponse(w, result); err != nil {
		err = fmt.Errorf("ProcessDeposits got Err when running JSONResponse %v", err)
		return err
	}
	return err
}

//GetGetBlockHash
func GetGetBlockHash(coinType string, tx string, w http.ResponseWriter) (err error) {
	provider, err := model_server.GetProvider(coinType)
	if err != nil {
		err = fmt.Errorf("CoinType (%s) not supported", coinType)
		return err
	}

	result, err := provider.GetGetBlockHash(tx)
	if err = utils.JSONResponse(w, result); err != nil {
		err = fmt.Errorf("ProcessDeposits got Err when running JSONResponse %v", err)
		return err
	}

	return nil
}

//GetBlockCount
func GetBlockCount(coinType string, w http.ResponseWriter) (err error) {
	provider, err := model_server.GetProvider(coinType)
	if err != nil {
		err = fmt.Errorf("CoinType (%s) not supported", coinType)
		return err
	}

	result := provider.GetBlockCount()

	if err = utils.JSONResponse(w, result); err != nil {
		err = fmt.Errorf("ProcessDeposits got Err when running JSONResponse %v", err)
		return err
	}

	return nil

}
