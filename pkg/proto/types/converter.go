package types

import "github.com/kdada/tinyvpn/pkg/proto"

// object types
const (
	TypeAuthentication = uint16(iota)
	TypeAuthorization  = uint16(iota)
	TypeConfigRequest  = uint16(iota)
	TypeConfig         = uint16(iota)
	TypePacket         = uint16(iota)
	TypeFail           = uint16(iota)
)

var (
	// DefaultConverter is the default converter to converts between data and object
	DefaultConverter proto.Converter

	// all types
	types = map[uint16]interface{}{
		TypeAuthentication: &Authentication{},
		TypeAuthorization:  &Authorization{},
		TypeConfigRequest:  &ConfigRequest{},
		TypeConfig:         &Config{},
		TypePacket:         Packet{},
		TypeFail:           &Fail{},
	}
)

func init() {
	//register types
	converter, _ := proto.NewXProtocolConverter()
	for k, v := range types {
		converter.BindRelation(k, v)
	}
	DefaultConverter = converter
}
