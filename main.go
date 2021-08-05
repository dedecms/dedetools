package main

import (
	"fmt"

	"github.com/dedecms/dedetools/backup"
	"github.com/dedecms/dedetools/clif"
	"github.com/dedecms/dedetools/info"
	"github.com/dedecms/dedetools/restore"
	"github.com/dedecms/dedetools/utf8"
	"github.com/kenkyu392/go-safe"
)

func main() {
	if err := safe.Do(func() error {
		clif.New("DedeCMS Manage Tools", "0.4.1", info.AppDesc()).
			New("utf8", "将DedeCMS GBK/BIG5等编码转码为UTF-8", utf8.Init).
			New("backup", "DedeCMS数据备份", backup.Init).
			New("restore", "DedeCMS数据恢复", restore.Init).
			New("copyright", "显示DedeCMS Manage Tools版权信息", info.CallCopyright).
			Run()
		return nil
	}); err != nil {
		fmt.Println("错误: ", err)
	}
}
