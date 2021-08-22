package util

import (
	"fmt"
	"os"
	"time"
)

func Log(msg string){
	stamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(stamp + ": " + msg)
}

func NoticeAndExit(msg string){
	fmt.Println(msg)
	os.Exit(1)
}