package main

import (
	"net/http"
	"os"

	"github.com/0x1un/godingtalk"
	"github.com/0x1un/robotServe/utils"
	log "github.com/sirupsen/logrus"
)

var (
	file   *os.File
	config *utils.Config
	ding   *godingtalk.DingtalkClient
	logger = log.New()
)

func init() {
	file, err := os.OpenFile("logs/robotServe.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger.SetOutput(file)

	conf, err := utils.ReadConfig()
	if err != nil {
		logger.Panic("failed to readConfig", err)
	}
	config = conf
	ding = godingtalk.NewDingtalkClient(config.Appkey, config.AppSecret)
	if err := ding.Init(); err != nil {
		logger.Panic("failed to init dingtalk client", err)
	}

	if err := classToFile(); err != nil {
		logger.Panic("failed store class ot file", err)
	}

}

func main() {
	defer file.Close()
	http.HandleFunc("/", robot)
	logger.Println("listen on ::443")
	if err := http.ListenAndServeTLS(":443", "cert/server.pem", "cert/server.key", nil); err != nil {
		logger.Println(err)
	}
}
