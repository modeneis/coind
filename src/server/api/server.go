package api

import (
	"fmt"
	"net/http"

	"github.com/modeneis/coind/src/server/model_server"
)

type HttpAPIServer struct {
	address string
	listen  *http.Server
	quit    chan struct{}
}

func NewHTTPAPIServer(address string) *HttpAPIServer {
	return &HttpAPIServer{
		address: address,
		quit:    make(chan struct{}),
	}
}

func InitRouting() *http.ServeMux {
	mux := http.NewServeMux()

	/*** BEGIN ROUTES ***/
	mux.HandleFunc("/api/nextdeposit", HttpHandleNextDeposit)

	mux.HandleFunc("/api/get_blocks", HttpHandleGetBlocks)

	mux.HandleFunc("/api/get_blocks_by_seq", HttpHandleGetBlocksBySeq)
	mux.HandleFunc("/api/get_last_blocks", HttpHandleGetBlocksBySeq)

	mux.HandleFunc("/api/get_block_count", HttpHandleGetBlockCount)
	mux.HandleFunc("/api/get_transaction", HttpHandleGetBlockHash)

	/*** END ROUTES ***/

	return mux
}

func (server *HttpAPIServer) Start() error {

	mux := InitRouting()

	server.listen = &http.Server{
		Addr:         server.address,
		Handler:      mux,
		ReadTimeout:  model_server.ServerReadTimeout,
		WriteTimeout: model_server.ServerWriteTimeout,
		IdleTimeout:  model_server.ServerIdleTimeout,
	}

	if err := server.listen.ListenAndServe(); err != nil {
		select {
		case <-server.quit:
			return nil
		default:
			return err
		}
	}
	return nil
}

func (server *HttpAPIServer) Stop() {
	close(server.quit)
	if server.listen != nil {
		if err := server.listen.Close(); err != nil {
			fmt.Println("http api server shutdown failed")
		}
	}
}
