package util

import (
	"io/ioutil"
	"os"
)

func ReadFile(path string) ([]byte, error){
	f, err := os.Open(path)
	if nil != err{
		Log("read local file error")
		// 读取文件异常，直接返回
		return nil, err
	}
	defer f.Close() // 延迟关闭文件
	return ioutil.ReadAll(f)
}