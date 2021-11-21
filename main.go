package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/bitly/go-simplejson"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

var wg sync.WaitGroup
var shareMap sync.Map

type info struct {
	AccessId     string `yaml:"AccessId"`
	AccessSecret string `yaml:"AccessSecret"`
	RecordId     string `yaml:"RecordId"`
}

func parseYaml(i *info) error {
	f, err := os.Open("./ali.yml")
	defer f.Close()
	if err != nil {
		return err
	}

	by, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(by, &i)
	if err != nil {
		return err
	}
	return nil
}

/**
 * 使用AK&SK初始化账号Client
 * @param accessKeyId
 * @param accessKeySecret
 * @return Client
 * @throws Exception
 */
func CreateClient() (_result *alidns20150109.Client, _err error) {
	var i info
	err := parseYaml(&i)
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: &i.AccessId,
		// 您的AccessKey Secret
		AccessKeySecret: &i.AccessSecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dns.aliyuncs.com")
	_result = &alidns20150109.Client{}
	_result, _err = alidns20150109.NewClient(config)
	return _result, _err
}

func main() {
	myParser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
	c := cron.New(cron.WithParser(myParser))
	c.AddFunc("0 2,14 * * * ", func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
			}
		}()

		log.Println("Start")
		wg.Add(2)

		go func() {
			defer wg.Done()

			c := exec.Command("/bin/bash", "-c", "curl -s -4 ip.sb")
			var b []byte
			b, err := c.CombinedOutput()
			if err != nil {
				log.Println(err)
			}

			// 解析IP
			ip := strings.Trim(string(b), "\n")
			address := net.ParseIP(ip)
			if address == nil {
				ip = ""
			}
			log.Println("Curl result: " + ip)
			shareMap.Store("curlRes", ip)
		}()

		go func() {
			defer wg.Done()

			ns, err := net.LookupHost("nextcloud.fanfan926.icu")
			if err != nil {
				fmt.Println(err)
			}

			// 解析IP
			var ip string
			if len(ns) == 0 {
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
		}()

		wg.Wait()

		curlIP, ok := shareMap.Load("curlRes")
		if !ok {
			log.Println("curlIP is nil")
		}
		dnsIP, ok := shareMap.Load("dnsRes")
		if !ok {
			log.Println("dnsIP is nil")
		}

		if curlIP == dnsIP {
			client, err := CreateClient()
			if err != nil {
				fmt.Println(err)
			}

			var i info
			err = parseYaml(&i)
			if err != nil {
				fmt.Println(err)
			}

			updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
				RecordId: tea.String(i.RecordId),
				RR:       tea.String("nextcloud"),
				Type:     tea.String("A"),
				Value:    tea.String(curlIP.(string)),
			}
			// 复制代码运行请自行打印 API 的返回值
			res, err := client.UpdateDomainRecord(updateDomainRecordRequest)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(res.Body.GoString())
			js, _ := simplejson.NewJson([]byte(res.Body.GoString()))
			fmt.Println(js.Get("RequestId").String())
		}

	})
	c.Start()
	defer c.Stop()
	select {}
}
