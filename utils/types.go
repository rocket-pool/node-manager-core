package utils

import (
	"encoding/json"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Unsigned integer type
type Uinteger uint64

func (i Uinteger) MarshalYAML() (interface{}, error) {
	return strconv.FormatUint(uint64(i), 10), nil
}

func (i *Uinteger) UnmarshalYAML(value *yaml.Node) error {
	intVal, err := strconv.ParseUint(value.Value, 10, 64)
	if err != nil {
		return err
	}
	*i = Uinteger(intVal)
	return nil
}

func (i Uinteger) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(i), 10))
}
func (i *Uinteger) UnmarshalJSON(data []byte) error {
	// Unmarshal string
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}

	// Parse integer value
	value, err := strconv.ParseUint(dataStr, 10, 64)
	if err != nil {
		return err
	}

	// Set value and return
	*i = Uinteger(value)
	return nil
}

// Byte array type
type ByteArray []byte

func (b ByteArray) MarshalYAML() (interface{}, error) {
	return EncodeHexWithPrefix(b), nil
}

func (b *ByteArray) UnmarshalYAML(value *yaml.Node) error {
	bytes, err := DecodeHex(value.Value)
	if err != nil {
		return err
	}
	*b = bytes
	return nil
}

func (b ByteArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(EncodeHexWithPrefix(b))
}

func (b *ByteArray) UnmarshalJSON(data []byte) error {
	// Unmarshal string
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}

	// Decode hex
	value, err := DecodeHex(dataStr)
	if err != nil {
		return err
	}

	// Set value and return
	*b = value
	return nil
}
