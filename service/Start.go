package service

import (
	"encoding/json"
	"fmt"
	"gin-chat-demo/conf"
	"gin-chat-demo/pkg/e"
	"github.com/gorilla/websocket"
)

func (manager *ClientManager) Start() {
	fmt.Println("-----start-----")
	for {
		fmt.Println("--------监听管道通信------")
		select {
		case conn := <-Manager.Register:
			fmt.Printf("有新链接 %v", conn.ID)
			Manager.Clients[conn.ID] = conn
			replyMsg := ReplyMsg{
				Code:    e.WebsocketSuccess,
				Content: "已经连接到服务器",
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)

		case conn := <-Manager.Unregister:
			fmt.Printf("连接失败%v", conn.ID)
			if _, ok := Manager.Clients[conn.ID]; ok {
				replyMsg := ReplyMsg{
					Code:    e.WebsocketEnd,
					Content: "连接中断",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
				close(conn.Send)
				delete(Manager.Clients, conn.ID)
			}
		case broadcast := <-Manager.Broadcast:
			message := broadcast.Message
			sendId := broadcast.Client.SendID
			flag := false // 默认对方是不在线
			for id, conn := range Manager.Clients {
				if id != sendId {
					continue
				}
				select {
				case conn.Send <- message:
					flag = true
				default:
					close(conn.Send)
					delete(Manager.Clients, conn.ID)
				}
			}
			id := broadcast.Client.ID
			if flag {
				replyMsg := ReplyMsg{
					Code:    e.WebsocketOnlineReply,
					Content: "对方在线应答",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err := InsertMsg(conf.MongoDBName, id, string(message), 1, int64(3*month))
				if err != nil {
					fmt.Println("InsetOne Err", err)
				}
			}
		}
	}
}
