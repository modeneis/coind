package model_server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/modeneis/coind/src/providers/faux"
	"github.com/modeneis/coind/src/providers/sky"
	"github.com/modeneis/coind/src/providers/waves"
	"github.com/modeneis/coind/src/server/model_server"
)

func Test_UseProviders(t *testing.T) {
	a := assert.New(t)

	//create sky provider
	s := &sky.Provider{}
	model_server.UseProviders(s)

	//create waves provider
	w := &waves.Provider{}
	model_server.UseProviders(w)

	//create test  provider
	fb := &faux.Provider{}
	model_server.UseProviders(fb)

	a.Equal(len(model_server.GetProviders()), 3)
	a.Equal(model_server.GetProviders()[s.GetType()], s)
	a.Equal(model_server.GetProviders()[w.GetType()], w)
	a.Equal(model_server.GetProviders()[fb.GetType()], fb)
	model_server.ClearProviders()
}

func Test_GetProvider(t *testing.T) {
	a := assert.New(t)

	//create sky provider
	s := &sky.Provider{}
	model_server.UseProviders(s)

	//create waves provider
	w := &waves.Provider{}
	model_server.UseProviders(w)

	//create faux provider
	fa := &faux.Provider{}
	model_server.UseProviders(fa)

	skyprovider, err := model_server.GetProvider(s.GetType())
	a.NoError(err)
	a.Equal(skyprovider, s)

	wavesprovider, err := model_server.GetProvider(w.GetType())
	a.NoError(err)
	a.Equal(wavesprovider, w)

	p, err := model_server.GetProvider(fa.GetType())
	a.NoError(err)
	a.Equal(p, fa)

	p, err = model_server.GetProvider("unknown")
	a.Error(err)
	a.Equal(err.Error(), "no provider for unknown exists")
	model_server.ClearProviders()
}
