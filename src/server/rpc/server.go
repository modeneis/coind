package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/modeneis/coind/src/server/model_server"
	"github.com/modeneis/coind/src/server/utils"
)

type RpcServer struct {
	Started                int32
	Shutdown               int32
	Listeners              []net.Listener
	Wg                     sync.WaitGroup
	RequestProcessShutdown chan struct{}
	Key                    string
	Cert                   string
	Address                string
	MaxConcurrentReqs      int
}

var rpcHandlers = map[string]commandHandler{
	"get_blocks":        handleGetBlock,
	"get_blocks_by_seq": handleGetBestBlock,
	"get_lastblocks":    handleGetBlockHash,
	"getblockcount":     handleGetBlockCount,
	"nextdeposit":       handleNextDeposit, // for triggering a fake deposit
}

type commandHandler func(*RpcServer, interface{}, <-chan struct{}) (interface{}, error)

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status code.  This function was lifted and
// adapted from the standard library HTTP server code since it's not exported.
func (s *RpcServer) httpStatusLine(req *http.Request) string {
	code := http.StatusOK
	proto11 := req.ProtoAtLeast(1, 1)

	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		return proto + " " + codeStr + " " + text + "\r\n"
	}

	text = "status code " + codeStr
	return proto + " " + codeStr + " " + text + "\r\n"
}

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, and a writer.
func (s *RpcServer) writeHTTPResponseHeaders(req *http.Request, headers http.Header, w io.Writer) error {
	_, err := io.WriteString(w, s.httpStatusLine(req))
	if err != nil {
		return err
	}

	err = headers.Write(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}

// jsonRPCRead handles reading and responding to RPC messages.
func (s *RpcServer) jsonRPCRead(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&s.Shutdown) != 0 {
		return
	}

	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	if err := r.Body.Close(); err != nil {
		fmt.Println("Failed to close response body:", err)
	}

	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v",
			errCode, err), errCode)
		return
	}

	// Unfortunately, the http server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		fmt.Print(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		fmt.Printf("Failed to hijack HTTP connection: %v", err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("conn.Close failed:", err)
		}
	}()
	defer func() {
		if err := buf.Flush(); err != nil {
			fmt.Println("buf.Flush failed:", err)
		}
	}()
	if err := conn.SetReadDeadline(model_server.TimeZeroVal); err != nil {
		fmt.Println("conn.SetReadDeadline failed:", err)
	}

	// Attempt to parse the raw body into a JSON-RPC request.
	var responseID interface{}
	var jsonErr error
	var result interface{}
	var request btcjson.Request
	if err := json.Unmarshal(body, &request); err != nil {
		jsonErr = &btcjson.RPCError{
			Code:    btcjson.ErrRPCParse.Code,
			Message: "Failed to parse request: " + err.Error(),
		}
	}
	if jsonErr == nil {
		// The JSON-RPC 1.0 spec defines that notifications must have their "id"
		// set to null and states that notifications do not have a response.
		//
		// A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and
		// without an "id" member. The specification states that notifications
		// must not be responded to. JSON-RPC 2.0 permits the null value as a
		// valid request id, therefore such requests are not notifications.
		//
		// Bitcoin Core serves requests with "id":null or even an absent "id",
		// and responds to such requests with "id":null in the response.
		//
		// Btcd does not respond to any request without and "id" or "id":null,
		// regardless the indicated JSON-RPC protocol version unless RPC quirks
		// are enabled. With RPC quirks enabled, such requests will be responded
		// to if the reqeust does not indicate JSON-RPC version.
		//
		// RPC quirks can be enabled by the user to avoid compatibility issues
		// with software relying on Core's behavior.
		if request.ID == nil && !(model_server.RpcQuirks && request.Jsonrpc == "") {
			return
		}

		// The parse was at least successful enough to have an ID so
		// set it for the response.
		responseID = request.ID

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()

		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			parsedCmd := parseCmd(&request)

			if parsedCmd.err != nil {
				jsonErr = parsedCmd.err
			} else {
				result, jsonErr = s.standardCmdResult(parsedCmd, closeChan)
			}
		}
	}

	// Marshal the response.
	msg, err := utils.CreateMarshalledReply(responseID, result, jsonErr)
	if err != nil {
		fmt.Printf("Failed to marshal reply: %v\n", err)
		return
	}

	// Write the response.
	err = s.writeHTTPResponseHeaders(r, w.Header(), buf)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		fmt.Printf("Failed to write marshalled reply: %v\n", err)
	}

	// Terminate with newline to maintain compatibility with Bitcoin Core.
	if err := buf.WriteByte('\n'); err != nil {
		fmt.Printf("Failed to append terminating newline to reply: %v\n", err)
	}
}

func (s *RpcServer) Start() {
	if atomic.AddInt32(&s.Started, 1) != 1 {
		return
	}

	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Handler: rpcServeMux,
	}

	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true

		// Read and respond to the request.
		s.jsonRPCRead(w, r)
	})

	listeners, err := setupRPCListeners(s.Address)
	if err != nil {
		fmt.Printf("Unexpected setupRPCListeners error: %v\n", err)
		return
	}

	s.Listeners = listeners

	for _, listener := range listeners {
		s.Wg.Add(1)
		go func(listener net.Listener) {
			defer s.Wg.Done()
			fmt.Printf("RPC server listening on %s\n", listener.Addr())
			if err := httpServer.Serve(listener); err != nil {
				fmt.Println("httpServer.Serve failed:", err)
				return
			}
			fmt.Printf("RPC listener done for %s\n", listener.Addr())
		}(listener)
	}
}

func (s *RpcServer) Stop() error {
	if atomic.AddInt32(&s.Shutdown, 1) != 1 {
		fmt.Printf("RPC server is already in the process of shutting down\n")
		return nil
	}

	for _, listener := range s.Listeners {
		err := listener.Close()
		if err != nil {
			fmt.Printf("Problem shutting down rpc: %v\n", err)
			return err
		}
	}
	s.Wg.Wait()
	fmt.Printf("RPC server shutdown complete\n")

	return nil
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (s *RpcServer) RequestedProcessShutdown() <-chan struct{} {
	return s.RequestProcessShutdown
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC
// command and runs the appropriate handler to reply to the command.  Any
// commands which are not recognized or not implemented will return an error
// suitable for use in replies.
func (s *RpcServer) standardCmdResult(cmd *parsedRPCCmd, closeChan <-chan struct{}) (interface{}, error) {
	handler, ok := rpcHandlers[cmd.method]
	if ok {
		return handler(s, cmd.cmd, closeChan)
	}
	return nil, btcjson.ErrRPCMethodNotFound
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, model_server.SimpleAddr{Net: "tcp4", Addr: addr})
			netAddrs = append(netAddrs, model_server.SimpleAddr{Net: "tcp6", Addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, model_server.SimpleAddr{Net: "tcp6", Addr: addr})
		} else {
			netAddrs = append(netAddrs, model_server.SimpleAddr{Net: "tcp4", Addr: addr})
		}
	}
	return netAddrs, nil
}

func setupRPCListeners(address string) ([]net.Listener, error) {

	// Change the standard net.Listen function to the tls one.

	netAddrs, err := parseListeners([]string{address})
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := net.Listen(addr.Network(), addr.String())
		if err != nil {
			fmt.Printf("Can't listen on %s: %v\n", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}
