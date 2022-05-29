package biz

import (
	"bufio"
	"io"
	"io/fs"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-basic/uuid"
	"neolong.me/file_transfer/base"
	"neolong.me/file_transfer/util"
	"neolong.me/neotools/cipher"
)

func DoServe(cfg *base.Cfg) {
	util.Log("start to bootstrap server at port " + strconv.Itoa(cfg.Port))
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

func handleTCP(conn *net.TCPConn, cfg *base.Cfg) {
	uuidStr := uuid.New() // requestId
	ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	util.Log(uuidStr + " got request from " + ip)
	defer conn.Close()
	reader := bufio.NewReader(conn)

	intLen := len(util.Int2Byte(base.TypeSend))
	funcBytes := make([]byte, intLen)
	// 读取鉴权信息
	readed, err := io.ReadFull(reader, funcBytes)
	if nil != err || readed != intLen {
		util.Log(uuidStr + " auth length illegal")
		return
	}
	funcLen := util.Byte2Int(funcBytes)
	if funcLen <= 0 || funcLen > 2000 { // illegal request
		util.Log(uuidStr + " illegal auth length: " + strconv.Itoa(funcLen))
		return
	}

	authBytes := make([]byte, funcLen)
	readed, err = reader.Read(authBytes)
	if nil != err || len(authBytes) != readed {
		util.Log(uuidStr + " auth data illegal")
		return
	}
	authData, err := cipher.RsaDecrypt(authBytes, cfg.RsaDecKey)
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

	readed, err = io.ReadFull(reader, funcBytes)
	if nil != err || len(funcBytes) != readed {
		return
	}

	// 读取客户端的请求功能
	funcCode := util.Byte2Int(funcBytes)
	switch funcCode {
	case base.TypeSend:
		receiveFile(reader, &funcBytes, cfg)
	case base.TypeList:
		listFile(conn, &funcBytes, cfg)
	case base.TypeFetch:
		downloadFile(reader, conn, &funcBytes, cfg)
	case base.TypeDelete:
		commandDeleteFile(reader, conn, &funcBytes, cfg)
	default:
		util.Log(uuidStr + " function code illegal")
	}
}

/** 接收文件，分块进行 */
func receiveFile(reader io.Reader, funcBytes *[]byte, cfg *base.Cfg) {
	// 读取文件名
	readed, err := io.ReadFull(reader, *funcBytes)
	if nil != err || len(*funcBytes) != readed {
		return
	}
	fileNameLen := util.Byte2Int(*funcBytes)
	fileNameBytes := make([]byte, fileNameLen)
	readed, err = io.ReadFull(reader,fileNameBytes)
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
	intLen := len(util.Int2Byte(base.TypeSend))
	intBuf := make([]byte, intLen)

	for {
		_, err = io.ReadFull(reader,intBuf)
		if nil != err {
			util.Log("function code read error: " + err.Error())
			return
		}
		funcCode := util.Byte2Int(intBuf)

		switch funcCode {
		case base.TypeFinish:
			util.Log("file receive finish")
			return
		case base.TypeFileBuck:
			_, err = io.ReadFull(reader,intBuf) // 读取文件分块的大小
			if nil != err {
				util.Log("file bucket size read error: " + err.Error())
				deleteFile(file)
				return
			}

			// 分块文件大小
			bucketSize := util.Byte2Int(intBuf)
			bucketBuf := make([]byte, bucketSize)
			n, err := io.ReadFull(reader, bucketBuf) // 这里要使用ReadFull，否则可能读取不完全
			if nil != err {
				if err == io.EOF {
					return
				}
				util.Log("file bucket read error 2: " + err.Error())
				deleteFile(file)
				return
			}
			util.Log(" --------- read bucket: " + strconv.Itoa(n) + ", " + strconv.Itoa(bucketSize))
			if n < bucketSize {
				util.Log("file bucket read uncomplete")
				deleteFile(file)
				return
			}

			file.Write(intBuf) // 记录分块大小
			file.Write(bucketBuf)
		}
	}
}

func deleteFile(f *os.File) {
	if nil != f {
		os.Remove(f.Name())
	}
}

func commandDeleteFile(reader io.Reader, conn *net.TCPConn, funcBytes *[]byte, cfg *base.Cfg) {
	readed, err := io.ReadFull(reader, *funcBytes)
	if nil != err || len(*funcBytes) != readed {
		return
	}
	fileNameLen := util.Byte2Int(*funcBytes)
	fileNameBytes := make([]byte, fileNameLen)
	readed, err = io.ReadFull(reader, fileNameBytes)
	if nil != err || len(fileNameBytes) != readed {
		return
	}
	fileName := string(fileNameBytes)

	writer := bufio.NewWriter(conn)
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		util.LogInfo("command delete file has illegal character: " + fileName)
		writeMsg(writer, base.RESULT_FAIL, "illegal file path")
		return
	}

	filePath := cfg.Warehouse + fileName
	_, err = os.Stat(filePath)
	if nil != err {
		util.LogInfo("command delete file check error: " + err.Error())
		writeMsg(writer, base.RESULT_FAIL, err.Error())
		return
	}

	os.Remove(filePath)
	writeMsg(writer, base.RESULT_SUCCESS, "delete success")
}

