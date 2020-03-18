package util

import (
	"fmt"
	"time"
)

func Log(format string, a ...interface{}) {
	fmt.Printf(fmt.Sprintf("[%s]%s\n", time.Now().Format("2006-01-02 15:04:05"), format), a...)
}
