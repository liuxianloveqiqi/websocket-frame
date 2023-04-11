package websocket_frame

import "net/http"

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

// 检查服务器是否支持子协议
func CheckSubProtocolFunc(p string) bool {
	// 随便整几个服务器支持的子协议
	supportedProtocols := []string{"chat", "text", "json", "protobuf"}

	// 遍历服务器支持的子协议列表，查看是否存在客户端提供的子协议
	for _, supportedProtocol := range supportedProtocols {
		if p == supportedProtocol {
			return true
		}
	}
	return false
}
