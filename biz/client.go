package biz

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"neolong.me/file_transfer/util"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func conn(cfg *Cfg) *net.TCPConn{
	tcpAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddress)
	if nil != err {
		util.NoticeAndExit("Resolve tcp error: " + err.Error())
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		util.NoticeAndExit("Tcp connect error: " + err.Error())
	}
	return tcpConn
}
func sendAuth(cfg *Cfg, writer *bufio.Writer){
	authData, err := util.RsaEncrypt(util.Int2Byte(int(time.Now().Unix())), cfg.RsaEncKey)
	if nil != err {
		util.NoticeAndExit("auth encrypt error: " + err.Error())
	}
	authLen := len(authData)

	// 发送权限验证数据
	writer.Write(util.Int2Byte(authLen))
	writer.Write(authData)
}

func UploadFile(cfg *Cfg){
	if !strings.HasPrefix(cfg.ToSendFilePath, "/") &&
		!strings.HasPrefix(cfg.ToSendFilePath, "~/") &&
		!strings.HasPrefix(cfg.ToSendFilePath, "./"){
		cfg.ToSendFilePath = "./" + cfg.ToSendFilePath
	}

	data, err := ioutil.ReadFile(cfg.ToSendFilePath)
	if nil != err{
		util.NoticeAndExit("read file error when send file: " + err.Error())
	}

	encData, err := util.AesEncrypt(data, cfg.FileEncryptPwd)
	if nil != err {
		util.NoticeAndExit("encrypt file bytes error: " + err.Error())
	}

	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)

	// 发送权限验证数据
	sendAuth(cfg, writer)
	// 发送功能代码
	writer.Write(util.Int2Byte(TypeSend))
	// 发送文件名数据
	fileNameBytes := []byte(path.Base(cfg.ToSendFilePath))
	nameLen := len(fileNameBytes)
	writer.Write(util.Int2Byte(nameLen))
	writer.Write(fileNameBytes)

	count, _ := writer.Write(encData)
	writer.Flush()
	tcpConn.CloseWrite()
	util.Log("file send finish, send size: " + strconv.Itoa(count))
}

func ListFile(cfg *Cfg){
	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)
	// 发送验证数据
	sendAuth(cfg, writer)
	writer.Write(util.Int2Byte(TypeList))
	writer.Flush()

	// 读取文件列表数据
	reader := bufio.NewReader(tcpConn)
	for{
		data, _, err := reader.ReadLine()
		if nil == err{
			fmt.Println(string(data))
		}
		if nil != err || len(data) <= 0{
			return
		}
	}
}

func DownloadFile(cfg *Cfg){
	if len(cfg.ServerFileName) <= 0 {
		fmt.Println("please specify file name")
		return
	}
	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)
	sendAuth(cfg, writer)
	writer.Write(util.Int2Byte(TypeFetch))
	fileName, err := util.AesEncryptString(cfg.ServerFileName, cfg.FileEncryptPwd)
	if nil != err {
		fmt.Println("file name encrypt fail")
		return
	}
	nameBytes := []byte(fileName)
	writer.Write(util.Int2Byte(len(nameBytes)))
	writer.Write(nameBytes)
	writer.Flush()

	reader := bufio.NewReader(tcpConn)
	funcLen := len(util.Int2Byte(TypeFetch))
	resultBytes := make([]byte, funcLen)

	readed, err := reader.Read(resultBytes)
	if nil != err || len(resultBytes)!=readed {
		return
	}

	resultCode := util.Byte2Int(resultBytes)
	switch resultCode {
	case RESULT_FAIL:
		readFailInfo(reader)
	case RESULT_SUCCESS:
		readFile(reader, cfg)
	}
}

func readFailInfo(reader *bufio.Reader){
	data, _, err := reader.ReadLine()
	if nil != err || len(data) <= 0 {
		return
	}
	fmt.Println("server response: " + string(data))
}

func readFile(reader *bufio.Reader, cfg *Cfg){
	infoLen := len(util.Int2Byte(TypeFetch))
	infoBytes := make([]byte, infoLen)
	readed, err := reader.Read(infoBytes)
	if nil != err || len(infoBytes)!=readed{
		return
	}
	fileLen := util.Byte2Int(infoBytes)

	file, err := os.Create(cfg.ServerFileName)
	if nil != err {
		return
	}
	defer file.Close()
	tempBytes := make([]byte, fileLen)
	readCount, err := reader.Read(tempBytes)
	if nil != err || fileLen!=readCount {
		return
	}

	fileData, err := util.AesDecrypt(tempBytes, cfg.FileEncryptPwd)
	if nil != err {
		return
	}

	file.Write(fileData)
}