package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0x1un/godingtalk"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type userSchedulerList struct {
	classList map[string]string
	users     map[string]string
}

var (
	weekday = [7]string{
		"星期天",
		"星期一",
		"星期二",
		"星期三",
		"星期四",
		"星期五",
		"星期六",
	}
)

type userShift struct {
	className string
	dateTime  time.Time
}

const (
	format = "2006-01-02"
)

type nameSlice []string

func (s nameSlice) Len() int      { return len(s) }
func (s nameSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s nameSlice) Less(i, j int) bool {
	a, _ := UTF82GBK(s[i])
	b, _ := UTF82GBK(s[j])
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

func zellerWeek(year, month, day uint16) string {
	var y, m, c uint16
	if month >= 3 {
		m = month
		y = year % 100
		c = year / 100
	} else {
		m = month + 12
		y = (year - 1) % 100
		c = (year - 1) / 100
	}
	week := y + (y / 4) + (c / 4) - 2*c + ((26 * (m + 1)) / 10) + day - 1
	if week < 0 {
		week = 7 - (-week)%7

	} else {
		week = week % 7
	}
	which_week := int(week)
	return weekday[which_week]
}

//UTF82GBK : transform UTF8 rune into GBK byte array
func UTF82GBK(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}

//GBK2UTF8 : transform  GBK byte array into UTF8 string
func GBK2UTF8(src []byte) (string, error) {
	GB18030 := simplifiedchinese.All[0]
	bytes, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(src), GB18030.NewDecoder()))
	return string(bytes), err
}

type weekList map[string][]userShift

func getLeaveStatus(useridList []string, startDate, endDate int64) map[string]struct {
	startTime int64
	endTime   int64
} {
	resp, err := ding.OapiAttendanceGetleavestatusRequest(useridList, startDate, endDate, 0, 20)
	if err != nil {
		logger.Panicln(err)
	}
	m := make(map[string]struct {
		startTime int64
		endTime   int64
	}, len(resp.Result.LeaveStatus))
	for _, v := range resp.Result.LeaveStatus {
		m[v.Userid] = struct {
			startTime int64
			endTime   int64
		}{v.StartTime, v.EndTime}
	}
	return m
}

func getTimeFromMs(ms int64) string {
	return time.Unix(0, ms*1e6).Format("2006-01-02 15:04:05")
}

func fillTemplate(wl weekList) string {
	buffer := &strings.Builder{}
	cssData, err := ioutil.ReadFile("atten.css")
	if err != nil {
		panic(err)
	}
	sortName := make(nameSlice, 0)
	for name := range wl {
		sortName = append(sortName, name)
	}
	sort.Strings(sortName)

	buffer.WriteString(fmt.Sprintf(`<head><style>%s%s</style></head><table class=topazCells><tbody><tr>`, string(cssData), `@font-face {
        font-family: 楷体;
        src: url('simkai.ttf');
    `))
	date := false

	for _, name := range sortName {
		if !date {
			buffer.WriteString("<td>姓名/日期</td>")
			for _, class := range wl[name] {
				dat := class.dateTime.Format(format)
				weekdy := class.dateTime.Weekday()
				buffer.WriteString(fmt.Sprintf("<td>%s\n%s</td>", dat, weekday[weekdy]))
			}
			date = true
			buffer.WriteString(`</tr>`)
		}
		buffer.WriteString(fmt.Sprintf("<tr><td>%s</td>", name))
		for _, class := range wl[name] {
			buffer.WriteString(fmt.Sprintf("<td>%s</td>", class.className))
		}
		buffer.WriteString("</tr>")
	}
	buffer.WriteString(`</tbody></table>`)
	return buffer.String()
}
func findFile(filename string) bool {
	fileInfo, err := ioutil.ReadDir("gen/")
	if err != nil {
		logger.Println(err)
	}
	for _, v := range fileInfo {
		if v.Name() == filename {
			return true
		}
	}
	return false
}

func getFileHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func queryDepartmentUserLeaveByDay(day int) string {
	if day == 0 {
		day--
	}
	umap := getDepartmentUsers()
	uidlist := convertMapKeyOrValueToList(umap, true)
	now := time.Now()
	date := now.AddDate(0, 0, day)
	fdate := date.UnixNano() / 1e6
	edate := date.AddDate(0, 0, now.Day()-date.Day()).UnixNano() / 1e6
	leavelist := getLeaveStatus(uidlist, fdate, edate)
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("从%s开始的所有请假人员:\n\n", date.Format(format)))
	for k, v := range leavelist {
		buf.WriteString(fmt.Sprintf("%s: %s -> %s\n", umap[k], parseUnixNano2Human(v.startTime*1e6), parseUnixNano2Human(v.endTime*1e6)))
	}
	return buf.String()
}

