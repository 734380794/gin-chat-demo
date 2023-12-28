package service

import (
	"encoding/json"
	"fmt"
	"gin-chat-demo/cache"
	"gin-chat-demo/conf"
	"gin-chat-demo/pkg/e"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const month = 60 * 60 * 24 * 30

// SendMsg 发送消息的结构体
type SendMsg struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

// ReplyMsg 回复消息的结构体
type ReplyMsg struct {
	From    string `json:"from"`
	Code    int    `json:"code"`
	Content string `json:"content"`
}

// Client 用户结构体
type Client struct {
	ID     string
	SendID string
	Socket *websocket.Conn
	Send   chan []byte
}

// Broadcast 广播类
type Broadcast struct {
	Client  *Client
	Message []byte
	Type    int
}

// ClientManager 用户管理
type ClientManager struct {
	Clients    map[string]*Client
	Broadcast  chan *Broadcast
	Reply      chan *Client
	Register   chan *Client
	Unregister chan *Client
}

// Message 信息转JSON
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManager{
	Clients:    make(map[string]*Client), // 参与连接的用户，出于性能的考虑，需要设置最大连接数
	Broadcast:  make(chan *Broadcast),
	Register:   make(chan *Client),
	Reply:      make(chan *Client),
	Unregister: make(chan *Client),
}

func CreateID(uid, toUid string) string {
	return uid + "->" + toUid
}
func Handler(c *gin.Context) {
	uid := c.Query("uid")
	toUid := c.Query("toUid")
	fmt.Println("-----ws开始启用-----")
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}).Upgrade(c.Writer, c.Request, nil) // 升级ws协议
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	// 创建一个用户实例
	client := &Client{
		ID:     CreateID(uid, toUid),
		SendID: CreateID(toUid, uid),
		Socket: conn,
		Send:   make(chan []byte),
	}
	// 用户注册到用户管理上
	Manager.Register <- client
	go client.Read()
	go client.Write()
}
func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		_ = c.Socket.Close()
	}()
	for {
		c.Socket.PongHandler()
		sendMsg := new(SendMsg)
		err := c.Socket.ReadJSON(&sendMsg)
		if err != nil {
			fmt.Println("数据格式有误", err)
			Manager.Unregister <- c
			_ = c.Socket.Close()
			break
		}
		if sendMsg.Type == 1 {
			r1, _ := cache.RedisClient.Get(c.ID).Result()
			r2, _ := cache.RedisClient.Get(c.SendID).Result()
			if r1 > "3" && r2 == "" {
				replyMsg := ReplyMsg{
					Code:    e.WebsocketLimit,
					Content: "达到限制",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
				_, _ = cache.RedisClient.Expire(c.ID, time.Hour*24*30).Result() // 防止重复骚扰，未建立连接刷新过期时间一个月
				continue
			} else {
				cache.RedisClient.Incr(c.ID)
				_, _ = cache.RedisClient.Expire(c.ID, time.Hour*24*24*30*3).Result()
			}
			Manager.Broadcast <- &Broadcast{
				Client:  c,
				Message: []byte(sendMsg.Content), // 发送过来的消息
			}
		} else if sendMsg.Type == 2 {
			// 获取历史消息
			results, _ := FindMany(conf.MongoDBName, c.SendID, c.ID, 10)
			for _, result := range results {
				replyMsg := ReplyMsg{
					From:    result.From,
					Content: fmt.Sprintf("%s", result.Msg),
				}
				msg, _ := json.Marshal(replyMsg)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}
func (c *Client) Write() {
	defer func() {
		_ = c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			replyMsg := ReplyMsg{
				Code:    e.WebsocketSuccessMessage,
				Content: fmt.Sprintf("%s", string(message)),
			}
			msg, _ := json.Marshal(replyMsg)
			_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
