package ws

import (
	"encoding/json"
	"log"
	"time"

	"imaptool/common"
	"imaptool/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func NewVisitorServer(c *gin.Context) {
	// go kefuServerBackend()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// 获取GET参数,创建WS
	vistorInfo := models.FindVisitorByVistorId(c.Query("visitor_id"))
	if vistorInfo.VisitorId == "" {
		c.JSON(200, gin.H{
			"code": 400,
			"msg":  "访客不存在",
		})
		return
	}
	user := &User{
		Conn:       conn,
		Name:       vistorInfo.Name,
		Avator:     vistorInfo.Avator,
		Id:         vistorInfo.VisitorId,
		To_id:      vistorInfo.ToId,
		UpdateTime: time.Now(),
	}
	go models.UpdateVisitorStatus(vistorInfo.VisitorId, 1)
	// go SendServerJiang(vistorInfo.Name, "来了", c.Request.Host)

	AddVisitorToList(user)

	for {
		// 接受消息
		var receive []byte
		messageType, receive, err := conn.ReadMessage()
		if err != nil {
			for _, visitor := range ClientList {
				if visitor.Conn == conn {
					log.Println("删除用户", visitor.Id)
					delete(ClientList, visitor.Id)
					VisitorOffline(visitor.To_id, visitor.Id, visitor.Name)
				}
			}
			log.Println(err)
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

func AddVisitorToList(user *User) {
	// 用户id对应的连接
	oldUser, ok := ClientList[user.Id]
	if oldUser != nil || ok {
		msg := TypeMessage{
			Type: "close",
			Data: user.Id,
		}
		str, _ := json.Marshal(msg)
		if err := oldUser.Conn.WriteMessage(websocket.TextMessage, str); err != nil {
			oldUser.Conn.Close()
			user.UpdateTime = oldUser.UpdateTime
			delete(ClientList, user.Id)
		}
	}
	ClientList[user.Id] = user
	lastMessage := models.FindLastMessageByVisitorId(user.Id)
	userInfo := make(map[string]string)
	userInfo["uid"] = user.Id
	userInfo["username"] = user.Name
	userInfo["avator"] = user.Avator
	userInfo["last_message"] = lastMessage.Content
	if userInfo["last_message"] == "" {
		userInfo["last_message"] = "新访客"
	}
	msg := TypeMessage{
		Type: "userOnline",
		Data: userInfo,
	}
	str, _ := json.Marshal(msg)

	// 新版
	OneKefuMessage(user.To_id, str)
}

func VisitorOnline(kefuId string, visitor models.Visitor) {
	lastMessage := models.FindLastMessageByVisitorId(visitor.VisitorId)
	userInfo := make(map[string]string)
	userInfo["uid"] = visitor.VisitorId
	userInfo["username"] = visitor.Name
	userInfo["avator"] = visitor.Avator
	userInfo["last_message"] = lastMessage.Content
	if userInfo["last_message"] == "" {
		userInfo["last_message"] = "新访客"
	}
	msg := TypeMessage{
		Type: "userOnline",
		Data: userInfo,
	}
	str, _ := json.Marshal(msg)
	OneKefuMessage(kefuId, str)
}

func VisitorOffline(kefuId string, visitorId string, visitorName string) {
	models.UpdateVisitorStatus(visitorId, 0)
	userInfo := make(map[string]string)
	userInfo["uid"] = visitorId
	userInfo["name"] = visitorName
	msg := TypeMessage{
		Type: "userOffline",
		Data: userInfo,
	}
	str, _ := json.Marshal(msg)
	// 新版
	OneKefuMessage(kefuId, str)
}

func VisitorNotice(visitorId string, notice string) {
	msg := TypeMessage{
		Type: "notice",
		Data: notice,
	}
	str, _ := json.Marshal(msg)
	visitor, ok := ClientList[visitorId]
	if !ok || visitor == nil || visitor.Conn == nil {
		return
	}
	visitor.Conn.WriteMessage(websocket.TextMessage, str)
}

// 客服发信息给用户;iskefu:no
func VisitorMessage(visitorId, content string, kefuInfo models.User) {
	msg := TypeMessage{
		Type: "message",
		Data: ClientMessage{
			Name:    kefuInfo.Nickname,
			Avator:  kefuInfo.Avator,
			Id:      kefuInfo.Name,
			Time:    time.Now().Format("2006-01-02 15:04:05"),
			ToId:    visitorId,
			Content: content,
			IsKefu:  "no",
		},
	}
	str, _ := json.Marshal(msg)
	visitor, ok := ClientList[visitorId]
	if !ok || visitor == nil || visitor.Conn == nil {
		return
	}
	visitor.Conn.WriteMessage(websocket.TextMessage, str)
}

// 自动回复客服
func VisitorAutoReply(vistorInfo models.Visitor, kefuInfo models.User, content string) {
	kefu, ok := KefuList[kefuInfo.Name]
	reply := models.FindReplyItemByUserIdTitle(kefuInfo.Name, content)
	if reply.Content != "" {
		time.Sleep(1 * time.Second)
		// 发给用户
		VisitorMessage(vistorInfo.VisitorId, reply.Content, kefuInfo)
		// 发给客服
		KefuMessage(vistorInfo.VisitorId, reply.Content, kefuInfo)
		// 数据库记录信息
		models.CreateMessage(kefuInfo.Name, vistorInfo.VisitorId, reply.Content, "kefu")
	}

	// 客服不在线
	if !ok || kefu == nil {
		time.Sleep(1 * time.Second)
		welcome := models.FindConfig("OfflineMessage")
		if welcome == "" || reply.Content != "" {
			return
		}
		// 发给用户
		VisitorMessage(vistorInfo.VisitorId, welcome, kefuInfo)
		// 数据库记录信息
		models.CreateMessage(kefuInfo.Name, vistorInfo.VisitorId, welcome, "kefu")
	}
}

func CleanVisitorExpire() {
	go func() {
		log.Println("cleanVisitorExpire start...")
		for {
			for _, user := range ClientList {
				diff := time.Now().Sub(user.UpdateTime).Seconds()
				if diff >= common.VisitorExpire {
					msg := TypeMessage{
						Type: "auto_close",
						Data: user.Id,
					}
					str, _ := json.Marshal(msg)
					if err := user.Conn.WriteMessage(websocket.TextMessage, str); err != nil {
						user.Conn.Close()
						delete(ClientList, user.Id)
					}
					log.Println(user.Name + ":cleanVisitorExpire finshed")
				}
			}
			t := time.NewTimer(time.Second * 5)
			<-t.C
		}
	}()
}
