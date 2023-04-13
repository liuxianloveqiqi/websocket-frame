package main

import (
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
