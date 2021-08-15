package parse

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
)

func Pack(obj interface{}) ([]byte, error) {
	var err error
	dataBuff := bytes.NewBuffer([]byte{})

	val := reflect.ValueOf(obj)
	r := val.Elem().NumField()
	for i := 0; i < r; i++ {
		fieldVal := val.Elem().Field(i)

		switch val := fieldVal.Interface().(type) {
		case int64:
			err = binary.Write(dataBuff, binary.BigEndian, val)
		case uint16:
			err = binary.Write(dataBuff, binary.BigEndian, val)
		case uint64:
			err = WriteUvarint(dataBuff, val)
		case string:
			err = WriteString(dataBuff, val)
		}
		if err != nil {
			return nil, err
		}
	}

	return dataBuff.Bytes(), nil
}

func UnPack(reader *bufio.Reader, obj interface{}) error {
	var err error

	val := reflect.ValueOf(obj)
	r := val.Elem().NumField()
	for i := 0; i < r; i++ {
		fieldVal := val.Elem().Field(i)

		switch fieldVal.Interface().(type) {
		case uint16:
			var temp uint16
			if err := binary.Read(reader, binary.BigEndian, &temp); err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(temp))
		case int64:
			var temp int64
			if err := binary.Read(reader, binary.BigEndian, &temp); err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(temp))
		case uint64:
			temp, err := binary.ReadUvarint(reader)
			if err != nil {
				return err
			}
			fieldVal.SetUint(temp)
		case string:
			temp, err := ReadString(reader)
			if err != nil {
				return err
			}
			fieldVal.SetString(temp)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func MsgBody(reader *bufio.Reader) (*bufio.Reader, error) {
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, length)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(bytes.NewReader(buffer)), nil
}

func ReadPackId(reader *bufio.Reader) (uint64, error) {
	return binary.ReadUvarint(reader)
}

func ConstructPack(msg []byte, packId uint64) ([]byte, error) {
	response := bytes.NewBuffer([]byte{})
	err := WriteUvarint(response, packId)
	if err != nil {
		return nil, err
	}
	_, err = response.Write(msg)
	if err != nil {
		return nil, err
	}
	response = AddLength(response)
	return response.Bytes(), nil
}
