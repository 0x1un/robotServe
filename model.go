package main

import "encoding/json"

type AtModel struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

// 需要返回的数据
type Msg struct {
	Msgtype  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At AtModel `json:"at"`
}

type Content struct {
	// 目前只支持text
	Msgtype string `json:"msgtype"`
	// 消息文本
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
	// 加密的消息ID
	MsgID string `json:"msgId"`
	// 消息的时间戳，单位ms
	CreateAt int64 `json:"createAt"`
	// 1-单聊、2-群聊
	ConversationType string `json:"conversationType"`
	// 加密的会话ID
	ConversationID string `json:"conversationId"`
	// 会话标题（群聊时才有）
	ConversationTitle string `json:"conversationTitle"`
	// 加密的发送者ID
	SenderID string `json:"senderId"`
	// 发送者昵称
	SenderNick string `json:"senderNick"`
	// 发送者当前群的企业corpId（企业内部群有）
	SenderCorpID string `json:"senderCorpId"`
	// 发送者在企业内的userid（企业内部群有）
	SenderStaffID string `json:"senderStaffId"`
	// 加密的机器人ID
	ChatbotUserID string `json:"chatbotUserId"`
	// 被@人的信息
	// dingtalkId: 加密的发送者ID
	// staffId: 发送者在企业内的userid（企业内部群有）
	AtUsers []struct {
		DingtalkID string `json:"dingtalkId"`
		StaffID    string `json:"staffId"`
	} `json:"atUsers"`
}

func NewMsg(msgtype, title, text string, atMobiles []string) *Msg {
	return &Msg{
		Msgtype: msgtype,
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{Text: text, Title: title},
		At: AtModel{
			AtMobiles: atMobiles,
			IsAtAll:   false,
		},
	}
}

type MsgText struct {
	Msgtype string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles     []string `json:"atMobiles"`
		AtDingtalkIds []string `json:"atDingtalkIds"`
		IsAtAll       bool     `json:"isAtAll"`
	} `json:"at"`
}

func NewMsgText(content string, atMobiles []string) *MsgText {
	return &MsgText{
		Msgtype: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: content,
		},
		At: struct {
			AtMobiles     []string `json:"atMobiles"`
			AtDingtalkIds []string `json:"atDingtalkIds"`
			IsAtAll       bool     `json:"isAtAll"`
		}{
			AtMobiles: atMobiles,
		},
	}
}

func marshalMsgText(content string, atMobiles []string) ([]byte, error) {
	return json.Marshal(NewMsgText(content, atMobiles))
}
