package pack

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var int2bytes_len int

func init() {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, int32(1))
	int2bytes_len = len(buf.Bytes())
}

func Pack(sessid int, msg []byte) ([]byte, error) {
	data := msg
	buf := bytes.NewBuffer([]byte{})
	if err := binary.Write(buf, binary.BigEndian, int32(int2bytes_len+len(data))); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, int32(sessid)); err != nil {
		return nil, err
	}
	if n, err := buf.Write(data); n != len(data) || err != nil {
		return nil, fmt.Errorf("pack partial")
	}
	return buf.Bytes(), nil
}

func Unpack(b []byte) (int, []byte, error) {
	buf := bytes.NewBuffer(b)

	var sz int32
	err := binary.Read(buf, binary.BigEndian, &sz)
	if err != nil {
		return 0, nil, err
	}
	var sessid int32
	err = binary.Read(buf, binary.BigEndian, &sessid)
	if err != nil {
		return 0, nil, err
	}
	return int(sessid), b[int2bytes_len:], nil
}

func ReadSize(buf *bytes.Buffer) (int, error) {
	if buf.Len() < int2bytes_len+int2bytes_len {
		return 0, nil
	}
	var sz int32
	err := binary.Read(buf, binary.BigEndian, &sz)
	if err != nil {
		return 0, err
	}
	return int(sz), nil
}

func ReadData(buf *bytes.Buffer, sz int) ([]byte, error) {
	if buf.Len() < sz {
		return nil, nil
	}
	bb := make([]byte, sz)
	buf.Read(bb)
	return bb, nil
}
