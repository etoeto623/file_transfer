module neolong.me/file_transfer/biz

go 1.18

require (
	github.com/go-basic/uuid v1.0.0
	neolong.me/file_transfer/base v0.0.0
	neolong.me/file_transfer/util v0.0.0
)

require (
	neolong.me/neotools/cipher v0.0.0 // indirect
	neolong.me/neotools/common v0.0.0 // indirect
)

replace (
	neolong.me/file_transfer/base v0.0.0 => ../base
	neolong.me/file_transfer/util v0.0.0 => ../util
	neolong.me/neotools/cipher v0.0.0 => ../../neotools/cipher
	neolong.me/neotools/common v0.0.0 => ../../neotools/common
	neolong.me/neotools/neoio v0.0.0 => ../../neotools/neoio
)
