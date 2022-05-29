package biz

import (
	"bufio"
	"fmt"
	"io"
	"neolong.me/neotools/cipher"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"neolong.me/file_transfer/base"
	"neolong.me/file_transfer/util"
)

func conn(cfg *base.Cfg) *net.TCPConn {
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
func sendAuth(cfg *base.Cfg, writer *bufio.Writer) {
	authData, err := cipher.RsaEncrypt(util.Int2Byte(int(time.Now().Unix())), cfg.RsaEncKey)
	if nil != err {
		util.NoticeAndExit("auth encrypt error: " + err.Error())
	}
	authLen := len(authData)

	// 发送权限验证数据
	writer.Write(util.Int2Byte(authLen))
	writer.Write(authData)
	writer.Flush()
}

func DeleteFile(cfg *base.Cfg) {
	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)

	// 发送权限验证数据
	sendAuth(cfg, writer)
	// 发送功能代码
	writer.Write(util.Int2Byte(base.TypeDelete))
	writer.Flush()
	// 发送文件名数据
	encFileName, err := util.AesEncryptString(path.Base(cfg.ToSendFilePath), cfg.FileEncryptPwd)
	if nil != err {
		util.NoticeAndExit("file name encrypt error: " + err.Error())
	}
	fileNameBytes := []byte(encFileName)
	nameLen := len(fileNameBytes)
	writer.Write(util.Int2Byte(nameLen))
	writer.Write(fileNameBytes)
	writer.Flush()

	// 接收信息
	reader := bufio.NewReader(tcpConn)
	intLen := len(util.Int2Byte(base.TypeDelete))
	intBytes := make([]byte, intLen)
	io.ReadFull(reader, intBytes)
	io.ReadFull(reader, intBytes)
	msgBytes := make([]byte, util.Byte2Int(intBytes))
	n, _ := io.ReadFull(reader, msgBytes)
	if n > 0 {
		util.LogInfo(string(msgBytes))
	}
}

func UploadFile(cfg *base.Cfg) {
	if !strings.HasPrefix(cfg.ToSendFilePath, "/") &&
		!strings.HasPrefix(cfg.ToSendFilePath, "~/") &&
		!strings.HasPrefix(cfg.ToSendFilePath, "./") {
		// 默认发送当前路径下的文件
		cfg.ToSendFilePath = "./" + cfg.ToSendFilePath
	}

	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)

	// 发送权限验证数据
	sendAuth(cfg, writer)
	// 发送功能代码
	writer.Write(util.Int2Byte(base.TypeSend))
	writer.Flush()
	// 发送文件名数据
	encFileName, err := util.AesEncryptString(path.Base(cfg.ToSendFilePath), cfg.FileEncryptPwd)
	if nil != err {
		util.NoticeAndExit("file name encrypt error: " + err.Error())
	}
	fileNameBytes := []byte(encFileName)
	nameLen := len(fileNameBytes)
	writer.Write(util.Int2Byte(nameLen))
	writer.Write(fileNameBytes)
	writer.Flush()

	// TODO 这里要进行文件分块读取
	f, err := os.Open(cfg.ToSendFilePath)
	if nil != err {
		util.NoticeAndExit("file read error: " + err.Error())
	}
	defer f.Close()
	buf := make([]byte, cfg.BuckSize)
	writeCount := 0
	for {
		n, err := io.ReadFull(f, buf)
		if nil != err && err != io.EOF {
			util.Log("file bucket read error: " + err.Error())
			sendFinishSignal(writer)
			return
		}
		if n <= 0 || (nil != err && err == io.EOF) {
			sendFinishSignal(writer)
			break
		}

		enced, err := util.AesEncrypt(buf[0:n], cfg.FileEncryptPwd)
		if nil != err {
			util.Log("file bucket encrypt error: " + err.Error())
			sendFinishSignal(writer)
			return
		}
		bucketLen := len(enced)
		writer.Write(util.Int2Byte(base.TypeFileBuck))
		writer.Write(util.Int2Byte(bucketLen))
		writer.Write(enced)
		writer.Flush()
		writeCount += n
	}

	writer.Flush()
	tcpConn.CloseWrite()
	util.Log("file send finish, send size: " + strconv.Itoa(writeCount))
}

