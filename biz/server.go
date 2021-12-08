package biz

import (
	"bufio"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-basic/uuid"
	"neolong.me/file_transfer/util"
)

func DoServe(cfg *Cfg) {
	util.Log("start to bootstrap server")
	tcpAddr, e := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(cfg.Port))
	if e != nil {
		util.NoticeAndExit("server start error:" + e.Error())
	}
	tcpListener, e := net.ListenTCP("tcp", tcpAddr)
	if e != nil {
		util.NoticeAndExit("server listen error: " + e.Error())
	}
	defer tcpListener.Close()

	// 开始进行连接处理
	for true {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			util.Log("server accept connect error: " + err.Error())
			continue
		}
		go handleTCP(conn, cfg)
	}
}

func handleTCP(conn *net.TCPConn, cfg *Cfg) {
	uuidStr := uuid.New()
	ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	util.Log(uuidStr + " got request from " + ip)
	defer conn.Close()
	reader := bufio.NewReader(conn)

	intLen := len(util.Int2Byte(TypeSend))
	funcBytes := make([]byte, intLen)
	// 读取鉴权信息
	readed, err := reader.Read(funcBytes)
	if nil != err || readed != intLen {
		util.Log(uuidStr + " auth length illegal")
		return
	}
	funcLen := util.Byte2Int(funcBytes)
	if funcLen <= 0 { // illegal request
		return
	}
	authBytes := make([]byte, funcLen)
	readed, err = reader.Read(authBytes)
	if nil != err || len(authBytes) != readed {
		util.Log(uuidStr + " auth data illegal")
		return
	}
	authData, err := util.RsaDecrypt(authBytes, cfg.RsaDecKey)
	if nil != err {
		util.Log(uuidStr + " auth data verify fail")
		return
	}
	clientTs := util.Byte2Int(authData)
	now := int(time.Now().Unix())
	if math.Abs(float64(now-clientTs)) > 5 {
		// 5s内有效
		util.Log(uuidStr + " auth expired")
		return
	}

	readed, err = reader.Read(funcBytes)
	if nil != err || len(funcBytes) != readed {
		return
	}

	// 读取客户端的请求功能
	funcCode := util.Byte2Int(funcBytes)
	switch funcCode {
	case TypeSend:
		receiveFile(reader, &funcBytes, cfg)
	case TypeList:
		listFile(conn, &funcBytes, cfg)
	case TypeFetch:
		downloadFile(reader, conn, &funcBytes, cfg)
	default:
		util.Log(uuidStr + " function code illegal")
	}
}

/** 接收文件，分块进行 */
func receiveFile(reader io.Reader, funcBytes *[]byte, cfg *Cfg) {
	// 读取文件名
	readed, err := reader.Read(*funcBytes)
	if nil != err || len(*funcBytes) != readed {
		return
	}
	fileNameLen := util.Byte2Int(*funcBytes)
	fileNameBytes := make([]byte, fileNameLen)
	readed, err = reader.Read(fileNameBytes)
	if nil != err || len(fileNameBytes) != readed {
		return
	}
	fileName := string(fileNameBytes)
	file, err := os.Create(cfg.Warehouse + fileName)
	if nil != err {
		return
	}
	defer file.Close()

	// 读取文件流
	intLen := len(util.Int2Byte(TypeSend))
	intBuf := make([]byte, intLen)

	receiveFileSize := 0
	for {
		_, err = reader.Read(intBuf)
		if nil != err {
			util.Log("function code read error: " + err.Error())
			return
		}
		funcCode := util.Byte2Int(intBuf)

		switch funcCode {
		case TypeFinish:
			util.Log("file receive finish, size: " + strconv.Itoa(receiveFileSize))
			return
		case TypeFileBuck:
			_, err = reader.Read(intBuf)
			if nil != err {
				util.Log("file bucket size read error: " + err.Error())
				return
			}
			bucketSize := util.Byte2Int(intBuf)
			bucketBuf := make([]byte, bucketSize)
			n, err := reader.Read(bucketBuf)
			if nil != err {
				util.Log("file bucket read error: " + err.Error())
				return
			}
			if n < bucketSize {
				util.Log("file bucket read uncomplete")
				return
			}
			decrypted, err := util.AesDecrypt(bucketBuf, cfg.FileEncryptPwd)
			if nil != err {
				util.Log("file bucket decrypt error: " + err.Error())
				return
			}

			file.Write(decrypted)
			receiveFileSize += len(decrypted)
		}
	}
}

func downloadFile(reader io.Reader, conn *net.TCPConn, funcBytes *[]byte, cfg *Cfg) {
	writer := bufio.NewWriter(conn)
	readed, err := reader.Read(*funcBytes)
	if nil != err || len(*funcBytes) != readed {
		writer.Write(util.Int2Byte(RESULT_FAIL))
		writer.Write(util.Int2Byte(len(FAIL_COMMON)))
		writer.Write([]byte(FAIL_COMMON))
		writer.Flush()
		return
	}
	fileNameLen := util.Byte2Int(*funcBytes)
	fileNameBytes := make([]byte, fileNameLen)
	readed, err = reader.Read(fileNameBytes)
	if nil != err || fileNameLen != readed {
		writer.Write(util.Int2Byte(RESULT_FAIL))
		writer.Write(util.Int2Byte(len(FAIL_COMMON)))
		writer.Write([]byte(FAIL_COMMON))
		writer.Flush()
		return
	}
	fileName := string(fileNameBytes)
	fileBytes, err := ioutil.ReadFile(cfg.Warehouse + fileName)
	if nil != err {
		errMsg := err.Error()
		writer.Write(util.Int2Byte(RESULT_FAIL))
		writer.Write(util.Int2Byte(len(errMsg)))
		writer.Write([]byte(errMsg))
		writer.Flush()
		return
	}

	writer.Write(util.Int2Byte(RESULT_SUCCESS))
	writer.Write(util.Int2Byte(len(fileBytes)))
	writer.Write(fileBytes)
	writer.Flush()
}

// 列出文件列表
func listFile(conn *net.TCPConn, funcBytes *[]byte, cfg *Cfg) {
	writer := bufio.NewWriter(conn)
	br := []byte("\n")
	filepath.Walk(cfg.Warehouse, func(fPath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fileName := path.Base(fPath)
		//fileName, err = util.AesDecryptString(fileName, cfg.FileEncryptPwd)
		//if nil != err {
		//	return nil
		//}
		writer.Write([]byte(fileName))
		writer.Write(br)
		return nil
	})
	writer.Flush()
}
