package proto

import "fmt"

// XProtocal describes a protocal of x.
// It should have at least 5 bytes and show as below:
//     ---------------------------------
//     0.......8.......16......24.....31
//     Version Type    ID      LengthH
//     LengthL Data.....................
//     ---------------------------------
type XProtocal struct {
	Version byte
	Type    byte
	ID      byte
	Length  uint16
	Data    []byte
}

// NewXProtocal translates a packet to protocal
func NewXProtocal(data []byte) (*XProtocal, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("a x protocal should have at least 5 bytes")
	}
	return &XProtocal{
		Version: data[0],
		Type:    data[1],
		ID:      data[2],
		Length:  uint16(data[3]) << 8 & uint16(data[4]),
		Data:    data[5:],
	}, nil
}

// DataSaver describes an interface of protocal data saver.
type DataSaver interface {
	// Marshal object to data
	Marshal() ([]byte, error)
	// Unmarshal data to object
	Unmarshal(data []byte) error
}
