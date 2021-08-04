package utf8

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/dedecms/dedetools/encode"
	"github.com/dedecms/dedetools/log"
	"github.com/dedecms/dedetools/orm"
	"github.com/dedecms/dedetools/util"
	"github.com/dedecms/snake"
	"github.com/dedecms/snake/pkg"
	"github.com/i582/cfmt/cmd/cfmt"
	"github.com/ukautz/clif"
)

type Module struct {
	mysqlPATH     string
	mysqldumpPATH string
}

var env = new(Module)

func Init(in clif.Input, out clif.Output) {

	env.mysqlPATH = "mysql"
	env.mysqldumpPATH = "mysqldump"

	style := clif.DefaultStyles
	style["query"] = ""
	out.SetFormatter(clif.NewDefaultFormatter(style))
	cfmt.Println(snake.String("DedeCMS Manage Tools ").Ln().
		Add("http://www.dedecms.com").Ln().
		Add("Function: ").Add("将GBK或BIG5版本的DedeCMS转码为UTF8").Ln().
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

	l := log.Start("检测mysql是否可以执行")
MYSQL:
	if path, err := exec.LookPath(env.mysqlPATH); err != nil {
		l.Err(err)
		env.mysqlPATH = util.Ask("请输入mysql或mysql.exe的位置。", "", "string", in)
		goto MYSQL
	} else {
		env.mysqlPATH = path
	}
	l.Done()

	l = log.Start("检测mysqldump是否可以执行")
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
	outputDIR := util.Ask("请输入DedeCMS转换后存储目录位置", "./output_dedecms_utf8", "makedir", in)

BACKUPSQL:
	common := util.Ask("请输入common.inc.php文件位置", "./data/common.inc.php", "existfile", in)
	orm.GetCommon(common)

	if err := backupSQL(outputDIR); err != nil {
		cfmt.Println(err.Error(), "\n")
		goto BACKUPSQL
	}
	backupWWW(wwwdir, outputDIR)
	fmt.Println()
	database := util.Ask("请输入UTF8导入数据库名", "dedecms_gbk_to_utf8", "database", in)
AUTH:
	user := util.Ask("请输入用于创建新数据库的用户名", orm.Conf.User, "string", in)
	pass := util.Ask("请输入密码", orm.Conf.Pass, "string", in)
	if err := importSQL(database, user, pass, outputDIR); err != nil {
		goto AUTH
	}

}

func backupWWW(wwwDIR, outputDIR string) error {

	l := log.Start("获取网站文件列表")
	arr := snake.FS(wwwDIR).Find("*")
	l.Done()

	l = log.Start("对网站文件进行转码", len(arr))
	cext := []string{
		".html",
		".htm",
		".php",
		".txt",
		".xml",
		".js",
		".css",
	}

	for _, v := range arr {

		dir := snake.FS(outputDIR).Add("www")
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
				defer f.Close()
				bytes := f.Byte()
				if snake.String(snake.FS(v).Ext()).ExistSlice(cext) {
					utf8, _ := encode.GetEncoding(bytes)
					if utf8.Charset != "UTF-8" {
						bytes = utf8.Bytes()
					}
				}

				snake.FS(outfile.Get()).ByteWriter(bytes)
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

	l := log.Start("将gbk、big5数据库转码为utf8", len(tables))
	for _, v := range tables {
		outputFILE := snake.FS(outputDIR).Add("sql").Add(snake.String("backup_", v, ".sql").Get())
		if !outputFILE.Exist() {
			outputFILE.MkFile()
		}
		outputFILE.Write(snake.String("DedeCMS Manage Tools Generation").Ln(2).
			Add("http://www.dedecms.com").Ln(2).
			Add("Host: ").Add(orm.Conf.Host).Ln().
			Add("Database: ").Add(orm.Conf.Database).Ln().
			Add("Table: ").Add(v).Ln().
			Add("Charset: ").Add("UTF-8").Ln().
			Add("Source Charset: ").Add(orm.Conf.Charset).Ln().
			Add("Generation Time: ").Add(time.Now().Format("2006-01-02 15:04:05")).
			DrawBox(64).Get())
		outputFILE.Write(orm.GetSQL(env.mysqldumpPATH, v).Replace(`(ENGINE.*CHARSET=)(gbk|big5)(\;)`, "${1}utf8${3}").Get(), true)
		l.Add()
	}
	l.Done()
	return nil
}

func importSQL(database, user, pass, outputDIR string) error {
	sql := snake.FS(outputDIR).Add("sql")
	files := sql.Find("*.sql")
	l := log.Start("将转码完毕的SQL文件导入数据库", len(files))
	for _, v := range files {
		if err := orm.ImportSQL(env.mysqlPATH, database, user, pass, v); err != nil {
			l.Err(err)
			return err
		}
		l.Add()
	}
	l.Done()
	return nil
}
