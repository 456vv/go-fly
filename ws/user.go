package ws

import (
	"encoding/json"
	"log"
	"time"

	"imaptool/models"
	"imaptool/tools"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func NewKefuServer(c *gin.Context) {
	kefuId, _ := c.Get("kefu_id")
	kefuInfo := models.FindUserById(kefuId)
	if kefuInfo.ID == 0 {
		c.JSON(200, gin.H{
			"code": 400,
			"msg":  "用户不存在",
		})
		return
	}

	// go kefuServerBackend()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// 获取GET参数,创建WS
	var kefu User
	kefu.Id = kefuInfo.Name
	kefu.Name = kefuInfo.Nickname
	kefu.Avator = kefuInfo.Avator
	kefu.Role_id = kefuInfo.RoleId
	kefu.Conn = conn
	AddKefuToList(&kefu)

	for {
		// 接受消息
		var receive []byte
		messageType, receive, err := conn.ReadMessage()
		if err != nil {
			log.Println("ws/user.go ", err)
			conn.Close()
			return
		}

		message <- &Message{
			conn:        conn,
			content:     receive,
			context:     c,
			messageType: messageType,
		}
	}
}

func AddKefuToList(kefu *User) {
	oldUser, ok := KefuList[kefu.Id]
	if oldUser != nil || ok {
		msg := TypeMessage{
			Type: "close",
			Data: kefu.Id,
		}
		str, _ := json.Marshal(msg)
		if err := oldUser.Conn.WriteMessage(websocket.TextMessage, str); err != nil {
			oldUser.Conn.Close()
		}
	}
	KefuList[kefu.Id] = kefu
}

// 给指定客服发消息;iskefu: custom
func OneKefuMessage(toId string, str []byte) bool {
	kefu, ok := KefuList[toId]
	if ok {
		kefu.Mux.Lock()
		defer kefu.Mux.Unlock()
		error := kefu.Conn.WriteMessage(websocket.TextMessage, str)
		tools.Logger().Println("send_kefu_message", error, string(str))
	}
	return ok
}

// 用户发信息给客服;iskefu:yes
func KefuMessage(visitorId, content string, kefuInfo models.User) {
	msg := TypeMessage{
		Type: "message",
		Data: ClientMessage{
			Name:    kefuInfo.Nickname,
			Avator:  kefuInfo.Avator,
			Id:      visitorId,
			Time:    time.Now().Format("2006-01-02 15:04:05"),
			ToId:    visitorId, // 这里/????/
			Content: content,
			IsKefu:  "yes",
		},
	}
	str, _ := json.Marshal(msg)
	OneKefuMessage(kefuInfo.Name, str)
}

// 判断客服端是否在线
func SendPingToKefuClient() {
	msg := TypeMessage{
		Type: "many pong",
	}
	str, _ := json.Marshal(msg)
	for kefuId, kefu := range KefuList {
		if kefu == nil {
			continue
		}
		kefu.Mux.Lock()
		defer kefu.Mux.Unlock()
		err := kefu.Conn.WriteMessage(websocket.TextMessage, str)
		if err != nil {
			log.Println("定时发送ping给客服，失败", err.Error())
			delete(KefuList, kefuId)
		}
	}
}
