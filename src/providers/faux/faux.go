// Package faux is used exclusive for testing purposes. I would strongly suggest you move along
// as there's nothing to see here.
package faux

import (
	"github.com/modeneis/coind/src/server/model_server"
)

// Provider is used only for testing.
type Provider struct {
}

// Name is used only for testing.
func (p Provider) Name() string {
	return "faux"
}

func (p Provider) GetType() string {
	return "Faux"
}

func (p *Provider) CreateFakeBlock(deposit model_server.Deposit) (retBlocks interface{}, err error) {
	return retBlocks, err
}

func (p *Provider) GetBlock(hash string) (block interface{}, err error) {
	return nil, nil
}

func (p *Provider) GetBestBlock(seq int64) (block interface{}, err error) {
	return block, err
}

func (p *Provider) GetGetBlockHash(tx string) (block interface{}, err error) {
	return block, nil
}

func (p *Provider) GetBlockCount() (count int32) {
	return count
}