// 下载文件，需要进行分块下载
func downloadFile(reader io.Reader, conn *net.TCPConn, funcBytes *[]byte, cfg *base.Cfg) {
	writer := bufio.NewWriter(conn)
	readed, err := io.ReadFull(reader, *funcBytes)

	// 发生错误
	if nil != err || len(*funcBytes) != readed {
		writeMsg(writer, base.RESULT_FAIL, base.FAIL_COMMON)
		return
	}

	// 读取需要传输的文件名
	fileNameLen := util.Byte2Int(*funcBytes)
	fileNameBytes := make([]byte, fileNameLen)
	readed, err = io.ReadFull(reader, fileNameBytes)
	if nil != err || fileNameLen != readed { // 读取文件失败
		writeMsg(writer, base.RESULT_FAIL, base.FAIL_COMMON)
		return
	}

	intLen := len(util.Int2Byte(base.TypeSend))
	fileName := string(fileNameBytes)

	// 开始分块读取文件
	file, err := os.Open(cfg.Warehouse + fileName)
	if nil != err {
		util.Log("target file [" + fileName + "] read error: " + err.Error())
		writeMsg(writer, base.RESULT_FAIL, err.Error())
		return
	}
	fs, _ := file.Stat()
	fileSize := fs.Size()

	intBytes := make([]byte, intLen)
	writer.Write(util.Int2Byte(base.TypeSend)) // 文件开始传输的标志
	writer.Write(util.Int2Byte(int(fileSize))) // 文件大小信息
	writer.Flush()
	sendTotal := 0
	for {
		n, err := io.ReadFull(file, intBytes)
		if nil != err {
			if err != io.EOF {
				util.Log("send file read bucket size error: " + err.Error())
				// writeMsg(writer, RESULT_FAIL, err.Error())
				return
			}
			if err == io.EOF {
				break
			}
		}
		if n <= 0 {
			break
		}

		bucketSize := util.Byte2Int(intBytes)
		bucketBuf := make([]byte, bucketSize)
		n, err = io.ReadFull(file, bucketBuf)
		if nil != err && err != io.EOF{
			util.Log("send file read bucket data error: " + err.Error())
			// writeMsg(writer, RESULT_FAIL, err.Error())
			return
		}
		if n < bucketSize {
			util.Log("send file read bucket not complete, expect: " + strconv.Itoa(bucketSize) + ", actual: " + strconv.Itoa(n))
			// writeMsg(writer, RESULT_FAIL, "file bucket read uncomplete")
			return
		}
		writer.Write(util.Int2Byte(base.TypeSend)) // 文件开始传输的标志
		writer.Write(intBytes)
		writer.Write(bucketBuf)
		writer.Flush()
		sendTotal += n
		util.Log("send file bucket: " + strconv.Itoa(sendTotal) + "/" + strconv.Itoa(int(fileSize)))
	}

	// 发送结束信号
	writer.Write(util.Int2Byte(base.TypeFinish))
	writer.Flush()
}

func writeMsg(writer *bufio.Writer, code int, msg string) {
	writer.Write(util.Int2Byte(code))
	writer.Write(util.Int2Byte(len(msg)))
	writer.Write([]byte(msg))
	writer.Flush()
}

// 列出文件列表
func listFile(conn *net.TCPConn, funcBytes *[]byte, cfg *base.Cfg) {
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
