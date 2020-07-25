package unpack

import (
	"encoding/binary"
	"errors"
	"io"
)

const MsgHeader = "12345678"

func Encode(buffer io.Writer, content string) error {
	if err := binary.Write(buffer, binary.BigEndian, []byte(MsgHeader)); err != nil {
		return err
	}

	conLen := int32(len([]byte(content)))
	if err := binary.Write(buffer, binary.BigEndian, conLen); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, []byte(content)); err != nil {
		return err
	}
	return nil
}

func Decode(buffer io.Reader) (bodyBuf []byte, err error) {
	headByte := make([]byte, len(MsgHeader))
	if _, err := io.ReadFull(buffer, headByte); err != nil {
		return nil, err
	}
	if string(headByte) != MsgHeader {
		return nil, errors.New("头信息错误 ：" + string(headByte))
	}

	conLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(buffer, conLenBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(conLenBuf)
	bodyBuf = make([]byte, length)
	if _, err := io.ReadFull(buffer, bodyBuf); err != nil {
		return nil, err
	}

	return bodyBuf, nil
}
