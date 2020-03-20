package main

import (
	"./cron"
	"./util"
	"container/heap"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var schedule *cron.Schedule
var jobChan chan *cron.CronJob

func main() {
	if len(os.Args) < 2 {
		fmt.Println("miss cron file")
		os.Exit(-1)
	}
	schedule = &cron.Schedule{}
	jobChan = make(chan *cron.CronJob)
	heap.Init(schedule)
	go StartSchedule()
	go PushCronJob()
	cron_file :=  os.Args[1]
	cronJobs := cron.ParseCronFile(cron_file)
	for _, cronJob := range cronJobs {
		cronJob.MoveNext()
		jobChan <- cronJob
	}
	c := make(chan os.Signal, 0)
	signal.Notify(c, os.Interrupt, os.Kill)
	<- c
}

func PushCronJob() {
	for {
		cronJob, ok := <-jobChan
		if ok {
			heap.Push(schedule, cronJob)
		}
	}
}

func PushCronJob2Chan(cronJobs ...*cron.CronJob) {
	for _, cronJob := range cronJobs {
		jobChan <- cronJob
	}
}

func StartSchedule() {
	timer := time.NewTicker(time.Second*1)
	for {
		ts := (<- timer.C).Unix()
		util.Log("Tick")
		runCronJobs := []*cron.CronJob{}
		for schedule.Len() > 0 && (*schedule)[0].NextRunTime().Unix() == ts {
			cj := heap.Pop(schedule).(*cron.CronJob)
			go cj.Run()
			cj.MoveNext()
			if !cj.IsEnd {
				runCronJobs = append(runCronJobs, cj)
			}
		}
		go PushCronJob2Chan(runCronJobs...)
	}
}
