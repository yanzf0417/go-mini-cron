package cron

import (
	"../util"
)

const  (
	Wait = 0
	Running = 1
	Error = -1
)

type Job struct {
	Pid int
	Action func(c chan JobResult)
	Desc string
	Status int
	LastRunResult int
}

type JobResult struct {
	Code int
	Msg string
}

func(job *Job) Run() {
	job.Status = Running
	util.Log("Start Job: %s", job.Desc)
	c := make(chan JobResult)
	go job.Action(c)
	result := <- c
	job.Status = Wait
	util.Log("Finish Job: %s. Code: %d, Msg: %s.", job.Desc, result.Code, result.Msg)
}