package utils

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	RobotAppSecret string `json:"robot_app_secret"`
	OpUserID       string `json:"op_user_id"`
	Appkey         string `json:"appkey"`
	AppSecret      string `json:"appsecret"`
	DepartmentID   string `json:"department_id"`
	ClassFile      string `json:"class_file"`
	ImgHost        string `json:"image_host"`
}

func ReadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("conf/config.json")
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
