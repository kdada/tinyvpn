package proto

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"fmt"
)

// Authentication stores Authentication info of client
type Authentication struct {
	// Account is the user name
	Account string
	// Timestamp is the time stamp of client
	Timestamp uint32
	// Key is the aes key of client
	Key []byte
	// SecretKey is used for encrypting auth data
	SecretKey []byte
}

// Length returns the length of mardhalled data
func (a *Authentication) Length() int {
	nameLength := len(a.Account) + 1
	dataLength := nameLength + 20
	dataLength = dataLength/16*16 + 16
	return nameLength + dataLength
}

// Marshal object to data
func (a *Authentication) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, a.Length()))
	buf.WriteByte(byte(len(a.Account)))
	buf.WriteString(a.Account)
	tBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tBytes, a.Timestamp)
	buf.Write(tBytes)
	buf.Write(a.Key)
	cipher, err := aes.NewCipher(a.SecretKey)
	if err != nil {
		return nil, err
	}
	result := buf.Bytes()
	pos := 1 + len(a.Account)
	cipher.Encrypt(result[pos:], result[pos:])
	return result, nil
}

// Unmarshal marshal data to object
func (a *Authentication) Unmarshal(data []byte) error {
	if len(data) <= 1 {
		return fmt.Errorf("wrong account length: %d", len(data))
	}
	nameLength := data[0]
	a.Account = string(data[1 : nameLength+1])
	if a.SecretKey == nil {
		return nil
	}
	encryptedData := data[1+nameLength:]
	if len(encryptedData) < 32 {
		return fmt.Errorf("wrong authentication data length: %d", len(data))
	}
	cipher, err := aes.NewCipher(a.SecretKey)
	if err != nil {
		return err
	}
	cipher.Decrypt(encryptedData, encryptedData)
	nameLength = encryptedData[0]
	name := string(encryptedData[1 : nameLength+1])
	if name != a.Account {
		return fmt.Errorf("unmatched accout: %s, %s", a.Account, name)
	}
	a.Timestamp = binary.BigEndian.Uint32(encryptedData[nameLength+1 : nameLength+5])
	a.Key = encryptedData[nameLength+5:]
	if len(a.Key) != 16 {
		return fmt.Errorf("invalid key: %x", a.Key)
	}
	return nil
}
