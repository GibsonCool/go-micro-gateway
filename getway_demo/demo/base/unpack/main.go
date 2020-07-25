package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func main() {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if err := Encode(bytesBuffer, "我要干翻苍穹!!"); err != nil {
		panic(err)
	}

	for {
		if bt, err := Decode(bytesBuffer); err == nil {
			fmt.Println("解码：", string(bt))
			continue
		}
		break
	}
}

const MsgHeader = "12345678"

func Encode(buffer io.Writer, content string) error {
	if err := binary.Write(buffer, binary.BigEndian, []byte(MsgHeader)); err != nil {
		return err
	}
	fmt.Println("头信息：", buffer)

	conLen := int32(len([]byte(content)))
	fmt.Println("conLen：", conLen)
	if err := binary.Write(buffer, binary.BigEndian, conLen); err != nil {
		return err
	}
	fmt.Println("内容长度：", buffer)

	if err := binary.Write(buffer, binary.BigEndian, []byte(content)); err != nil {
		return err
	}
	fmt.Println("内容：", buffer)
	return nil
}

func Decode(buffer io.Reader) (bodyBuf []byte, err error) {
	headByte := make([]byte, len(MsgHeader))
	if _, err := io.ReadFull(buffer, headByte); err != nil {
		return nil, err
	}
	fmt.Println("头信息：", string(headByte))
	if string(headByte) != MsgHeader {
		return nil, errors.New("头信息错误 ：" + string(headByte))
	}

	conLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(buffer, conLenBuf); err != nil {
		return nil, err
	}
	fmt.Println("内容长度：", conLenBuf)
	fmt.Println("内容长度：", string(conLenBuf))

	length := binary.BigEndian.Uint32(conLenBuf)
	bodyBuf = make([]byte, length)
	if _, err := io.ReadFull(buffer, bodyBuf); err != nil {
		return nil, err
	}
	fmt.Println("内容：", string(bodyBuf))

	return bodyBuf, nil
}
