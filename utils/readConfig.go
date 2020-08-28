package utils

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	AgentID        string `json:"agent_id"`
	DepartmentID   string `json:"department_id"`
	RobotAppSecret string `json:"robot_app_secret"`
	OpUserID       string `json:"op_user_id"`
	Appkey         string `json:"appkey"`
	AppSecret      string `json:"appsecret"`
	ClassFile      string `json:"class_file"`
	ImgHost        string `json:"image_host"`
	WebdavURI      string `json:"webdav_uri"`
	WebdavUSER     string `json:"webdav_user"`
	WebdavPASS     string `json:"webdav_pass"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func CheckErr(e []error) bool {
	return false
}
