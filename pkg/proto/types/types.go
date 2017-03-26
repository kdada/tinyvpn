package types

type Authentication struct {
	Version  string
	Account  string
	Password string
}

type Authorization struct {
	Version  string
	ServerIP uint32
	ClientIP uint32
}

type ConfigRequest struct {
}

type Config struct {
	Routes [][5]byte
}

type Packet []byte

type Fail struct {
	Code    uint16
	Message string
}
