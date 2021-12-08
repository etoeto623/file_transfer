package biz

import "bufio"

type Transfer interface {
	Send(cfg *Cfg, writer *bufio.Writer)
}
