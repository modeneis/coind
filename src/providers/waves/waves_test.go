package waves_test

import (
	"testing"

	"github.com/modeneis/waves-go-client/model"
	"github.com/stretchr/testify/require"

	"github.com/modeneis/coind/src/server/api"
	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/providers/waves"
)

func TestCreateFakeDepositWaves(t *testing.T) {

	var tt = []struct {
		name    string
		Deposit model_server.Deposit
	}{
		{
			"create new block with a deposit",
			model_server.Deposit{
				Address:  "3PFnbq8kQjYyPwHMaSnbyQ78t15uU6nbkqi",
				Value:    560100000000,
				N:        4,
				CoinType: api.CoinTypeWAVES,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			waves := waves.Provider{}
			waves.Start()

			blocks, err := waves.CreateFakeBlock(tc.Deposit)
			require.NoError(t, err)

			txFound := model.Transactions{}
			for _, tx := range blocks.(*model.Blocks).Transactions {
				if tx.Recipient == tc.Deposit.Address {
					txFound = tx
				}
			}

			require.NotEmpty(t, txFound)
			require.Equal(t, txFound.Amount, tc.Deposit.Value)

		})

	}

}
