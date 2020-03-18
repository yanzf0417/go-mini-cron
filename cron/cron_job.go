package cron

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

type CronJob struct {
	LastRunTime time.Time
	*CronExpression
	*Job
}

func (cj *CronJob) NextRunTime() time.Time{
	return cj.CronExpression.ToTime()
}

func ParseCronFile(filepath string) []*CronJob {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("read cron file error: %v\n", err)
		os.Exit(-1)
	}
	return ParseCronData(string(content))
}

func ParseCronData(content string) []*CronJob {
	regexExpression := regexp.MustCompile("\r?\n")
	expressions := regexExpression.Split(content,-1)
	cronJobs := []*CronJob{}
	for _, expression := range expressions {
		cj := ParseCronJob(expression)
		cronJobs = append(cronJobs, cj)
	}
	return cronJobs
}

func ParseCronJob(line string) *CronJob {
	cj := &CronJob{}
	regexLine := regexp.MustCompile(`^(?P<cron>(.*? .*? .*? .*? .*? .*? .*?))\s+(?P<job>(.+))$`)
	match := regexLine.FindStringSubmatch(line)
	if match == nil {
		panic(line)
	}
	result := make(map[string]string)
	groupNames := regexLine.SubexpNames()
	for i, name := range groupNames {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	cj.CronExpression = ParseCronExpression(result["cron"])
	cj.Job = &Job{}
	cj.Script = result["job"]
	return cj
}

