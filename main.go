package main

import (
	"fmt"

	"github.com/dedecms/dedetools/info"
	"github.com/dedecms/dedetools/utf8"
	"github.com/kenkyu392/go-safe"
	"github.com/ukautz/clif"
)

func main() {
	if err := safe.Do(func() error {
		clif.New("DedeCMS Manage Tools", "0.0.1", info.AppDesc()).
			New("utf8", "DedeCMS any charssset to utf-8", utf8.Init).
			New("backup", "Backup DedeCMS", utf8.Init).
			New("copyright", "Show DedeCMS manage tools copyright.", info.CallCopyright).
			Run()
		return nil
	}); err != nil {
		fmt.Println("错误: ", err)
	}
}
