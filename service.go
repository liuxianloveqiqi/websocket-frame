package main

import (
	"crypto/rand"
	"io"

	"net/http"
)

// 处理 pong 消息的逻辑
func PongHandler(appData string) error {

	return nil
}

// 处理 ping 消息的逻辑
func PingHandler(appData string) error {
	return nil
}

func CheckOriginFunc(r *http.Request) bool {
	// 这里弄简单点就允许所有来源了
	return true
}

var supportedProtocols []string

// 检查服务器是否支持子协议
func CheckSubProtocolFunc(p string) bool {
	// 随便整几个服务器支持的子协议
	supportedProtocols = []string{"chat", "text", "json", "protobuf"}

	// 遍历服务器支持的子协议列表，查看是否存在客户端提供的子协议
	for _, supportedProtocol := range supportedProtocols {
		if p == supportedProtocol {
			return true
		}
	}
	return false
}

// 协议拓展
func ExtensionFunc(extension string) bool {
	// 如果客户端提供的协议拓展是 "permessage-deflate"，则表示服务器支持该协议拓展
	if extension == "permessage-deflate" {
		return true
	}

	// 如果客户端提供的协议拓展不是服务器所支持的任何一个协议拓展，则返回 false
	return false
}

// 写入数据帧
func WriteDataFrame(w io.Writer, payload []byte, payloadLen int, useMask bool, opcode int) error {
	var header [10]byte
	var mask [4]byte

	// 设置opcode
	header[0] = byte(opcode)

	// 判断是否设置掩码
	if useMask {
		maskingKey := make([]byte, 4)
		if _, err := rand.Read(maskingKey); err != nil {
			return err
		}
		// 随机生成一个掩码键 maskingKey，将它拷贝到 mask 数组中
		copy(mask[:], maskingKey[:])
		// 对payload中的每个字节与掩码键进行异或操作，实现掩码操作
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskingKey[i%4]
		}
	}
	_, err := w.Write(header[:])
	if err != nil {
		return err
	}
	_, err = w.Write(payload[:payloadLen])
	if err != nil {
		return err
	}

	return nil
}

// 关闭
func WriteCloseDataFrame(w io.Writer, useMask bool) error {
	//关闭帧的payload是一个空的byte切片，长度为0
	payload := []byte{}
	return WriteDataFrame(w, payload, len(payload), useMask, int(CloseMessage))
}
