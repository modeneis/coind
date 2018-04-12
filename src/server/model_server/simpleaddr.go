package model_server

// onionAddr implements the net.Addr interface with two struct fields
type SimpleAddr struct {
	Net, Addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a SimpleAddr) String() string {
	return a.Addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a SimpleAddr) Network() string {
	return a.Net
}
