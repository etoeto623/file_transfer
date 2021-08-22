package util

import (
	"encoding/json"
	"os"
)

func Struct2JSON(obj interface{})string{
	jsonStr, e := json.Marshal(obj)
	if e!=nil{
		Log("json format error")
		os.Exit(1)
	}
	return string(jsonStr)
}