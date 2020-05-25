package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func robot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			logger.Println(err)
		}
		originSign := r.Header.Get("Sign")
		localSign := sign(r.Header.Get("Timestamp"), config.RobotAppSecret)
		if !compareSign(originSign, localSign) {
			w.Write([]byte(`{"msgtype": "empty"}`))
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			logger.Println(err)
		}
		resp := &Content{}
		err = json.Unmarshal(body, resp)
		if err != nil {
			logger.Println(err)
		}
		// 记录消息记录
		if resp.ConversationType == "1" {
			loggerx.Printf("%s sent a message: %s conversationType: %s\n", resp.SenderNick, resp.Text.Content, getConversation(resp.ConversationType))
		} else if resp.ConversationType == "2" {
			loggerx.Printf("%s sent a message: %s conversationType: %s, groupName: %s\n", resp.SenderNick, resp.Text.Content, getConversation(resp.ConversationType), resp.ConversationTitle)
		}
		if attenCmd := strings.Trim(resp.Text.Content, "\n\r\t.?*,，。!'\"()[]【】"); strings.Contains(attenCmd, "排班") {
			days := split(attenCmd, "\n/。，. ")
			if len(days) > 3 {
				if msg, err := marshalMsgText("一次性最多查询两天的排班", nil); err != nil {
					logger.Fatal(err)
				} else {
					w.Write(msg)
					return
				}
			}
			buffer := strings.Builder{}
			for _, day := range days[1:] {
				d, err := strconv.Atoi(day)
				if err != nil {
					logger.Println(err)
				}
				buffer.WriteString(fetchUsersScheList(d) + "\n\n")
			}
			msg := NewMsgText(strings.Trim(buffer.String(), "\n"), nil)
			data, err := json.Marshal(msg)
			if err != nil {
				logger.Println(err)
			}
			w.Write(data)
		}
	}
}

func getConversation(s string) string {
	if s == "1" {
		return "私聊"
	}
	return "群聊"
}

func split(s, seps string) []string {
	sp := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, sp)
}

// header中的timestamp + "\n" + 机器人的appSecret 当做签名字符串，使用HmacSHA256算法计算签名，然后进行Base64 encode，得到最终的签名值。
func sign(ts string, appsecret string) string {
	h := hmac.New(sha256.New, []byte(appsecret))
	_, _ = h.Write([]byte(ts + "\n" + appsecret))
	enc := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return enc
}

func compareSign(origin, local string) bool {
	return origin == local
}
