package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
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

//RSA公钥私钥产生
func GenRsaKey(bits int) error {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	err = pem.Encode(os.Stdout, block)
	if err != nil {
		return err
	}
	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	err = pem.Encode(os.Stdout, block)
	if err != nil {
		return err
	}
	return nil
}