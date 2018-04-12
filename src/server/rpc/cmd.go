package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/modeneis/coind/src/server/model_server"
)

// parsedRPCCmd represents a JSON-RPC request object that has been parsed into
// a known concrete command along with any error that might have happened while
// parsing it.
type parsedRPCCmd struct {
	id     interface{}
	method string
	cmd    interface{}
	err    *btcjson.RPCError
}

// parseCmd parses a JSON-RPC request object into known concrete command.  The
// err field of the returned parsedRPCCmd struct will contain an RPC error that
// is suitable for use in replies if the command is invalid in some way such as
// an unregistered command or invalid parameters.
func parseCmd(request *btcjson.Request) *parsedRPCCmd {
	var parsedCmd parsedRPCCmd
	parsedCmd.id = request.ID
	parsedCmd.method = request.Method

	cmd, err := btcjson.UnmarshalCmd(request)

	// Handle new commands except btcd cmds
	if request.Method == "nextdeposit" {
		if len(request.Params) == 1 {
			var deposit []model_server.Deposit
			err := json.Unmarshal(request.Params[0], &deposit)
			if err != nil {
				parsedCmd.err = btcjson.ErrRPCMethodNotFound
				return &parsedCmd
			}
			fmt.Printf("%v\n", deposit)
			parsedCmd.cmd = deposit
		}
		return &parsedCmd
	}

	if err != nil {
		// When the error is because the method is not registered,
		// produce a method not found RPC error.
		if jerr, ok := err.(btcjson.Error); ok &&
			jerr.ErrorCode == btcjson.ErrUnregisteredMethod {

			parsedCmd.err = btcjson.ErrRPCMethodNotFound
			return &parsedCmd
		}

		// Otherwise, some type of invalid parameters is the
		// cause, so produce the equivalent RPC error.
		parsedCmd.err = btcjson.NewRPCError(
			btcjson.ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}

	parsedCmd.cmd = cmd
	return &parsedCmd
}
