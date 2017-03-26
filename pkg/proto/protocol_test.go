package proto

import (
	"bytes"
	"reflect"
	"testing"
)

var (
	data = []byte{0x55, 0, 10, 0, 8, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 0xAA}
)

func TestProtocol(t *testing.T) {
	r := bytes.NewReader(data)
	p := NewXProtocol()
	count, err := p.ReadFrom(r)
	t.Log("protocol:", p)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatalf("not read a packet: %d,%d", count, len(data))
	}
	w := bytes.NewBuffer(nil)
	count, err = p.WriteTo(w)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatalf("not write a packet: %d,%d", count, len(data))
	}
	t.Log(data)
	t.Log(w.Bytes())
	if !reflect.DeepEqual(w.Bytes(), data) {
		t.Fatal("data not equal")
	}
}

type TestStruct struct {
	ID       uint32
	Account  string
	Password []byte
	IsAdmin  bool
}

type TestStruct2 TestStruct

var (
	testStruct = TestStruct{
		ID:       10000,
		Account:  "Admin",
		Password: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		IsAdmin:  true,
	}
	testStruct2 = TestStruct2(testStruct)
	testPointer = &testStruct2
	testData    = []byte{0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
)

func TestProtocolConverter(t *testing.T) {
	pc, err := NewXProtocolConverter()
	if err != nil {
		t.Fatal(err)
	}
	err = pc.BindRelation(1, testStruct)
	if err != nil {
		t.Fatal(err)
	}
	err = pc.BindRelation(2, testPointer)
	if err != nil {
		t.Fatal(err)
	}
	err = pc.BindRelation(3, testData)
	if err != nil {
		t.Fatal(err)
	}
	Convert(t, pc, testStruct)
	Convert(t, pc, testPointer)
	Convert(t, pc, testData)

}
func Convert(t *testing.T, pc *XProtocolConverter, obj interface{}) {
	t.Log("type:", reflect.TypeOf(obj).String())
	t.Log("original:", obj)
	result, err := pc.ToXProtocol(obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("protocol:", result)
	nObj, err := pc.ToObject(result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("new:", nObj)
	if !reflect.DeepEqual(obj, nObj) {
		t.Fatal("can't resolve object")
	}
}
