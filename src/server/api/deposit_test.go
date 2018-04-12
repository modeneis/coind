package api_test

import (
	"testing"

	"github.com/drewolson/testflight"


	"encoding/json"
	"net/http"
	"strconv"

	"github.com/modeneis/waves-go-client/model"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/stretchr/testify/require"

	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/server/utils"
	"github.com/modeneis/coind/src/server/api"
)

func TestNextDeposit(t *testing.T) {

	var tt = []struct {
		name       string
		method     string
		expectCode int
		endpoint   string
		Deposits   []model_server.Deposit
	}{
		{
			"create new valid deposit for skycoin",
			"POST",
			http.StatusOK,
			"/api/nextdeposit",
			[]model_server.Deposit{
				{
					Address:  "1FeDtFhARLxjKUPPkQqEBL78tisenc9znS",
					Value:    10000,
					Hours:    3455,
					CoinType: api.CoinTypeSKY,
				},
			},
		},
		{
			"create new valid deposit for waves",
			"POST",
			http.StatusOK,
			"/api/nextdeposit",
			[]model_server.Deposit{
				{
					Address:  "3PFnbq8kQjYyPwHMaSnbyQ78t15uU6nbkqi",
					Value:    560100000000,
					CoinType: api.CoinTypeWAVES,
				},
			},
		},
		{
			"return error when deposit for unsupported coin",
			"POST",
			http.StatusBadRequest,
			"/api/nextdeposit",
			[]model_server.Deposit{
				{
					Address:  "1FeDtFhARLxjKUPPkQqEBL78tisenc9znS",
					Value:    10000,
					N:        4,
					Hours:    3455,
					CoinType: api.CoinTypeETH,
				},
			},
		},
	}

	mux := api.InitRouting()

	testflight.WithServer(mux, func(r *testflight.Requester) {

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {

				raw, err := json.Marshal(&tc.Deposits)
				require.NoError(t, err)

				var response *testflight.Response

				response = r.Post(tc.endpoint, "application/json", string(raw))
				defer utils.CheckError(t, response.RawResponse.Body.Close)
				require.Equal(t, tc.expectCode, response.StatusCode)

				if response.StatusCode == http.StatusOK {

					if tc.Deposits[0].CoinType == api.CoinTypeSKY {
						var b visor.ReadableBlocks
						err = json.Unmarshal(response.RawBody, &b)

						require.True(t, len(b.Blocks) > 0)

						txFound := visor.ReadableTransactionOutput{}

						require.NotEmpty(t, b)
						require.NotEmpty(t, b.Blocks)

						for _, block := range b.Blocks {

							require.NotEmpty(t, block.Body.Transactions)

							for _, tx := range block.Body.Transactions {

								for _, out := range tx.Out {

									if out.Address == tc.Deposits[0].Address {
										txFound = out
									}
								}
							}

						}

						require.NotEmpty(t, txFound)
						val, err := strconv.Atoi(txFound.Coins)
						require.NoError(t, err)
						require.Equal(t, tc.Deposits[0].Value, int64(val))

					} else if tc.Deposits[0].CoinType == api.CoinTypeWAVES {
						var blocks *model.Blocks
						err = json.Unmarshal(response.RawBody, &blocks)

						require.True(t, len(blocks.Transactions) > 0)

						txFound := model.Transactions{}
						for _, tx := range blocks.Transactions {
							if tx.Recipient == tc.Deposits[0].Address {
								txFound = tx
							}
						}

						require.NotEmpty(t, txFound)
						require.Equal(t, txFound.Amount, tc.Deposits[0].Value)

					}

					require.NoError(t, err)

				}

			})

		}

	})

}