func ListFile(cfg *base.Cfg) {
	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)
	// 发送验证数据
	sendAuth(cfg, writer)
	writer.Write(util.Int2Byte(base.TypeList))
	writer.Flush()

	// 读取文件列表数据
	reader := bufio.NewReader(tcpConn)
	for {
		data, _, err := reader.ReadLine()
		if nil == err {
			fileName, e := util.AesDecryptString(string(data), cfg.FileEncryptPwd)
			if nil != e {
				util.NoticeAndExit("file name decrypt error when list: " + e.Error())
			}
			fmt.Println(fileName)
		}
		if nil != err || len(data) <= 0 {
			return
		}
	}
}

func sendFinishSignal(writer *bufio.Writer) {
	writer.Write(util.Int2Byte(base.TypeFinish))
	writer.Flush()
}

// 从服务器下载文件
func DownloadFile(cfg *base.Cfg) {
	if len(cfg.ServerFileName) <= 0 {
		util.Log("please specify file name")
		return
	}
	tcpConn := conn(cfg)
	defer tcpConn.Close()

	writer := bufio.NewWriter(tcpConn)
	sendAuth(cfg, writer)
	writer.Write(util.Int2Byte(base.TypeFetch))
	writer.Flush()
	fileName, err := util.AesEncryptString(cfg.ServerFileName, cfg.FileEncryptPwd)
	if nil != err {
		util.Log("file name encrypt fail")
		return
	}

	// 发送文件名称
	nameBytes := []byte(fileName)
	writer.Write(util.Int2Byte(len(nameBytes)))
	writer.Write(nameBytes)
	writer.Flush()

	// 开始准备接收数据
	reader := bufio.NewReader(tcpConn)
	funcLen := len(util.Int2Byte(base.TypeFetch))
	resultBytes := make([]byte, funcLen)

	// TODO 这里要分段接受数据
	readed, err := io.ReadFull(reader, resultBytes)
	if nil != err || len(resultBytes) != readed {
		return
	}

	resultCode := util.Byte2Int(resultBytes)
	switch resultCode {
	case base.RESULT_FAIL:
		readFailInfo(reader)
	case base.RESULT_SUCCESS:
	case base.TypeSend:
		readFile(reader, cfg)
	}
}

func readFailInfo(reader *bufio.Reader) {
	data, _, err := reader.ReadLine()
	if nil != err || len(data) <= 0 {
		return
	}
	fmt.Println("server response: " + string(data))
}

// 开始读取文件
func readFile(reader *bufio.Reader, cfg *base.Cfg) {
	intLen := len(util.Int2Byte(base.TypeFetch))
	intBytes := make([]byte, intLen)

	file, err := os.Create(cfg.ServerFileName)
	if nil != err {
		return
	}
	defer file.Close()

	for {
		n, err := io.ReadFull(reader, intBytes)
		if nil != err {
			if err == io.EOF {
				break
			}
			util.Log("int bytes read error: " + err.Error())
			removeFile(file)
			return
		}
		if n < intLen {
			if n <= 0 {
				return
			}
			util.Log("data receive uncomplete")
			removeFile(file)
			return
		}

		typeInfo := util.Byte2Int(intBytes)
		if typeInfo == base.TypeFinish {
			// 文件接收完成
			break
		}
		if typeInfo != base.TypeSend {
			util.Log("Illegal data transfer format")
			removeFile(file)
			return
		}

		n, err = io.ReadFull(reader, intBytes)
		if n < intLen {
			if n <= 0 {
				return
			}
			util.Log("data receive uncomplete")
			removeFile(file)
			return
		}

		bucketSize := util.Byte2Int(intBytes)
		bucketBytes := make([]byte, bucketSize)
		n, err = io.ReadFull(reader, bucketBytes)
		if nil != err {
			util.Log("file bucket read error: " + err.Error())
			removeFile(file)
			return
		}
		if n < bucketSize {
			util.Log("file bucket read uncomplete")
			removeFile(file)
			return
		}

		decryptBytes, err := util.AesDecrypt(bucketBytes, cfg.FileEncryptPwd)
		if nil != err {
			util.Log("file bucket decrypt error: " + err.Error())
			removeFile(file)
			return
		}
		file.Write(decryptBytes)
	}
}

func removeFile(file *os.File) {
	if nil != file {
		os.Remove(file.Name())
	}
}
