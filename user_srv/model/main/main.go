package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func main() {

}

func genMd5(code string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, code)
	return hex.EncodeToString(hash.Sum(nil))
}
