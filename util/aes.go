package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AesEncrypt(data []byte, pwd string)([]byte, error){
	pwdKeyBytes := HashSHA256([]byte(pwd))
	block, err := aes.NewCipher(pwdKeyBytes)
	if nil != err{
		Log("aes encrypt password illegal")
		return nil, err
	}
	blockSize := block.BlockSize()
	data = pkcs5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, pwdKeyBytes[:blockSize])
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	return crypted, nil
}
func AesDecrypt(data []byte, pwd string)([]byte, error){
	pwdKeyBytes := HashSHA256([]byte(pwd))
	block, err := aes.NewCipher(pwdKeyBytes)
	if nil != err {
		Log("aes decrypt password illegal")
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, pwdKeyBytes[:blockSize])
	originData := make([]byte, len(data))
	blockMode.CryptBlocks(originData, data)
	return pkcs5UnPadding(originData), nil
}

func AesEncryptString(data, pwd string)(string, error){
	bytes, err := AesEncrypt([]byte(data), pwd)
	if nil != err{
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func AesDecryptString(data, pwd string)(string, error){
	bytes, error := base64.URLEncoding.DecodeString(data)
	if nil != error {
		return "", nil
	}
	bytes, error = AesDecrypt(bytes, pwd)
	if nil != error{
		return "", nil
	}
	return string(bytes), nil
}


func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}