package util

import (
	"fmt"
	"os"
	"time"

	"neolong.me/file_transfer/base"
)

type LogLevel int

// 日志级别
const (
	DEBUG int = iota
	INFO
	WARNING
	ERROR
	FATAL
	OFF
)

func log(msg string, level int) {
	realLevel := base.GetCfg().LogLevel
	if realLevel > level {
		return
	}
	stamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := "OFF"
	switch level {
	case DEBUG:
		levelName = "DEBUG"
		break
	case INFO:
		levelName = "INFO"
		break
	case WARNING:
		levelName = "WARN"
		break
	case ERROR:
		levelName = "ERROR"
		break
	case FATAL:
		levelName = "FATAL"
		break
	}
	fmt.Println(stamp + " [" + levelName + "]: " + msg)
}

func Log(msg string) {
	log(msg, INFO)
}
func LogDebug(msg string) {
	log(msg, DEBUG)
}
func LogInfo(msg string) {
	log(msg, INFO)
}
func LogWarning(msg string) {
	log(msg, WARNING)
}
func LogError(msg string) {
	log(msg, ERROR)
}
func LogFatal(msg string) {
	log(msg, FATAL)
}

func NoticeAndExit(msg string) {
	LogFatal(msg)
	os.Exit(1)
}
