package main

import (
	"flag"
	"neolong.me/neotools/cipher"
	"os"

	"neolong.me/file_transfer/base"
	"neolong.me/file_transfer/biz"
	"neolong.me/file_transfer/util"
)

func send(cfg *base.Cfg) {
	biz.UploadFile(cfg)
}
func serve(cfg *base.Cfg) {
	biz.DoServe(cfg)
}
func fetch(cfg *base.Cfg) {
	biz.DownloadFile(cfg)
}

func prepareCall(paramFunc, worker func(cfg *base.Cfg)) {
	cfg := base.GetCfg()
	paramFunc(&cfg)
	worker(&cfg)
}

func main() {
	//configPath := flag.String("c", "", "配置文件路径")
	filePasswd := flag.String("pwd", "", "传输文件的加密密钥")
	serverPort := flag.Int("port", 8090, "服务器的端口")
	//serverAddress := flag.String("s", "", "服务器的地址，包含端口号")
	//fileName := flag.String("n", "", "传输文件的文件名")
	flag.Parse()

	//fmt.Printf("%s%d%s%s\n", filePasswd, serverPort, serverAddress, fileName)
	if len(os.Args) < 2 {
		util.NoticeAndExit("parameter error")
	}
	switch os.Args[1] {
	case "send":
		prepareCall(func(cfg *base.Cfg) {
			if len(os.Args) < 3 {
				util.NoticeAndExit("param error")
			}
			if len(*filePasswd) > 0 {
				cfg.FileEncryptPwd = *filePasswd
			}
			localFilePath := os.Args[len(os.Args)-1]
			cfg.ToSendFilePath = localFilePath
		}, send)
	case "list":
		prepareCall(func(cfg *base.Cfg) {
		}, func(cfg *base.Cfg) {
			biz.ListFile(cfg)
		})
	case "serve":
		prepareCall(func(cfg *base.Cfg) {
			if *serverPort < 100 || *serverPort >= 65535 {
				util.NoticeAndExit("port is illegal, which should be greater than 100 and less than 65535")
			}
			cfg.Port = *serverPort
		}, serve)
	case "fetch":
		prepareCall(func(cfg *base.Cfg) {
			if len(os.Args) < 3 {
				util.NoticeAndExit("param error")
			}
			serverFileName := os.Args[len(os.Args)-1]
			cfg.ServerFileName = serverFileName
		}, fetch)
	case "genkey":
		prepareCall(func(cfg *base.Cfg) {
		}, func(cfg *base.Cfg) {
			cipher.GenRsaKey(1024)
		})
	case "delete":
		prepareCall(func(cfg *base.Cfg) {
			if len(os.Args) < 3 {
				util.NoticeAndExit("param error")
			}
			if len(*filePasswd) > 0 {
				cfg.FileEncryptPwd = *filePasswd
			}
			localFilePath := os.Args[len(os.Args)-1]
			cfg.ToSendFilePath = localFilePath
		}, biz.DeleteFile)
	default:
		util.NoticeAndExit("illegal mode")
	}
}
