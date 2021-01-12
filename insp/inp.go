package insp

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/0x1un/godingtalk"
	"github.com/0x1un/omtools/zbxgraph"
	"github.com/0x1un/robotServe/utils"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
)

var (
	cfg         *ini.File
	metaConfig  *utils.Config
	webdavURI   string
	webdavUSER  string
	webdavPASS  string
	cfgPath     interface{} = "conf/init.ini"
	metaCfgPath             = "conf/config.json"
	dateNow     string
)

var randRange = func(dirPath string) int {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logrus.Fatal(fmt.Errorf("缺少现场、策略目录: %s", err.Error()))
	}
	var cnt int
	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".jpg") {
			cnt++
		}
	}
	return rand.Intn((cnt/2)-1) + 1
}

func initConfigure(log *logrus.Logger) {
	rand.Seed(time.Now().UnixNano())
	var err error
	cfg, err = ini.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	metaConfig, err = utils.ReadConfig(metaCfgPath)
	if err != nil {
		log.Fatal(err)
	}
	webdavURI = metaConfig.WebdavURI
	webdavUSER = metaConfig.WebdavUSER
	webdavPASS = metaConfig.WebdavPASS
	dateNow = time.Now().Format("2006-01-02")
}

func Insp(opUID string, log *logrus.Logger) ([]godingtalk.ProcessinstanceCreateReq, error) {
	initConfigure(log)
	wc := gowebdav.NewClient(webdavURI, webdavUSER, webdavPASS)
	bytes, err := wc.Read("inspect.ini")
	if err != nil {
		log.Errorln(err)
	} else {
		cfgPath = bytes
	}
	outputMap, err := zbxgraph.Run(cfgPath, "gen/", true)
	if err != nil {
		return nil, err
	}
	// insps := getInspSection()
	inspectionsURIs := make(map[string][]string)

	for k, v := range outputMap {
		for _, v2 := range v {
			url := fmt.Sprintf("%s%s", metaConfig.ImgHost, func(s string) string {
				sp := strings.Split(s, "/")
				if len(sp) > 1 {
					return strings.Join(sp[1:], "/")
				}
				return ""
			}(v2))
			inspectionsURIs[k] = append(inspectionsURIs[k], url)
		}
	}

	return GenModel(inspectionsURIs, opUID), nil
}

func GenModel(insps map[string][]string, opUID string) []godingtalk.ProcessinstanceCreateReq {
	reqs := make([]godingtalk.ProcessinstanceCreateReq, 0)
	reqs = append(reqs, AliModel(insps, opUID))
	// reqs = append(reqs, VkSdbModel(insps, opUID))
	reqs = append(reqs, DiDiModel(insps, opUID))
	return reqs
}

func getFileBase(path string) string {
	return filepath.Base(path)
}

// getFilename: 获取路径的最终文件名，并去除后缀
func getFilename(path string) string {
	filename := filepath.Base(path)
	if sp := strings.Split(filename, "."); len(sp) > 1 {
		return sp[0]
	}
	return ""
}

// getInspSection: 获取配置文件中所有的巡检项
func getInspSection() map[string]map[string]string {
	ret := make(map[string]map[string]string)
	for _, section := range cfg.Sections() {
		if name := section.Name(); name == "GENERAL" ||
			name == "Default" || len(section.KeysHash()) == 0 ||
			name == "INSPECTION" {
			continue
		}
		ret[section.Name()] = section.KeysHash()
	}
	return ret
}
