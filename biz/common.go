package biz

import (
	"bufio"

	"neolong.me/file_transfer/base"
)

type Transfer interface {
	Send(cfg *base.Cfg, writer *bufio.Writer)
}
