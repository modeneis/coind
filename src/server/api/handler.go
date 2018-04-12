package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/modeneis/coind/src/server/model_server"
)

// httpHandleNextDeposit accept deposits and create a new block, returns the block height.
// Method: POST
// URI: /api/nextdeposit
// The request body is an array of deposits, for example:
//  [{
//     "Address": "1FeDtFhARLxjKUPPkQqEBL78tisenc9znS",
//     "Value":   10000,
//     "N":       4
//  }]
func HttpHandleNextDeposit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	r.Close = true

	if r.Method != http.MethodPost {
		errCode := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Accepts POST requests only"), errCode)
	}

	// Read and respond to the request.
	decoder := json.NewDecoder(r.Body)
	var deposits []model_server.Deposit
	err := decoder.Decode(&deposits)
	defer func() {
		if err := r.Body.Close(); err != nil {
			fmt.Println("Failed to close response body:", err)
		}
	}()

	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v", errCode, err), errCode)
		return
	}

	err = ProcessDeposits(deposits, w)
	if err != nil {
		errCode := http.StatusBadRequest
		errMsg := fmt.Sprintf("%d error processing data: %v", errCode, err)
		log.Println(errMsg)
		http.Error(w, errMsg, errCode)
		return
	}
}

// GetBlocksBySeq get blocks by seq
func HttpHandleGetBlocksBySeq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	r.Close = true

	if r.Method != http.MethodGet {
		errCode := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Accepts GET requests only"), errCode)
	}

	coinType := r.FormValue("cointype")
	if coinType == "" {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data, CoinType is invalid %v", errCode, coinType), errCode)
		return
	}

	seqStr := r.FormValue("seq")
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		seq = 0
	}

	err = GetBestBlock(coinType, int64(seq), w)
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data: %v", errCode, err), errCode)
		return
	}
}

func HttpHandleGetBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	r.Close = true

	if r.Method != http.MethodGet {
		errCode := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Accepts GET requests only"), errCode)
	}

	coinType := r.FormValue("cointype")
	if coinType == "" {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data, CoinType is invalid %v", errCode, coinType), errCode)
		return
	}

	hash := r.FormValue("hash")
	if coinType == "" {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data, CoinType is invalid %v", errCode, coinType), errCode)
		return
	}

	err := GetBlock(coinType, hash, w)
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data: %v", errCode, err), errCode)
		return
	}
}

func HttpHandleGetBlockHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	r.Close = true

	if r.Method != http.MethodGet {
		errCode := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Accepts GET requests only"), errCode)
	}

	coinType := r.FormValue("cointype")
	if coinType == "" {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data, CoinType is invalid %v", errCode, coinType), errCode)
		return
	}

	tx := r.FormValue("tx")

	err := GetGetBlockHash(coinType, tx, w)
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data: %v", errCode, err), errCode)
		return
	}
}

func HttpHandleGetBlockCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	r.Close = true

	if r.Method != http.MethodGet {
		errCode := http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Accepts GET requests only"), errCode)
	}

	coinType := r.FormValue("cointype")
	if coinType == "" {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data, CoinType is invalid %v", errCode, coinType), errCode)
		return
	}

	err := GetBlockCount(coinType, w)
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error processing data: %v", errCode, err), errCode)
		return
	}
}
