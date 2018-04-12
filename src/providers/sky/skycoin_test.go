package sky_test

import (
	"strconv"
	"testing"

	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/gui"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/stretchr/testify/require"

	"github.com/modeneis/coind/src/server/api"
	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/providers/sky"
)

func TestCreateFakeDepositSkycoin(t *testing.T) {

	var tt = []struct {
		name    string
		Deposit model_server.Deposit
	}{
		{
			"create new block with a deposit",
			model_server.Deposit{
				Address:  "cBnu9sUvv12dovBmjQKTtfE4rbjMmf3fzW",
				Value:    10000,
				N:        4,
				CoinType: api.CoinTypeSKY,
			},
		},
	}

	skyrpc := &webrpc.Client{
		Addr: "https://explorer.skycoin.net" + ":" + "443" + "/api/",
	}

	skyrest := &gui.Client{
		Addr: "https://explorer.skycoin.net" + ":" + "443" + "/api/",
	}

	sky := sky.Provider{
		SkyRPCClient:  skyrpc,
		SkyRESTClinet: skyrest,
	}

	sky.Start()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			b, err := sky.CreateFakeBlock(tc.Deposit)
			require.NoError(t, err)

			txFound := visor.ReadableTransactionOutput{}

			require.NotEmpty(t, b)
			require.NotEmpty(t, b.(*visor.ReadableBlocks).Blocks)

			for _, block := range b.(*visor.ReadableBlocks).Blocks {

				require.NotEmpty(t, block.Body.Transactions)

				for _, tx := range block.Body.Transactions {

					for _, out := range tx.Out {

						if out.Address == tc.Deposit.Address {
							txFound = out
						}
					}
				}

			}

			require.NotEmpty(t, txFound)
			val, err := strconv.Atoi(txFound.Coins)
			require.NoError(t, err)
			require.Equal(t, tc.Deposit.Value, int64(val))

		})

	}

}
