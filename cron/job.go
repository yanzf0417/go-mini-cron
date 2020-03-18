package cron

import (
	"../util"
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

const  (
	Wait = 0
	Running = 1
	Error = -1
)

type Job struct {
	Pid int
	Script string
	Status int
	LastRunResult int
}

func(job *Job) Run() {
	osterminal := ""
	switch runtime.GOOS {
	case "windows": osterminal = "cmd"
	default:
		osterminal = "bash"
	}
	cmd := exec.Command(osterminal)
	in := bytes.NewBuffer(nil)
	cmd.Stdin = in
	in.WriteString(job.Script + "\n")
	in.WriteString("exit\n")
	err := cmd.Run()
	job.Status = Running
	if err != nil {
		fmt.Println("error", err)
		job.Status = Error
		util.Log("Finish Job: %s. Occurs error %v", job.Script, err)
	} else {
		cmd.Wait()
		job.LastRunResult = cmd.ProcessState.ExitCode()
		util.Log("Finish Job: %s. Exit with code %v", job.Script, job.LastRunResult)
	}
}