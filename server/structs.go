package server

type HandshakeServer struct {
	Json string
}

type HandshakeClient struct {
	Version uint64
	Address string
	Port    uint16
	State   uint64
}

type PingClient struct {
	Payload int64
}
