package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/0x1un/robotServe/insp"
)

const (
	stripChars = "][)(./\n"
)

func robot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			logger.Println(err)
		}
		originSign := r.Header.Get("Sign")
		localSign := sign(r.Header.Get("Timestamp"), config.RobotAppSecret)
		if !compareSign(originSign, localSign) {
			_, err := w.Write([]byte(`{"msgtype": "empty"}`))
			if err != nil {
				logger.Error(err)
			}
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
		if data := inspectCommand(resp); len(data) != 0 {
			dataLen, err := w.Write(data)
			if dataLen != len(data) || err != nil {
				logger.Errorf("消息写入失败: %v", err)
			}
		}
		if data := attenCommand(resp); len(data) != 0 {
			dataLen, err := w.Write(data)
			if dataLen != len(data) || err != nil {
				logger.Errorf("消息写入失败: %v", err)
			}
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

func inspectCommand(resp *Content) []byte {
	if strings.Trim(resp.Text.Content, stripChars+" ") != "inspect" {
		return nil
	}
	models, err := insp.Insp(resp.SenderStaffID, logger)
	if err != nil {
		return markdown("inspect command failed: ", err.Error())
	}
	for _, model := range models {
		resp, err := ding.OapiProcessinstanceCreateRequest(model)
		if err != nil {
			return markdown("inspect command failed: ", err.Error())
		}
		if resp.ErrCode != 0 {
			return markdown("inspect command failed: ", resp.ErrMsg)
		}
	}
	return markdown("inspect command", "网络巡检已生成，请务必认真审批！")
}

func attenCommand(resp *Content) []byte {
	cmdStr := strings.Trim(resp.Text.Content, stripChars)
	buffer := strings.Builder{}
	if strings.HasPrefix(cmdStr, "排班") || strings.HasPrefix(cmdStr, "shift") {
		// shift single
		cmdList := split(cmdStr, "./ ")
		all := false

		if strings.Contains(cmdStr, "leave") {
			if strings.ContainsAny(cmdStr, "0123456789") {
				for _, v := range cmdList {
					day, err := strconv.Atoi(v)
					if err != nil {
						continue
					}
					buffer.WriteString(queryDepartmentUserLeaveByDay(day))
				}
				return markdown("shift week command", buffer.String())
			} else {
				buffer.WriteString(
					queryDepartmentUserLeaveByDay(0),
				)
				return markdown("shift week command", buffer.String())
			}
		}

		if strings.Contains(cmdStr, "week") {
			if strings.Contains(cmdStr, "refresh") {
				if err := os.Remove(".class"); err != nil {
					return markdown("error:", err.Error())
				}
				if err := classToFile(); err != nil {
					return markdown("err: ", "failed to write class file")
				}
				return markdown("resp: ", "refresh class succeed!")
			}
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
