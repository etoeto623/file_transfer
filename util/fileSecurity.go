package util

import (
	"os"
	"neolong.me/neotools/cipher"
)

/*
 * 负责文件的安全问题，加密和解密
 */


func EncryptFile(path, password string) ([]byte, error) {
	data, err := ReadFile(path)
	if nil != err || nil == data{
		Log("read file error: " + path)
		os.Exit(1)
	}

	return cipher.RsaEncrypt(data, password)
}

func DecryptBytes(encrypted []byte, password string)([]byte, error) {
	return cipher.RsaDecrypt(encrypted, password)
}