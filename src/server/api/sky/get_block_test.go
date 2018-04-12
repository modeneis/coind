package sky

import (
	"testing"

	"github.com/drewolson/testflight"

	"github.com/modeneis/coind/src/server/api"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/skycoin/skycoin/src/visor"
	"github.com/stretchr/testify/require"

	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/server/utils"
)

// Deposit records information about a POST deposit
type DepositTest struct {
	name       string
	CoinType   string
	endpoint   string
	method     string
	expectCode int
	Deposits   []model_server.Deposit
}

func TestGetBlock(t *testing.T) {

	var tt = []struct {
		name string

		DepositTest []DepositTest
	}{
		{
			"Given I create a new valid deposit then I shall be able to query it ",
			[]DepositTest{
				{
					method:     "POST",
					expectCode: http.StatusOK,
					endpoint:   "/api/nextdeposit",
					Deposits: []model_server.Deposit{
						{
							Address:  "1FeDtFhARLxjKUPPkQqEBL78tisenc9znS",
							Value:    10000,
							Hours:    3455,
							CoinType: api.CoinTypeSKY,
						},
					},
				},
				{
					name:       "get_blocks",
					method:     "GET",
					expectCode: http.StatusOK,
					endpoint:   "/api/get_blocks?cointype=" + api.CoinTypeSKY + "&hash=",
				},
				{
					name:       "get_blocks_by_seq",
					method:     "GET",
					expectCode: http.StatusOK,
					endpoint:   "/api/get_blocks_by_seq?cointype=" + api.CoinTypeSKY + "&seq=",
				},
				{
					name:       "get_last_blocks",
					method:     "GET",
					expectCode: http.StatusOK,
					endpoint:   "/api/get_last_blocks?cointype=" + api.CoinTypeSKY,
				},
				{
					name:       "get_block_count",
					method:     "GET",
					expectCode: http.StatusOK,
					endpoint:   "/api/get_block_count?cointype=" + api.CoinTypeSKY,
				},
				{
					name:       "get_transaction",
					method:     "GET",
					expectCode: http.StatusOK,
					endpoint:   "/api/get_transaction?cointype=" + api.CoinTypeSKY + "&index=",
				},
			},
		},
	}

	mux := api.InitRouting()

	testflight.WithServer(mux, func(r *testflight.Requester) {

		for _, tc := range tt {

			var targetInterface *visor.ReadableBlocks
			hash := ""
			seqStr := ""
			for _, deposit := range tc.DepositTest {

				if targetInterface != nil {
					blk := targetInterface.Blocks[0]

					if deposit.name == "get_blocks" {
						hash = blk.Head.BlockHash
					} else if deposit.name == "get_blocks_by_seq" {
						hash = strconv.Itoa(int(blk.Head.BkSeq))
					} else if deposit.name == "get_last_blocks" || deposit.name == "get_block_count" {
						hash = ""
						seqStr = strconv.Itoa(int(blk.Head.BkSeq))
					} else if deposit.name == "get_transaction" {
						hash = strconv.Itoa(int(blk.Head.BkSeq))
					}
				}

				t.Run(tc.name+deposit.endpoint+hash, func(t *testing.T) {

					var response *testflight.Response

					if deposit.method == "POST" {
						raw, err := json.Marshal(&deposit.Deposits)
						require.NoError(t, err)

						response = r.Post(deposit.endpoint, "application/json", string(raw))
						require.Equal(t, deposit.expectCode, response.StatusCode)

						err = json.Unmarshal(response.RawBody, &targetInterface)
						require.NoError(t, err)

					} else if deposit.method == "GET" {

						if deposit.name == "get_last_blocks" {
							require.NotEmpty(t, seqStr)
						} else if deposit.name == "get_block_count" {
						} else {
							require.NotEmpty(t, hash)
						}

						response = r.Get(deposit.endpoint + hash)

						defer utils.CheckError(t, response.RawResponse.Body.Close)

						require.Equal(t, deposit.expectCode, response.StatusCode)

						if deposit.name == "get_block_count" {
							require.Equal(t, seqStr, response.Body)
						}

					}

				})

			}
		}

	})

}
