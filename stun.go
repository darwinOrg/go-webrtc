package dgwrtc

import (
	"fmt"
	"github.com/pion/stun/v2"
	"log"
	"net"
)

func StartStunServer(port int) {
	// 创建UDP监听器
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("STUN server started on", addr.String())

	// 监听并处理请求
	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		req := buffer[:n]
		if stun.IsMessage(req) {
			handleSTUNRequest(conn, addr, req)
		}
	}
}

func handleSTUNRequest(conn *net.UDPConn, addr *net.UDPAddr, req []byte) {
	log.Println("Received STUN request from", addr.String())

	// 解析STUN消息
	var message stun.Message
	err := stun.Decode(req, &message)
	if err != nil {
		log.Println("stun Decode message fail, err:", err)
		return
	}

	// 创建STUN Binding Success响应
	response := stun.MustBuild(
		&stun.XORMappedAddress{
			IP:   addr.IP,
			Port: addr.Port,
		},
	)

	// 发送STUN响应
	_, err = conn.WriteToUDP(response.Raw, addr)
	if err != nil {
		log.Println("Error sending STUN response:", err)
		return
	}

	log.Println("Sent STUN response to", addr.String())
}
