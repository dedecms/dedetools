package backup

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/dedecms/dedetools/clif"
	"github.com/dedecms/dedetools/log"
	"github.com/dedecms/dedetools/orm"
	"github.com/dedecms/dedetools/util"
	"github.com/dedecms/snake"
	"github.com/dedecms/snake/pkg"
	"github.com/i582/cfmt/cmd/cfmt"
)

type Module struct {
	mysqlPATH     string
	mysqldumpPATH string
}

var env = new(Module)
var now = time.Now().Format("200601021504")

func Init(in clif.Input, out clif.Output) {

	env.mysqlPATH = "mysql"
	env.mysqldumpPATH = "mysqldump"

	style := clif.DefaultStyles
	style["query"] = ""
	out.SetFormatter(clif.NewDefaultFormatter(style))
	cfmt.Println(snake.String("DedeCMS Manage Tools ").Ln().
		Add("http://www.dedecms.com").Ln().
		Add("Function: ").Add("DedeCMS数据备份").Ln().
		DrawBox(64, pkg.Box9Slice{
			Top:         "=",
			TopRight:    "#",
			Right:       "#",
			BottomRight: "#",
			Bottom:      "=",
			BottomLeft:  "#",
			Left:        "#",
			TopLeft:     "#",
		}).Get())

	l := log.Start("检查mysql是否可以执行")
MYSQL:
	if path, err := exec.LookPath(env.mysqlPATH); err != nil {
		l.Err(err)
		env.mysqlPATH = util.Ask("请输入mysql或mysql.exe的位置。", "", "string", in)
		goto MYSQL
	} else {
		env.mysqlPATH = path
	}
	l.Done()

	l = log.Start("检查mysqldump是否可以执行")
MYSQLDUMP:
	if path, err := exec.LookPath(env.mysqldumpPATH); err != nil {
		l.Err(err)
		env.mysqldumpPATH = util.Ask("请输入mysqldump或mysqldump.exe的位置。", "", "string", in)
		goto MYSQLDUMP
	} else {
		env.mysqldumpPATH = path
	}
	l.Done()

	wwwdir := util.Ask("请输入WEB服务器中DedeCMS根目录位置", "./", "existdir", in)
	outputDIR := util.Ask("请输入DedeCMS备份目录位置", "./backup_dedecms", "makedir", in)

BACKUPSQL:
	common := util.Ask("请输入./data/common.inc.php文件位置", "./data/common.inc.php", "existfile", in)
	orm.GetCommon(common)

	if err := backupSQL(outputDIR); err != nil {
		cfmt.Println(err.Error(), "\n")
		goto BACKUPSQL
	}
	backupWWW(wwwdir, outputDIR)

	dir := snake.FS(outputDIR).Add(now)
	cfmt.Println("备份完成: 已将数据备份至", dir.Get())
	fmt.Println()
}

func backupWWW(wwwDIR, outputDIR string) error {

	l := log.Start("获取网站文件列表")
	arr := snake.FS(wwwDIR).Find("*")
	l.Done()

	l = log.Start("对网站文件进行备份", len(arr))

	for _, v := range arr {

		dir := snake.FS(outputDIR).Add(now).Add("www")
		outfile := snake.String(v).ReplaceOne(wwwDIR, dir.Get())

		if wwwDIR == "./" {
			outfile = snake.String(dir.Get()).Add("/").Add(v)
		}

		if !strings.HasPrefix(v, snake.FS(outputDIR).Get()) {

			if snake.FS(v).IsDir() {
				snake.FS(outfile.Get()).MkDir()
			}

			if snake.FS(v).IsFile() {
				f, _ := snake.FS(v).Open()
				f.Close()
				snake.FS(outfile.Get()).ByteWriter(f.Byte())
			}
		}

		l.Add()
	}
	l.Done()
	return nil
}

func backupSQL(outputDIR string) error {
	tables, err := orm.GetTables()
	if err != nil {
		return err
	}

	l := log.Start("对数据库进行备份", len(tables))
	for _, v := range tables {
		outputFILE := snake.FS(outputDIR).Add(now).Add("sql").Add(snake.String("backup_", v, ".sql").Get())
		if !outputFILE.Exist() {
			outputFILE.MkFile()
		}
		outputFILE.Write(snake.String("DedeCMS Manage Tools Generation").Ln(2).
			Add("http://www.dedecms.com").Ln(2).
			Add("Host: ").Add(orm.Conf.Host).Ln().
			Add("Database: ").Add(orm.Conf.Database).Ln().
			Add("Table: ").Add(v).Ln().
			Add("Source Charset: ").Add(orm.Conf.Charset).Ln().
			Add("Generation Time: ").Add(time.Now().Format("2006-01-02 15:04:05")).
			DrawBox(64).Get())
		outputFILE.Write(orm.GetSQL(env.mysqldumpPATH, v).Get(), true)
		l.Add()
	}
	l.Done()
	return nil
}
