package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const (
	TextMessage   byte = 1
	BinaryMessage byte = 2
	CloseMessage  byte = 8
	PingMessage   byte = 9
	PongMessage   byte = 10
)

type Msg struct {
	Type       int
	Content    []byte // 消息内容
	PayloadLen int    // 有效载荷长
}
type Handler func(appData string) error

type MyConn struct {
	conn        net.Conn // 连接
	ReadLimit   int      // 读取限制，即消息的最大长度
	WriteLimit  int      // 发送限制，即一次发送的最大消息长度
	PongHandle  Handler  // pong消息的处理函数
	PingHandle  Handler  // ping消息的处理函数
	SubProtocol string   // 子协议

}

type Upgrader struct {
	ReadBufferSize  int                      // 读缓冲区的大小
	WriteBufferSize int                      // 写缓冲区的大小
	CheckOrigin     func(*http.Request) bool // 检查请求的来源是否合法
}

func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (conn MyConn, err error) {
	// 创建一个MyConn
	conn = MyConn{
		ReadLimit:  u.ReadBufferSize,
		WriteLimit: u.WriteBufferSize,
		PongHandle: PongHandler,
		PingHandle: PingHandler,
	}
	//检查请求头 Connection
	if r.Header.Get("Connection") != "Upgrade" {
		err = errors.New("connection没有upgrade")
		return
	}
	//检查请求头 Upgrade
	if r.Header.Get("Upgrade") != "websocket" {
		// 如果 Upgrade 不是 websocket，返回错误
		err = errors.New("不是websocket连接")
		return
	}

	//检查请求方式是不是GET
	if r.Method != http.MethodGet {
		err = errors.New("请求不为GET")
		return
	}
	//检查请求头 Sec-Websocket-Version 是否为13
	if r.Header.Get("Sec-Websocket-Version") != "13" {
		err = errors.New("Sec-Websocket-Version不是13")
		return
	}

	//检查Origin是否是允许的
	origin := r.Header.Get("Origin")
	u.CheckOrigin = CheckOriginFunc
	if !u.CheckOrigin(r) {
		err = fmt.Errorf("origin:%s 不被允许访问", origin)
		return
	}

	//检查请求头 Sec-Websocket-Key 是否存在且不为空
	if key := r.Header.Get("Sec-Websocket-Key"); key == "" {
		err = errors.New("Sec-Websocket-Key不存在或为空")
		return
	}

	//处理子协议字段
	subProtocol := r.Header.Get("Sec-Websocket-Protocol")
	if subProtocol != "" {
		protocols := strings.Split(subProtocol, ",")
		// 遍历客户端提供的子协议列表
		for _, p := range protocols {
			// 检查服务器是否支持子协议，如果支持则返回
			if CheckSubProtocolFunc(p) {
				responseHeader.Set("Sec-Websocket-Protocol", p)
				conn.SubProtocol = p
				break
			}
			// 如果所有子协议都不支持，就不需要做任何处理
		}
	}

	//处理协议拓展
	extension := r.Header.Get("Sec-WebSocket-Extensions")
	if extension != "" {
		extensions := strings.Split(extension, ",")
		// 遍历客户端提供的协议拓展列表
		for _, e := range extensions {
			// 检查服务器是否支持协议拓展，如果支持则返回接受该拓展的响应头
			if ExtensionFunc(e) {
				responseHeader.Set("Sec-WebSocket-Extensions", e)
				break
			}
		}
	}

	// 从http.ResponseWriter重新拿到conn
	// 调用 http.Hijacker 拿到这个连接现在开始就可以使用websocket通信了
	// Hijack的中文意思是劫持的意思。
	h, ok := w.(http.Hijacker)
	if !ok {
		err = errors.New("fail to hijacker the request")
		return
	}

	// 截获请求，建立websocket通信
	conn.conn, _, err = h.Hijack()
	if err != nil {
		return
	}

	// 创建一个GUID（全局唯一标识符）常量
	const GUID = "0ed958ca-8b16-4cbf-a475-641f30e7d6ff"

	// 构造回复报文的Sec-WebSocket-Accept字段的值
	secWebSocketKey := r.Header.Get("Sec-WebSocket-Key")
	accept := sha1.Sum([]byte(secWebSocketKey + GUID))
	acceptString := base64.StdEncoding.EncodeToString(accept[:])

	// 构造回复报文的Sec-WebSocket-Accept和Sec-WebSocket-Protocol字段
	resp := []byte("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptString + "\r\n")
	// 回复报文的头部信息拼接成一个字节数组 resp
	if conn.SubProtocol != "" {
		resp = append(resp, []byte("Sec-WebSocket-Protocol: "+conn.SubProtocol+"\r\n")...)

	}
	//最后别忘了换行
	resp = append(resp, "\r\n"...)

	//将请求报文写入
	_, err = conn.conn.Write(resp)
	return
}

func (c *MyConn) ReadMsg() (m Msg, err error) {
	// 根据数据帧读取数据

	// 读取第一个字节
	firstByte := make([]byte, 1)
	_, err = c.conn.Read(firstByte)
	if err != nil {
		return
	}
	// TODO 用位运算处理这些字节
	// 根据第一个字节的内容读取后续的数据
	// FIN是否提示为最终消息
	// RSV1~3的协议拓展判断
	fin := firstByte[0]>>7 == 1    // 判断 FIN 字段是否为 1
	rsv1 := firstByte[0]>>6&1 == 1 // 判断 RSV1 字段是否为 1
	rsv2 := firstByte[0]>>5&1 == 1 // 判断 RSV2 字段是否为 1
	rsv3 := firstByte[0]>>4&1 == 1 // 判断 RSV3 字段是否为 1
	opcode := firstByte[0] & 0x0F  // 获取 Opcode 字段
	fmt.Println(fin, rsv1, rsv2, rsv3)
	// 读取第二个字节
	secondByte := make([]byte, 1)
	_, err = c.conn.Read(secondByte)
	if err != nil {
		return
	}
	// 检查是否使用mask
	mask := secondByte[0] >> 7
	if mask != 1 {
		err = errors.New("没有mask")
		return
	}
	// 读取4字节的掩码值
	maskingKey := make([]byte, 4)
	_, err = c.conn.Read(maskingKey)
	if err != nil {
		return
	}
	payloadLen := int(secondByte[0] & 0x7F) // 取1-7位
	// 然后掩码处理
	maskedData := make([]byte, payloadLen)
	_, err = c.conn.Read(maskedData)
	if err != nil {
		return
	}
	for i := 0; i < payloadLen; i++ {
		maskedData[i] ^= maskingKey[i%4]
	}

	// 处理payload len
	switch {
	case 125 >= payloadLen && payloadLen > 0:
		// 如果有效载荷长度小于126，则直接使用该长度
		m.PayloadLen = payloadLen
	case payloadLen == 126:
		// 如果长度为126，则使用接下来的2个字节表示长度
		lenBytes := make([]byte, 2)
		_, err = c.conn.Read(lenBytes)
		if err != nil {
			return
		}
	case payloadLen == 127:
		// 如果长度为127，则使用接下来的8个字节表示长度
		lenBytes := make([]byte, 8)
		_, err = c.conn.Read(lenBytes)
		if err != nil {
			return
		}
	default:
		// 都不是？ 那发个锤
		return
	}

	// 如果你有ReadLimit这个功能 该咋搞呢
	// 检查是否超过读取限制
	if c.ReadLimit > 0 && payloadLen > c.ReadLimit {
		fmt.Println(errors.New("超过读取限制"))
		return
	}

	m.Type = int(opcode)

	switch opcode {
	case PingMessage:
		m.Type = 9
	case PongMessage:
		m.Type = 10
	case TextMessage:
		m.Type = 1
	case BinaryMessage:
		m.Type = 2
	case CloseMessage:
		m.Type = 8

	default:
		err = fmt.Errorf("未知的操作码%d", opcode)
		return
	}

	return m, nil
}

func (c *MyConn) WriteMsg(m Msg) (err error) {
	// 按照数据帧写出数据
	// 消息内容

	// payloadLen怎么处理？
	// 啥时候应该消息分片？

	// 写出发送的数据
	_, err = c.conn.Write(data)
	if err != nil {
		// 写不进去，好寄
	}
	return
}
func (c *MyConn) Close() {
	// 应该怎么关闭？
}
