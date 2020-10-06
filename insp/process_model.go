package insp

import (
	"fmt"
	"strings"

	"github.com/0x1un/godingtalk"
)

var (
	getElem = func(s string, at int) string {
		sp := strings.Split(s, "|")
		if len(sp) > 1 {
			return sp[at]
		}
		return ""
	}
)

func AliModel(components map[string][]string, opUID string) godingtalk.ProcessinstanceCreateReq {
	var fvs godingtalk.FormComponentValues
	var subFvs godingtalk.FormComponentValues
	subTemplateName := []string{
		"联通外网|流量图1",
		"电信外网|流量图2",
		"CD-HZ联通点对点专线|流量图3",
		"CD-HZ电信点对点专线|流量图4",
	}
	fvs.Add("巡检人", []string{
		opUID,
	})
	fvs.Add("日期", dateNow)
	subFvs.Add("现场网络", "正常")
	subFvs.Add("网络策略", "正常")
	subFvs.Add("现场截图", []string{
		fmt.Sprintf("%s阿里现场/现场截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	subFvs.Add("策略截图", []string{
		fmt.Sprintf("%s阿里现场/策略截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	fvs.Add("巡检地点", []string{
		"成都高朋全部",
	})
	fvs.Add("巡检总结", "正常")

	for _, v := range subTemplateName {
		first := getElem(v, 0)
		subFvs.Add(first, "正常")
		subFvs.Add(getElem(v, 1), components[first])
	}

	fvs.Add("高朋阿里网络", []interface{}{subFvs})

	ret := godingtalk.ProcessinstanceCreateReq{
		AgentID:             metaConfig.AgentID,
		ProcessCode:         cfg.Section("INSPECTION").Key("成都高朋阿里网络巡检").String(),
		OriginatorUserID:    opUID,
		DeptID:              metaConfig.DepartmentID,
		FormComponentValues: fvs,
	}

	return ret
}

func DiDiModel(components map[string][]string, opUID string) godingtalk.ProcessinstanceCreateReq {
	var fvs godingtalk.FormComponentValues
	fvs.Add("巡检人", []string{
		opUID,
	})
	fvs.Add("日期", dateNow)
	var subFvs godingtalk.FormComponentValues
	subTemplateName := []string{
		"成都滴滴联通外网|流量图1",
		"成都滴滴电信外网|流量图2",
		"CD-WJ 联通专线|流量图5",
		"CD-WJ 电信专线|流量图6",
		"WJ-BJ 联通点对点专线|流量图7",
		"WJ-HZ电信点对点专线|流量图8",
	}
	subFvs.Add("现场网络", "正常")
	subFvs.Add("网络策略", "正常")
	subFvs.Add("现场截图", []string{
		fmt.Sprintf("%s滴滴现场/现场截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	subFvs.Add("策略截图", []string{
		fmt.Sprintf("%s滴滴现场/策略截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	fvs.Add("巡检地点", []string{
		"成都高朋全部",
	})
	fvs.Add("巡检总结", "正常")

	for _, v := range subTemplateName {
		first := getElem(v, 0)
		subFvs.Add(first, "正常")
		subFvs.Add(getElem(v, 1), components[first])
	}

	fvs.Add("高朋滴滴网络", []interface{}{subFvs})

	ret := godingtalk.ProcessinstanceCreateReq{
		AgentID:             metaConfig.AgentID,
		ProcessCode:         cfg.Section("INSPECTION").Key("成都高朋滴滴网络巡检").String(),
		OriginatorUserID:    opUID,
		DeptID:              metaConfig.DepartmentID,
		FormComponentValues: fvs,
	}

	return ret
}

func VkSdbModel(components map[string][]string, opUID string) godingtalk.ProcessinstanceCreateReq {
	var fvs godingtalk.FormComponentValues
	fvs.Add("巡检人", []string{
		opUID,
	})
	fvs.Add("日期", dateNow)
	var subFvs godingtalk.FormComponentValues
	subTemplateName := []string{
		"成都VK联通外网|流量图1",
		"成都vk电信外网|流量图2",
	}
	subFvs.Add("现场网络", "正常")
	subFvs.Add("网络策略", "正常")
	subFvs.Add("现场截图", []string{
		fmt.Sprintf("%svk现场/现场截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	subFvs.Add("策略截图", []string{
		fmt.Sprintf("%svk现场/策略截图-%d.jpg", metaConfig.ImgHost, randNum),
	})
	fvs.Add("巡检地点", []string{"成都高朋全部"})
	fvs.Add("巡检总结", "正常")

	for _, v := range subTemplateName {
		first := getElem(v, 0)
		subFvs.Add(first, "正常")
		subFvs.Add(getElem(v, 1), components[first])
	}

	fvs.Add("高朋VK/水滴网络", []interface{}{subFvs})

	ret := godingtalk.ProcessinstanceCreateReq{
		AgentID:             metaConfig.AgentID,
		ProcessCode:         cfg.Section("INSPECTION").Key("成都高朋vk+水滴网络巡检").String(),
		OriginatorUserID:    opUID,
		DeptID:              metaConfig.DepartmentID,
		FormComponentValues: fvs,
	}

	return ret
}
