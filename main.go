package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/sirupsen/logrus"
	"test.v1/utils"
)

var wg sync.WaitGroup

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	// 开始定时任务
	c, schedule := utils.CreateCron()
	c.AddFunc(schedule, func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Println(r)
			}
		}()

		log.Println("Start")
		wg.Add(2)

		go func() { // 执行 crul
			defer wg.Done()
			err := utils.ExecCurl()
			if err != nil {
				fmt.Println(err)
				return
			}

		}()

		go func() { // 执行 lookup
			defer wg.Done()
			err := utils.ExecLookup()
			if err != nil {
				fmt.Println(err)
				return
			}
		}()

		err := utils.UpdateDomainRecord() // 进行对比更新
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Wait()
	})
	c.Start()
	defer c.Stop()
	select {}
}
