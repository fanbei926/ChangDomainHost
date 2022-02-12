package main

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"test.v1/utils"
)

var wg sync.WaitGroup
var logs = logrus.New()

func main() {
	logs.Out = os.Stdout
	logs.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	// 开始定时任务
	c, schedule := utils.CreateCron()
	c.AddFunc(schedule, func() {
		defer func() {
			if r := recover(); r != nil {
				logs.Errorln(r, " -->")
			}
		}()

		logs.Println("Start cron schedule...")
		wg.Add(2)

		go func() { // 执行 curl
			defer wg.Done()
			err := utils.ExecCurl()
			if err != nil {
				logs.WithFields(logrus.Fields{
					"func": "curl",
				}).Error(err, " -->")
				return
			}
		}()

		go func() { // 执行 lookup
			defer wg.Done()
			err := utils.ExecLookup()
			if err != nil {
				logs.WithFields(logrus.Fields{
					"func": "lookup",
				}).Error(err, " -->")
				return
			}
		}()

		wg.Wait()
		err := utils.UpdateDomainRecord() // 进行对比更新
		if err != nil {
			logs.WithFields(logrus.Fields{
				"func": "updateDomain",
			}).Error(err, " -->")
			return
		}
	})
	c.Start()
	defer c.Stop()
	select {}
}
