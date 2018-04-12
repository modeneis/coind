package utils

import (
	"encoding/json"
	"net/http"

	"github.com/btcsuite/btcd/btcjson"
)

// CreateMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func CreateMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *btcjson.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*btcjson.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = btcjson.NewRPCError(btcjson.ErrRPCInternal.Code, replyErr.Error())
		}
	}
	return btcjson.MarshalResponse(id, result, jsonErr)
}

// JSONResponse marshal data into json and write response
func JSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(d)
	return err
}
