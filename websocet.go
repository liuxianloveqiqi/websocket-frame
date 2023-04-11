package websocket_frame

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Msg struct {
	Typ     int    // 消息类型
	Content []byte // 消息内容
}
type Handler func(appData string) error

type MyConn struct {
	conn       net.Conn // 连接
	ReadLimit  int      // 读取限制，即消息的最大长度
	WriteLimit int      // 发送限制，即一次发送的最大消息长度
	PongHandle Handler  // pong消息的处理函数
	PingHandle Handler  // ping消息的处理函数
	SubProtocol string  // 子协议
}

type Upgrader struct {
	ReadBufferSize  int                      // 读缓冲区的大小
	WriteBufferSize int                      // 写缓冲区的大小
	CheckOrigin     func(*http.Request) bool // 检查请求的来源是否合法
}



func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (conn MyConn, err error) {
	// 创建一个MyConn
	conn = MyConn{
		ReadLimit: u.ReadBufferSize,
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
	u.CheckOrigin=CheckOriginFunc
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


	// 处理协议拓展

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

	// 回复报文 一系列请求头
	var resp []byte
	resp = append(resp, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n "...)
	//Sec-WebSocket-Accept：
	//Sec-WebSocket-Protocol:
	//请求头写完别忘了换行
	resp = append.........(省略号是代表我懒得抄了)

	//将请求报文写入
	_, err = conn.conn.Write(resp)
	return
}

func (c *MyConn) ReadMsg() (m Msg, err error) {
	// 根据数据帧读取数据

	// 读取第一个字节
	firstByte := make([]byte, 1)
	// TODO 用位运算处理这些字节
	// FIN是否提示为最终消息
	// RSV1~3的协议拓展判断

	// 读取第二个字节
	// 检查是否使用mask
	// 然后掩码处理
	if mask != 1 {
		err = ErrNoMask
		return
	}

	// 处理payload len
	switch {
	case 125 >= payloadLen && payloadLen > 0:
	case payloadLen == 126:
	case payloadLen == 127:
	default:
		// 都不是？ 那发个锤
		return
	}

	// 如果你有ReadLimit这个功能 该咋搞呢

	m.Typ = int(opcode)
	switch opcode {
	case PingMessage:
		// 按照用户设置的执行
	case PongMessage:
	case TextMessage:
	case BinaryMessage:
	case CloseMessage:
	default:
	}
	return
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
