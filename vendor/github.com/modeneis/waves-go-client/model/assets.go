package model

// Assets is asset response body
type Assets struct {
	Address  string         `json:"address"`
	Balances []Balances       `json:"balances"`
}

// Balances represent balances
type Balances struct {
	Address string         `json:"address,omitempty"`
	AssetID string    `json:"assetId"`
	Issued  bool    `json:"issued,omitempty"`
	Balance int    `json:"balance,omitempty"`
}