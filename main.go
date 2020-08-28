package main

import (
	"net/http"
	"os"
	"time"

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
	logger.SetReportCaller(true)

	conf, err := utils.ReadConfig("conf/config.json")
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
	srv := &http.Server{
		Addr:         ":443",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/", robot)
	logger.Println("listen on ::443")
	if err := srv.ListenAndServeTLS("cert/server.pem", "cert/server.key"); err != nil {
		logger.Println(err)
	}
}
