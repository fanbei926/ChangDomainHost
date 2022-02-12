package utils

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/bitly/go-simplejson"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type Info struct {
	AccessID     string
	AccessSecret string
	RecordID     string
	Schedule     string
	DomainName   string
}

var info = &Info{}
var shareMap sync.Map

func init() {
	viper.SetConfigFile("./ali.yml") // 使用 Viper 获取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	info = &Info{
		viper.GetString("AccessId"),
		viper.GetString("AccessSecret"),
		viper.GetString("RecordId"),
		viper.GetString("Schedule"),
		viper.GetString("DomainName"),
	}
}

/**
 * 使用AK&SK初始化账号Client
 * @return Client
 * @throws Exception
 */
func CreateClient() (client *alidns20150109.Client, err error) {
	config := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: &info.AccessID,
		// 您的AccessKey Secret
		AccessKeySecret: &info.AccessSecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dns.aliyuncs.com")
	// client = &alidns20150109.Client{}
	client, err = alidns20150109.NewClient(config)
	return client, err
}

/**
 * 初始化cron
 * @return cron
 * @throws Exception
 */
func CreateCron() (c *cron.Cron, schedule string) {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
	c = cron.New(cron.WithParser(parser))
	return c, info.Schedule
}

/**
 * 执行本地 curl 命令并返回获取的 IP
 * @throws Exception
 */
func ExecCurl() (err error) {
	var out []byte

	c := exec.Command("/bin/bash", "-c", "curl -s -4 ip.sb")
	out, err = c.CombinedOutput()
	if err != nil {
		return err
	}

	// 解析IP
	ip := strings.Trim(string(out), "\n")
	address := net.ParseIP(ip)
	if address == nil {
		ip = ""
	}
	log.Println("Curl result: " + ip)
	shareMap.Store("curlRes", ip)
	return nil
}

/**
 * 执行本地 Lookup 命令并返回对应域名的 IP
 * @throws Exception
 */
func ExecLookup() (err error) {
	var ip string
	ns, err := net.LookupHost(info.DomainName)
	if err != nil {
		return err
	}

	// 解析IP
	if len(ns) == 0 { // Lookup 可能会返回多个地址，需要做下判断
		ip = ""
	} else {
		ip = ns[0]
	}
	address := net.ParseIP(ip)
	if address == nil {
		ip = ""
	}
	log.Println("DNS result: " + ip)
	shareMap.Store("dnsRes", ip)
	return nil
}

/**
 * 对比 curl 跟 lookup 的结果，不一样则更新域名记录
 * @throws Exception
 */
func UpdateDomainRecord() error {
	curlIP, ok := shareMap.Load("curlRes")
	if !ok {
		return errors.New("curlIP is nil")
	}
	dnsIP, ok := shareMap.Load("dnsRes")
	if !ok {
		return errors.New("dnsIP is nil")
	}

	if dnsIP == "" || curlIP == "" {
		return errors.New("neither of them has a value")
	}

	if curlIP != dnsIP {
		client, err := CreateClient()
		if err != nil {
			return err
		}

		tmps := strings.Split(info.DomainName, ".") // 解析出二级域名头
		rr := tmps[0]
		updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
			RecordId: tea.String(info.RecordID),
			RR:       tea.String(rr),
			Type:     tea.String("A"),
			Value:    tea.String(curlIP.(string)),
		}
		// 复制代码运行请自行打印 API 的返回值
		res, err := client.UpdateDomainRecord(updateDomainRecordRequest)
		if err != nil {
			return err
		}
		// fmt.Println(res.Body.GoString())
		js, err := simplejson.NewJson([]byte(res.Body.GoString()))
		if err != nil {
			return err
		}
		fmt.Println(js.Get("RequestId").String())
	}
	return nil
}
