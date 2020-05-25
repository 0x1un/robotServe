package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/0x1un/godingtalk"
)

func fetchUsersScheList(day int) string {
	ts := time.Now().AddDate(0, 0, day)
	users, errs := getDepartmentUsers()
	for _, err := range errs {
		if err != nil {
			log.Fatal(err)
		}
	}
	classList := make(map[string]string)
	for userid, name := range users {
		resp, err := ding.OapiAttendanceScheduleListbydayRequest(config.OpUserID, userid, ts.UnixNano()/1e6)
		if err != nil {
			logger.Println(err)
		}
		for _, v := range resp.Result {
			if v.ClassName == "" {
				classList[name] = "休"
			} else {
				classList[name] = v.ClassName
			}
		}
	}
	buffer := &strings.Builder{}
	buffer.WriteString(ts.Format("2006-01-02") + "\n")
	return getSignStatus(users, classList, ts, ts, buffer)
}

func getSignStatus(users, classList map[string]string, ftm, ttm time.Time, buffer *strings.Builder) string {
	userids := make([]string, 0, len(users))
	for userid := range users {
		userids = append(userids, userid)
	}
	resp, err := ding.OapiAttendanceListRequest(userids, ftm.Format("2006-01-02 00:00:00"), ttm.Format("2006-01-02 00:00:00"), 0, 50)
	if err != nil {
		logger.Println(err)
	}
	n := len(resp.Recordresult)
	if n == 0 {
		for k, v := range classList {
			buffer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
		return buffer.String()
	}
	offs := make(map[string]string)
	onss := make(map[string]bool)
	for i, v := range resp.Recordresult {
		user := users[v.UserID]
		if v.CheckType == "OnDuty" {
			buffer.WriteString(fmt.Sprintf("%s: %s • (%s)", user, classList[users[v.UserID]], hour(v.UserCheckTime)))
			if off, ok := offs[user]; ok {
				buffer.WriteString(off)
			}
			onss[user] = true
			if i+1 < n && resp.Recordresult[i+1].CheckType != "OffDuty" {
				buffer.WriteString("\n")
			}
		} else if v.CheckType == "OffDuty" {
			of := fmt.Sprintf("/[%s]\n", hour(v.UserCheckTime))
			offs[user] = of
			if _, ok := onss[user]; ok {
				buffer.WriteString(of)
			}
		}
	}
	return buffer.String()
}

func hour(ts int64) string {
	return time.Unix(int64(ts)/1000, 0).Local().Format("15:04:05")
}
func classToFile() error {
	if _, err := os.Stat(config.ClassFile); err == nil {
		return nil
	}
	resp, err := ding.OapiAttendanceShiftListRequest(config.OpUserID, 0)
	if err != nil {
		return err
	}
	class := make(map[int]string, len(resp.Result.Result))
	for _, v := range resp.Result.Result {
		class[v.ID] = v.Name
	}
	data, err := json.Marshal(class)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(config.ClassFile, data, 0666); err != nil {
		return err
	}
	return nil
}

func readClass() (map[int]string, error) {
	data, err := ioutil.ReadFile(config.ClassFile)
	if err != nil {
		return nil, err
	}
	class := make(map[int]string)
	err = json.Unmarshal(data, &class)
	if err != nil {
		return nil, err
	}
	return class, nil
}

func getDepartmentUsers() (map[string]string, []error) {
	errs := []error{}
	users := make(map[string]string)
	var (
		offset int64 = 0
		size   int64 = 1
	)
	for {
		resp, err := req(ding.OapiUserSimplelistRequest, config.DepartmentID, offset, size, "entry_desc")
		if err != nil {
			errs = append(errs, err)
			continue
		}
		appendUsers(users, resp)
		if !resp.HasMore {
			appendUsers(users, resp)
			break
		} else {
			offset += size
			resp, err := req(ding.OapiUserSimplelistRequest, config.DepartmentID, offset, size, "entry_desc")
			if err != nil {
				errs = append(errs, err)
				continue
			}
			appendUsers(users, resp)
		}
	}
	return users, errs
}

func req(f func(string, int64, int64, string) (godingtalk.UserSimplelistResp, error), a string, b, c int64, d string) (godingtalk.UserSimplelistResp, error) {
	return f(a, b, c, d)
}

func appendUsers(box map[string]string, s godingtalk.UserSimplelistResp) {
	for _, v := range s.Userlist {
		box[v.Userid] = v.Name
	}
}