// 批量查询部门成员最近七天的班次
func queryDepartmentUserSchedulerListWeeks(at int) string {
	xDate := time.Now().AddDate(0, 0, at)
	userMap := getDepartmentUsers()
	data, err := ioutil.ReadFile(config.ClassFile)
	if err != nil {
		panic(err)
	}
	class := make(map[int]string)
	if err := json.Unmarshal(data, &class); err != nil {
		panic(err)
	}
	fromTime := xDate.UnixNano() / 1e6
	endTime := xDate.AddDate(0, 0, 6).UnixNano() / 1e6
	useridList := convertMapKeyOrValueToList(userMap, true)
	userStr := strings.Join(useridList, ",")
	leaveList := getLeaveStatus(useridList, fromTime, endTime)
	resp, err := ding.OapiAttendanceScheduleListbyusersRequest(config.OpUserID, userStr, fromTime, endTime)
	if err != nil {
		logger.Println(err)
		return ""
	}
	wl := make(map[string][]userShift, len(userMap))
	for i, v := range resp.Result {
		if v.CheckType == "OffDuty" {
			continue
		}
		workDate, err := time.Parse(format+" 15:04:05", v.WorkDate)
		if err != nil {
			logger.Println(err)
		}
		wd := workDate.UnixNano() / 1e6
		u := userShift{
			className: func(sid int) string {
				if sid == 0 {
					return "休息"
				}
				return class[sid]
			}(v.ShiftID),
			dateTime: workDate,
		}
		// 修复加班到第二天导致的排班表错乱重复的现象
		if x := i + 1; x < len(resp.Result) && v.IsRest == "Y" && resp.Result[x].WorkDate == v.WorkDate {
			continue
		}
		// 添加查询当天是否有请假
		wd = (wd - 28800000) * 1e6
		if leave, ok := leaveList[v.Userid]; ok && time.Unix(0, wd).Format(format) == time.Unix(0, leave.startTime*1e6).Format(format) {
			wdt := time.Unix(0, wd)
			sdt := time.Unix(0, leave.startTime*1e6)
			edt := time.Unix(0, leave.endTime*1e6)
			// 请假开始时间在当前工作时间之后，并且，请假结束时间在当前工作日之前
			if sdt.After(wdt) && edt.After(sdt) {
				u.className += "(请假)"
			}
			delete(leaveList, v.Userid)
		}
		wl[userMap[v.Userid]] = append(wl[userMap[v.Userid]], u)
	}
	oput := fillTemplate(wl)
	filename := getFileHash([]byte(oput)) + ".png"
	imgOpt := ImageOptions{Input: "-", Format: "png", Output: "gen/" + filename, Html: oput, BinaryPath: `/usr/local/bin/wkhtmltoimage`, Height: 400, Width: 700}
	output, err := GenerateImage(&imgOpt)
	if err != nil {
		logger.Println(err)
	}
	logger.Println(output)
	return filename
}

// getDepartmentUserSchedulerListWeeks 查询最近一周部门成员排班
// 此方法使用的单个天数获取 (已经弃用了)
func getDepartmentUserSchedulerListWeeks() {
	userMap := getDepartmentUsers()
	wl := make(map[string][]userShift, len(userMap))
	for id, name := range userMap {
		class := []userShift{}
		// currTime := time.Time{}
		for day := 0; day < 7; day++ {
			tm := time.Now()
			tm = tm.AddDate(0, 0, day)
			resp, err := ding.OapiAttendanceScheduleListbydayRequest(config.OpUserID, id, tm.UnixNano()/1e6)
			if err != nil {
				panic(err)
			}
			// currTime = tm
			for _, v := range resp.Result {
				if v.ClassName == "" || v.ClassID == 0 {
					class = append(class, userShift{
						className: "休息",
						dateTime:  tm,
					})
					continue
				}

				if v.CheckType == "OnDuty" {
					continue
				}
				class = append(class, userShift{
					className: v.ClassName,
					dateTime:  tm,
				})
			}
		}
		wl[name] = class
	}

	imgOpt := ImageOptions{Input: "-", Format: "png", Output: "gen/example.png", Html: fillTemplate(wl), BinaryPath: `/usr/local/bin/wkhtmltoimage`, Height: 400, Width: 700}
	output, err := GenerateImage(&imgOpt)
	if err != nil {
		logger.Println(err)
	}
	logger.Println(output)
}

func parseUnixNano2Human(tm int64) string {
	tme := time.Unix(0, tm)
	timestr := tme.Format("2006-01-02 15:04:05")
	if strings.HasPrefix(timestr, "1970-01-01") {
		return strings.Split(timestr, " ")[1]
	}
	return timestr
}

