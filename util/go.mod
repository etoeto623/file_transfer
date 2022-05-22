module neolong.me/file_transfer/util

go 1.18

require (
	neolong.me/file_transfer/base v0.0.0
	neolong.me/neotools/cipher v0.0.0
)

require neolong.me/neotools/common v0.0.0 // indirect

replace (
	neolong.me/file_transfer/base => ../base
	neolong.me/neotools/cipher v0.0.0 => /Users/longhai/go/src/neolong.me/neotools/cipher
	neolong.me/neotools/common v0.0.0 => /Users/longhai/go/src/neolong.me/neotools/common
)
