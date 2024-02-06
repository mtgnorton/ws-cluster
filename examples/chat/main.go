package main

import (
	"encoding/json"

	"github.com/mtgnorton/ws-cluster/shared"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/gorilla/websocket"
	"github.com/mtgnorton/ws-cluster/message/queuemessage"
	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"log"
)

var onlineUsers = make(map[string]bool)

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
func main() {

	chatServer := g.Server("ws-chat")
	chatServer.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(MiddlewareCORS)
		group.GET("/token", func(r *ghttp.Request) {
			pid, uid := r.Get("pid").String(), r.Get("uid").String()
			token, err := shared.DefaultJwtWs.Generate(pid, uid, 0)
			if err != nil {
				r.Response.WriteJson(g.Map{"error": err})
				return
			}
			r.Response.WriteJson(g.Map{"token": token})
		})
	})
	chatServer.SetServerRoot("")
	chatServer.SetPort(8089)
	go chatServer.Run()

	// 定义WebSocket服务器的地址
	serverURL := "ws://localhost:8084/connect?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOiIxIiwidWlkIjoiOTk5OTkiLCJjbGllbnRfdHlwZSI6MSwiZXhwIjoyMTc5NzQ0NTM0LCJpc3MiOiJ3cy1jbHVzdGVyIn0.lDVixBSnT9nK7gsVvHsUtDk8qXA9BdwhX7Y2bqrrvtg"

	// 建立WebSocket连接
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("连接到WebSocket服务器失败:", err)
	}
	defer conn.Close()

	// 启动一个goroutine用于接收服务器的消息
	go func() {
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println("接收消息错误:", err)
				return
			}
			log.Println("收到:", string(messageBytes))
			wsMsg, err := wsmessage.DefaultProcessor.Decode(messageBytes)
			if err != nil {
				log.Printf("解析messageBytes消息错误:%v\n", err)
				continue
			}

			if wsMsg.Payload == nil {
				log.Println("wsMsg.Payload is nil")
				continue
			}

			payloadBytes, err := json.Marshal(wsMsg.Payload)
			if err != nil {
				log.Printf("解析wsMsg.Payload消息错误:%v\n", err)
				continue
			}
			if len(payloadBytes) == 2 {
				log.Println("payloadBytes length  is 2")
				continue
			}

			msg, err := queuemessage.DefaultProcessor.Decode(payloadBytes)
			if err != nil {
				log.Printf("解析payloadBytes消息错误:%v\n", err)
				continue
			}
			switch msg.Type {
			case queuemessage.TypeConnect:
				concreteMsg, err := queuemessage.ParseConnect(msg.Payload)
				if err != nil {
					log.Printf("解析TypeConnect消息错误:%v\n", err)
					continue
				}
				onlineUsers[concreteMsg.UID] = true
			case queuemessage.TypeDisconnect:
				concreteMsg, err := queuemessage.ParseDisConnect(msg.Payload)
				if err != nil {
					log.Printf("解析TypeDisconnect消息错误:%v\n", err)
					continue
				}
				onlineUsers[concreteMsg.UID] = false
			case queuemessage.TypeRequest:
				content, err := parseSms(msg.Payload)
				if err != nil {
					log.Printf("解析concreteMsg.Content消息错误:%v\n", err)
					continue
				}
				if _, ok := onlineUsers[content.To]; ok {
					log.Println("用户在线")
					pushMsg := NewPushMsg(content.To, content)
					pushMsgBytes, err := json.Marshal(pushMsg)
					if err != nil {
						return
					}
					log.Printf("push msg:%s \n", pushMsgBytes)
					err = conn.WriteMessage(websocket.TextMessage, pushMsgBytes)
					if err != nil {
						log.Printf("WriteMessage err:%v \n", err)
					}
				} else {
					log.Println("用户不在线")
				}
			}
			log.Println("当前在线用户:", onlineUsers)

		}
	}()

	select {}
}

func NewPushMsg(uids string, data interface{}) interface{} {
	return wsmessage.Req{
		Type: wsmessage.TypePush,
		Payload: queuemessage.PayLoadPush{
			PID:  "1",
			UIDs: uids,
			CIDs: "",
			Tags: "",
			Data: data,
		},
	}
}