func parseTimeRetHour(tm string) string {
	t, e := time.Parse("2006-01-02 15:04:05", tm)
	if e != nil {
		logger.Println(e)
	}
	return t.Format("15:04:05")
}
func convertMapKeyOrValueToList(m map[string]string, key bool) (ids []string) {
	for k, v := range m {
		if key {
			ids = append(ids, k)
		} else {
			ids = append(ids, v)
		}
	}
	return
}

// fetchUsersScheListByDays 获取部门成员的排班
// 因为批量查询的api没有返回classid和class name，所以只能使用单个查询api
func fetchUsersScheListByDays(day int) userSchedulerList {
	ts := time.Now().AddDate(0, 0, day)
	users := getDepartmentUsers()
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
	return userSchedulerList{
		classList: classList,
		users:     users,
	}
}

func fetchUsersScheList(day int, all bool) string {
	userScheduler := fetchUsersScheListByDays(day)
	buffer := &strings.Builder{}
	ts := time.Now().AddDate(0, 0, day)
	buffer.WriteString(ts.Format("2006-01-02") + "\n")
	return getSignStatus(userScheduler.users, userScheduler.classList, ts, ts, all, buffer)
}

func removeSpaceLine(s *strings.Builder) string {
	s1 := s.String()
	s1 = strings.TrimSpace(s1)
	s.Reset()
	for i := 0; i < len(s1); i++ {
		if s1[i] == 10 && s1[i+1] == 10 {
			continue
		}
		s.WriteByte((s1[i]))
	}
	return s.String()
}

func getSignStatus(users, classList map[string]string, ftm, ttm time.Time, all bool, buffer *strings.Builder) string {
	userids := make([]string, 0, len(users))
	for userid := range users {
		userids = append(userids, userid)
	}
	resp, err := ding.OapiAttendanceListRequest(userids, ftm.Format("2006-01-02 00:00:00"), ttm.Format("2006-01-02 00:00:00"), 0, 50)
	if err != nil {
		logger.Println(err)
	}
	n := len(resp.Recordresult)
	if all || n == 0 {
		for k, v := range classList {
			buffer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
		return buffer.String()
	}
	offs := make(map[string]string)
	onss := make(map[string]bool)
	for _, v := range resp.Recordresult {
		user := users[v.UserID]
		if v.CheckType == "OnDuty" {
			if buffer.String()[buffer.Len()-1] != '\n' {
				buffer.WriteString("\n")
			}
			buffer.WriteString(fmt.Sprintf("%s: %s • (%s)", user, classList[users[v.UserID]], hour(v.UserCheckTime)))
			if off, ok := offs[user]; ok {
				buffer.WriteString(off)
			}
			onss[user] = true
		} else if v.CheckType == "OffDuty" {
			of := fmt.Sprintf("/[%s]\n", hour(v.UserCheckTime))
			offs[user] = of
			if _, ok := onss[user]; ok {
				buffer.WriteString(of)
			}
		}
	}
	return removeSpaceLine(buffer)
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

// getDepartmentUsers 这里不返回错误，获取不全直接抛出异常
func getDepartmentUsers() map[string]string {
	users := struct {
		sync.RWMutex
		user map[string]string
	}{
		user: make(map[string]string),
	}
	var (
		offset int64 = 0
		size   int64 = 100
	)
	for {
		resp, err := req(ding.OapiUserSimplelistRequest, config.DepartmentID, offset, size, "entry_desc")
		if err != nil {
			logger.Warningln(err)
		}
		users.Lock()
		appendUsers(users.user, resp)
		if !resp.HasMore {
			appendUsers(users.user, resp)
			break
		} else {
			offset += size
			resp, err := req(ding.OapiUserSimplelistRequest, config.DepartmentID, offset, size, "entry_desc")
			if err != nil {
				logger.Warningln(err)
			}
			appendUsers(users.user, resp)
		}
		users.Unlock()
	}
	return users.user
}

func req(f func(string, int64, int64, string) (godingtalk.UserSimplelistResp, error), a string, b, c int64, d string) (godingtalk.UserSimplelistResp, error) {
	return f(a, b, c, d)
}

func appendUsers(box map[string]string, s godingtalk.UserSimplelistResp) {
	for _, v := range s.Userlist {
		box[v.Userid] = v.Name
	}
}

func getAttendanceGroups() []godingtalk.AttendanceGetsimplegroupsResp {
	groupList := []godingtalk.AttendanceGetsimplegroupsResp{}
	getSimpleGroupsReq := ding.OapiAttendanceGetsimplegroupsRequest
	for offset, size := 0, 10; ; offset = size + offset {
		resp, err := getSimpleGroupsReq(int64(offset), int64(size))
		if err != nil {
			// 这里可能会获取不完整，数据量过大的情况下, 直接忽略
			logger.Warningf("获取考勤组失败", err.Error())
			continue
		}
		groupList = append(groupList, resp)
		if !resp.Result.HasMore {
			break
		}
	}
	return groupList
}
