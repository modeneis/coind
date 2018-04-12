package model_server

import (
	"time"
)

const (
	RpcQuirks = true

	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// The timeout configuration is necessary for public servers, or else
	// connections will be used up
	ServerReadTimeout  = time.Second * 10
	ServerWriteTimeout = time.Second * 20
	ServerIdleTimeout  = time.Second * 120
)

// timeZeroVal is simply the zero value for a time.Time and is used to avoid
// creating multiple instances.
var TimeZeroVal time.Time
