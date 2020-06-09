package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	stripChars = "][)(./\n"
	helpStr    = `
	指令集:
		获取指定天数的排班(n <= 0 || n >= 0)
			shift n => 获取第n天上班人员(如果已有打卡，则只显示打卡成员)
			shift n all => 获取第n天所有上班人员(不显示打卡时间)
		获取一周的排班(n <= 0 || n >= 0)
			shift week n => 获取从第n天开始的往后七天的排班
	`
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
		logger.Printf("%s sent a message: {{%s}} conversationType: %s, groupName: %s", resp.SenderNick, resp.Text.Content, el(resp.ConversationType == "1", "私聊", "群聊"), resp.ConversationTitle)
		if data := attenCommand(resp); len(data) != 0 {
			w.Write(data)
		}
	}
}

func loopCmdList(cmdList []string, typ string, all bool, buffer *strings.Builder) {
	for _, day := range cmdList {
		if dy, err := strconv.Atoi(day); err != nil {
			continue
		} else {
			if typ == "week" {
				buffer.WriteString(fmt.Sprintf("![](%s)",
					config.ImgHost+queryDepartmentUserSchedulerListWeeks(dy)))
			} else if typ == "sigle" {
				buffer.WriteString(fetchUsersScheList(dy, all))
			}
		}

	}
}

func attenCommand(resp *Content) []byte {
	cmdStr := strings.Trim(resp.Text.Content, stripChars)
	buffer := strings.Builder{}
	if strings.HasPrefix(cmdStr, "排班") || strings.HasPrefix(cmdStr, "shift") {
		// shift single
		cmdList := split(cmdStr, "./ ")
		all := false

		if strings.Contains(cmdStr, "week") {
			if strings.ContainsAny(cmdStr, "0123456789") {
				loopCmdList(cmdList, "week", false, &buffer)
				return markdown("shift week command", buffer.String())
			} else {
				buffer.WriteString(
					fmt.Sprintf("![](%s)", config.ImgHost+queryDepartmentUserSchedulerListWeeks(0)),
				)
				return markdown("shift week command", buffer.String())
			}

		}
		if strings.Contains(cmdStr, "all") {
			all = true
		}
		if strings.ContainsAny(cmdStr, "0123456789") {
			loopCmdList(cmdList, "sigle", all, &buffer)
		} else {
			buffer.WriteString(fetchUsersScheList(0, all))
		}
		msg := NewMsgText(strings.Trim(buffer.String(), "\n"), nil)
		return msg

	}
	return nil
}

func el(condition bool, tv, fv interface{}) interface{} {
	if condition {
		return tv
	}
	return fv
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
