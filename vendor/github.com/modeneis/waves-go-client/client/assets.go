package client

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"

	"github.com/modeneis/waves-go-client/model"
	"log"
)

// AssetsService holds sling instance
type AssetsService struct {
	sling *sling.Sling
}

// NewAssetsService returns a new AccountService.
func NewAssetsService(url string) *AssetsService {
	if url == "" {
		url = MainNET
	}
	return &AssetsService{
		sling: sling.New().Client(nil).Base(url),
	}
}

// GetAssetsBalanceAddress Balances for all assets that the given account ever had (besides WAVES).
// https://github.com/wavesplatform/Waves/wiki/Waves-Node-REST-API#get-assetsbalanceaddress
func (s *AssetsService) GetAssetsBalanceAddress(address string) (*model.Assets, *http.Response, error) {

	assets := new(model.Assets)
	apiError := new(model.APIError)
	path := fmt.Sprintf("/assets/balance/%s", address)
	res, err := s.sling.New().Get(path).Receive(assets, apiError)
	if err != nil {
		log.Println("ERROR: GetAssetsBalanceAddress, ", err)
	}

	return assets, res, model.FirstError(err, apiError)
}


// GetAssetsBalanceAddressAssetID Account's balance for the given asset.
// https://github.com/wavesplatform/Waves/wiki/Waves-Node-REST-API#get-assetsbalanceaddressassetid
func (s *AssetsService) GetAssetsBalanceAddressAssetID(address, assetID string) (*model.Balances, *http.Response, error) {

	balances := new(model.Balances)
	apiError := new(model.APIError)
	path := fmt.Sprintf("/assets/balance/%s/%s", address,assetID)
	res, err := s.sling.New().Get(path).Receive(balances, apiError)
	if err != nil {
		log.Println("ERROR: GetAssetsBalanceAddressAssetID, ", err)
	}

	return balances, res, model.FirstError(err, apiError)
}
