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
	outputDIR := util.Ask("请输入DedeCMS转换后存储目录位置", "./output_dedecms_utf8", "makedir", in)

BACKUPSQL:
	common := util.Ask("请输入./data/common.inc.php文件位置", "./data/common.inc.php", "existfile", in)
	orm.GetCommon(common)

	if err := backupSQL(outputDIR); err != nil {
		cfmt.Println(err.Error(), "\n")
		goto BACKUPSQL
	}
	backupWWW(wwwdir, outputDIR)
	fmt.Println()
	database := util.Ask("请输入UTF8数据库名 <warn>(禁止覆盖原数据库)<reset>", "db_output_dedecms_utf8", "database", in)
AUTH:
	user := util.Ask("请输入用于创建新数据库的用户名", orm.Conf.User, "string", in)
	pass := util.Ask("请输入密码", orm.Conf.Pass, "string", in)
	if err := importSQL(database, user, pass, outputDIR); err != nil {
		goto AUTH
	}
	commonPATH := snake.FS(outputDIR, "www/data/common.inc.php").Get()
	editCommon(commonPATH, database, user, pass)

}

func editCommon(file, database, user, pass string) {
	l := log.Start("更新./data/common.inc.php文件")
	defer l.Done()
	f, ok := snake.FS(file).Open()
	if !ok {
		return
	}
	defer f.Close()
	f.String().
		Replace(`((.*|)\$cfg_dbname(.*)=(.*)("|')).*(('|");)`, "${1}"+database+"${6}"). // 替换数据库名
		Replace(`((.*|)\$cfg_dbuser(.*)=(.*)("|')).*(('|");)`, "${1}"+user+"${6}").     // 替换用户名
		Replace(`((.*|)\$cfg_dbpwd(.*)=(.*)("|')).*(('|");)`, "${1}"+pass+"${6}").      // 替换密码
		Write(file)
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
		".inc",
	}

	rep := []string{
		".html",
		".htm",
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
						if snake.String(snake.FS(v).Ext()).ExistSlice(rep) {
							bytes = snake.String(utf8.Text()).
								Replace(`(<meta ((http-equiv|content).*(http-equiv|content)|).*charset.*=.*)(?i)(gbk|gb2312|big5)(.*>)`, "${1}utf-8${6}"). // 替换 *.HTML, *.HTM META CHARSET
								Byte()
						} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".xml"}) {
							bytes = snake.String(utf8.Text()).
								Replace(`(lang(.*|)=(.*|))(?i)(gbk|gb2312|big5)`, "${1}utf-8"). // 替换 *.XML LANG
								Byte()
						} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".js"}) {
							bytes = snake.String(utf8.Text()).
								Replace(`((.*|)this.sendlang(.*|)=(.*|))(?i)(gbk|gb2312|big5)`, "${1}utf-8"). // 替换 *.JS this.sendlang
								Byte()
						} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".css"}) {
							bytes = snake.String(utf8.Text()).
								Replace(`((.*|)@charset.*("|'))(?i)(gbk|gb2312|big5)`, "${1}utf-8"). // 替换 *.CSS charset
								Byte()
						} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".php"}) {
							bytes = snake.String(utf8.Text()).
								Replace(`((.*|)\$cfg_db_language(.*)=(.*)("|'))(?i)(gbk|gb2312|big5)`, "${1}utf8mb4"). // 替换 *.CSS charset
								Replace(`((.*|)\$cfg_soft_lang(.*)=(.*)("|'))(?i)(gbk|gb2312|big5)`, "${1}utf-8").     // 替换 *.CSS charset
								Replace(`((.*|)\$cfg_version(.*)=(.*)("|')V(.*)_)(?i)(gbk|gb2312|big5)`, "${1}UTF8").  // 替换 *.CSS charset
								Byte()
						} else {
							bytes = utf8.Bytes()
						}
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
			Add("Charset: ").Add("UTF8MB4").Ln().
			Add("Source Charset: ").Add(orm.Conf.Charset).Ln().
			Add("Generation Time: ").Add(time.Now().Format("2006-01-02 15:04:05")).
			DrawBox(64).Get())
		outputFILE.Write(orm.GetSQL(env.mysqldumpPATH, v).Replace(`(ENGINE.*CHARSET=)(gbk|big5)(\;)`, "${1}utf8mb4${3}").Get(), true)
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
