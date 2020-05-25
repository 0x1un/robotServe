package main

import (
	"log"
	"net/http"
	"os"

	"github.com/0x1un/godingtalk"
	"github.com/0x1un/robotServe/utils"
)

var (
	config  *utils.Config
	ding    *godingtalk.DingtalkClient
	logger  *log.Logger
	loggerx *log.Logger
	errFile *os.File
	accFile *os.File
)

func init() {
	errFile, err := os.OpenFile("errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger = log.New(errFile, "[Error] ", log.LstdFlags)

	accFile, err := os.OpenFile("access.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	loggerx = log.New(accFile, "[Info] ", log.LstdFlags)

	conf, err := utils.ReadConfig()
	if err != nil {
		logger.Panic(err)
	}
	config = conf
	ding = godingtalk.NewDingtalkClient(config.Appkey, config.AppSecret)
	if err := ding.Init(); err != nil {
		logger.Panic(err)
	}
}

func main() {
	// free file handler
	defer errFile.Close()
	defer accFile.Close()

	http.HandleFunc("/", robot)
	loggerx.Println("listen on ::444")
	if err := http.ListenAndServeTLS(":444", "/root/pki/server.pem", "/root/pki/server.key", nil); err != nil {
		logger.Fatalln(err)
	}
}
