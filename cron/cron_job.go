package cron

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
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
	regexLine := regexp.MustCompile(`^(?P<cron>((\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)(\s+[0-9\-\*,]+)?))\s+(?P<job>(.+))$`)
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
	cj.Desc = result["job"]
	cj.Action = MakeAction(result["job"])
	return cj
}

func MakeAction(script string) func(c chan JobResult) {
	return func(c chan JobResult) {
		osterminal := ""
		switch runtime.GOOS {
		case "windows":
			osterminal = "cmd"
		default:
			osterminal = "bash"
		}
		cmd := exec.Command(osterminal)
		in := bytes.NewBuffer(nil)
		cmd.Stdin = in
		in.WriteString(script + "\n")
		in.WriteString("exit\n")
		err := cmd.Run()
		if err != nil {
			c <- JobResult{Code:-1000, Msg:fmt.Sprint(err)}
		} else {
			cmd.Wait()
			c <- JobResult{cmd.ProcessState.ExitCode(), cmd.ProcessState.String()}
		}
	}
}
