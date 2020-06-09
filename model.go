package main

import "encoding/json"

// Content: dingtalk robot message
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
type AtModel struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
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

func NewMsgText(content string, atMobiles []string) []byte {
	msg := &MsgText{
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
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Panic(err)
	}
	return data
}

type Markdown struct {
	Msgtype  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
}

func markdown(title, msg string) []byte {
	md := &Markdown{
		Msgtype: "markdown",
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{
			Title: title,
			Text:  msg,
		},
	}
	data, err := json.Marshal(md)
	if err != nil {
		logger.Panic(err)
	}
	return data
}
