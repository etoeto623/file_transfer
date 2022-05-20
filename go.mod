module neolong.me/file_transfer

go 1.18

replace (
	neolong.me/file_transfer/base => ./base
	neolong.me/file_transfer/util => ./util
)

require (
	github.com/go-basic/uuid v1.0.0
	neolong.me/file_transfer/base v0.0.0
	neolong.me/file_transfer/util v0.0.0-00010101000000-000000000000
)
