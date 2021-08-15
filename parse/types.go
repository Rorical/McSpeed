package parse

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

func ReadString(reader *bufio.Reader) (string, error) {
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	err = binary.Read(reader, binary.BigEndian, &buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func WriteString(w io.Writer, val string) error {
	length := uint64(len(val))
	err := WriteUvarint(w, length)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, []byte(val))
	return err
}

func WriteUvarint(w io.Writer, val uint64) error {
	buffer := make([]byte, 5)
	length := binary.PutUvarint(buffer, val)
	_, err := w.Write(buffer[:length])
	return err
}

func AddLength(w *bytes.Buffer) *bytes.Buffer {
	allbyte := w.Bytes()
	res := bytes.NewBuffer([]byte{})
	WriteUvarint(res, uint64(len(allbyte)))
	res.Write(allbyte)
	return res
}
