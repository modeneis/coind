package model_server

// Deposit records information about a POST deposit
type Deposit struct {
	Address  string // deposit address
	Value    int64  // deposit amount. For BTC, measured in satoshis.
	Hours    uint64 // hours amount.
	Height   int64  // the block height
	Tx       string // the transaction id
	N        uint32 // the index of vout in the tx [BTC]
	CoinType string
}
