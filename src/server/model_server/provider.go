package model_server

import (
	"fmt"
)

// Provider needs to be implemented for each 3rd party provider
type Provider interface {
	Name() string
	GetType() string
	CreateFakeBlock(deposit Deposit) (blocks interface{}, err error)
	GetBlock(hash string) (block interface{}, err error)
	GetBestBlock(seq int64) (block interface{}, err error)
	GetGetBlockHash(tx string) (block interface{}, err error)
	GetBlockCount() (count int32)
}

// Providers is list of known/available providers.
type Providers map[string]Provider

var providers = Providers{}

// UseProviders sets a list of available providers
func UseProviders(p ...Provider) {
	for _, provider := range p {
		providers[provider.GetType()] = provider
	}
}

// GetProviders returns a list of all the providers currently in use.
func GetProviders() Providers {
	return providers
}

// GetProvider returns a previously created provider.
func GetProvider(name string) (Provider, error) {
	provider := providers[name]
	if provider == nil {
		return nil, fmt.Errorf("no provider for %s exists", name)
	}
	return provider, nil
}

// ClearProviders will remove all providers currently in use.
func ClearProviders() {
	providers = Providers{}
}
