package proto

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
)

const (
	// MaxProtocolLength defines max length of protocol
	MaxProtocolLength = 60000
	ProtocolHead      = byte(0x55)
	ProtocolTail      = byte(0xAA)
)

// XProtocol describes a protocol of x.
// It should have at least 4 bytes and show as below:
//     ---------------------------------
//     0.......8.......16......24.....31
//     0x55    TypeH   TypeL   LengthH
//     LengthL Data...
//     ...
//     0xAA
//     ---------------------------------
// A protocol package should start with 0x55 and
// end with 0xAA.
type XProtocol struct {
	Type   uint16
	Length uint16
	Data   []byte
}

// NewXProtocol creates a protocol
func NewXProtocol() *XProtocol {
	return &XProtocol{}
}

// ReadFrom reads data from a reader
func (xp *XProtocol) ReadFrom(r io.Reader) (int, error) {
	buf := [4]byte{}
	total, err := r.Read(buf[0:1])
	if err != nil {
		return total, err
	}
	if total != 1 {
		return total, fmt.Errorf("bad reader can't read head: %d", total)
	}
	if buf[0] != ProtocolHead {
		return total, fmt.Errorf("bad packet head: %d", buf[0])
	}
	count, err := r.Read(buf[:])
	total += count
	if err != nil {
		return total, err
	}
	if count != len(buf) {
		return total, fmt.Errorf("bad reader: %d", total)
	}
	xp.Type, xp.Length = uint16(buf[0])<<8+uint16(buf[1]), uint16(buf[2])<<8+uint16(buf[3])
	if xp.Length > MaxProtocolLength {
		return total, fmt.Errorf("bad reader with protocol data length exceed: %d", xp.Length)
	}
	xp.Data = make([]byte, xp.Length)
	count, err = r.Read(xp.Data)
	total += count
	if err != nil {
		return total, err
	}
	if count != int(xp.Length) {
		return total, fmt.Errorf("bad reader with wrong data length %d", count)
	}
	count, err = r.Read(buf[0:1])
	total += count
	if err != nil {
		return total, err
	}
	if count != 1 {
		return total, fmt.Errorf("bad reader can't read tail: %d", count)
	}
	if buf[0] != ProtocolTail {
		return total, fmt.Errorf("bad packet tail: %d", buf[0])
	}
	return total, nil
}

// WriteTo writes protocol to a writer
func (xp *XProtocol) WriteTo(r io.Writer) (int, error) {
	data := make([]byte, xp.Length+6)
	data[0] = ProtocolHead
	data[1] = byte(xp.Type >> 8)
	data[2] = byte(xp.Type)
	data[3] = byte(xp.Length >> 8)
	data[4] = byte(xp.Length)
	data[len(data)-1] = ProtocolTail
	copy(data[5:len(data)-1], xp.Data)
	count, err := r.Write(data)
	if err != nil {
		return count, err
	}
	if count < len(data) {
		return count, fmt.Errorf("bad writer only write: %d/%d", count, len(data))
	}
	return count, nil
}

// Converter converts data between XProtocol and object
type Converter interface {
	// ToObject converts XProtocol to object
	ToObject(protocol *XProtocol) (interface{}, error)
	// ToXProtocol converts object to XProtocol
	ToXProtocol(obj interface{}) (*XProtocol, error)
}

// XProtocolConverter converts between protocol and object
type XProtocolConverter struct {
	ProtocolToObject map[uint16]reflect.Type
	ObjectToProtocol map[reflect.Type]uint16
}

// NewXProtocolConverter create a converter
func NewXProtocolConverter() (*XProtocolConverter, error) {
	return &XProtocolConverter{
		ProtocolToObject: make(map[uint16]reflect.Type),
		ObjectToProtocol: make(map[reflect.Type]uint16),
	}, nil
}

// BingRelation adds a map between type and object. Obj can be any type.
// Don't bind a type twice. A pointer to a type is same as the type.
func (pc *XProtocolConverter) BindRelation(typ uint16, obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("obj should not be nil")
	}
	objType := reflect.TypeOf(obj)
	pc.ProtocolToObject[typ] = objType
	pc.ObjectToProtocol[objType] = typ
	gob.Register(obj)
	return nil
}

// ToObject converts protocol to an object
func (pc *XProtocolConverter) ToObject(protocol *XProtocol) (interface{}, error) {
	if protocol == nil {
		return nil, fmt.Errorf("can't convert nil protocol to object")
	}
	if int(protocol.Length) != len(protocol.Data) {
		return nil, fmt.Errorf("protocol length is not matched: %d/%d", protocol.Length, len(protocol.Data))
	}
	if int(protocol.Length) > MaxProtocolLength {
		return nil, fmt.Errorf("protocol data length exceed: %d", protocol.Length)
	}
	typ, ok := pc.ProtocolToObject[protocol.Type]
	if !ok {
		return nil, fmt.Errorf("protocol type %d does not exist", protocol.Type)
	}
	decoder := gob.NewDecoder(bytes.NewReader(protocol.Data))
	var value reflect.Value
	if typ.Kind() == reflect.Ptr {
		value = reflect.New(typ.Elem())
	} else {
		value = reflect.New(typ)
	}
	err := decoder.Decode(value.Interface())
	if err != nil {
		return nil, err
	}
	if typ.Kind() != reflect.Ptr {
		return value.Elem().Interface(), nil
	}
	return value.Interface(), nil
}

// ToXProtocol converts object to protocol
func (pc *XProtocolConverter) ToXProtocol(obj interface{}) (*XProtocol, error) {
	if obj == nil {
		return nil, fmt.Errorf("can't convert nil object to protocol")
	}
	objType := reflect.TypeOf(obj)
	typ, ok := pc.ObjectToProtocol[objType]
	if !ok {
		return nil, fmt.Errorf("there is no relation of type %s", objType.String())
	}
	buf := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(obj)
	if err != nil {
		return nil, err
	}
	if buf.Len() > MaxProtocolLength {
		return nil, fmt.Errorf("data length exceed: %d", buf.Len())
	}
	return &XProtocol{
		Length: uint16(buf.Len()),
		Type:   typ,
		Data:   buf.Bytes(),
	}, nil
}
