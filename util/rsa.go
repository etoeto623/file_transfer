package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

/**
 * 使用rsa加密
 */
func RsaEncrypt(data []byte, encryptKey string) ([]byte, error){
	block, _ := pem.Decode([]byte(encryptKey))
	if nil == block {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if nil != err {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

func RsaDecrypt(data []byte, decryptKey string) ([]byte, error){
	block, _ := pem.Decode([]byte(decryptKey))
	if nil == block{
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if nil != err{
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}