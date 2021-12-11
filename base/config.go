package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"sync"
)

const TypeSend int = 233
const TypeList int = 666
const TypeFetch int = 555

// 文件块传输
const TypeFileBuck = 623

// 结束标志
const TypeFinish = 600

const RESULT_SUCCESS int = 88
const RESULT_FAIL int = 55

const FAIL_COMMON string = "FAIL"

func homeDir() string {
	u, _ := user.Current()
	return u.HomeDir
}

var cfgPath = homeDir() + "/.ftrc"

// 配置信息
type BaseCfg struct {
	ServerAddress string // 服务器地址
	Warehouse     string // 服务器文件保存文件夹
	RsaEncKey     string // rsa的密钥
	RsaDecKey     string
}
type Cfg struct {
	*BaseCfg
	Port           int    // 服务器的端口
	ToSendFilePath string // 待上传的文件的路径
	ServerFileName string // 需要下载的服务器上的文件名
	FileEncryptPwd string // 文件加密的密钥
	BuckSize       int    // 文件传输块大小
	LogLevel       int    // 日志级别
}

func (cfg BaseCfg) toString() string {
	data, e := json.Marshal(cfg)
	if nil != e {
		return ""
	}
	return string(data)
}

var config Cfg
var inited = false
var lock sync.Mutex

func GetCfg() Cfg {
	lock.Lock()
	defer lock.Unlock()
	if inited {
		return config
	}
	file, err := os.Open(cfgPath)
	if nil != err {
		fmt.Println("config file open error: " + err.Error())
		os.Exit(1)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if nil != err {
		fmt.Println("config file read error")
		os.Exit(1)
	}
	e := json.Unmarshal(data, &config)
	if nil != e {
		fmt.Println("config file parse error")
		os.Exit(1)
	}
	inited = true
	return config
}

func getCmdCfg(args []string, typePort string) (string, bool) {
	if nil == args || len(args) == 0 {
		return "", false
	}
	prefix := "-" + typePort + "="
	for i := range args {
		cfg := args[i]
		if strings.HasPrefix(cfg, prefix) {
			return strings.Replace(cfg, prefix, "", 1), true
		}
	}
	return "", false
}
