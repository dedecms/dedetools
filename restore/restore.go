package restore

import (
	"fmt"
	"os/exec"
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
	mysqlPATH string
}

var env = new(Module)
var now = time.Now().Format("200601021504")

func Init(in clif.Input, out clif.Output) {

	style := clif.DefaultStyles
	style["query"] = ""
	out.SetFormatter(clif.NewDefaultFormatter(style))
	cfmt.Println(snake.String("DedeCMS Manage Tools ").Ln().
		Add("http://www.dedecms.com").Ln().
		Add("Function: ").Add("DedeCMS数据恢复").Ln().
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

	mode := ""
	m := map[string]string{
		"1": "恢复DedeCMS数据库",
	}
	for {
		mode = in.Choose("请选择恢复模式", m)
		if in.Confirm(fmt.Sprintf("确定%s? (y/n)", m[mode])) {
			break
		} else {
			out.Printf("\n")
		}
	}

	switch mode {
	case "1":
		restoredatabase(in)
	}
}

func restoredatabase(in clif.Input) {

	env.mysqlPATH = "mysql"

	l := log.Start("检查mysql是否可以执行")
MYSQL:
	if path, err := exec.LookPath(env.mysqlPATH); err != nil {
		l.Err(err)
		env.mysqlPATH = util.Ask("请输入mysql或mysql.exe的位置。", "", "file", in)
		goto MYSQL
	} else {
		env.mysqlPATH = path
	}

	l.Done()

	outputDIR := util.Ask("请输入要恢复的DedeCMS数据库备份目录", "./backup_dedecms", "existdir", in)

	database := util.Ask("请输入数据库名 <warn>(禁止覆盖原数据库)<reset>", "db_output_dedecms_utf8", "database", in)
AUTH:
	user := util.Ask("请输入用于创建新数据库的用户名", orm.Conf.User, "string", in)
	pass := util.Ask("请输入密码", orm.Conf.Pass, "string", in)
	charset := util.Ask("数据库编码格式", "utf8", "string", in)
	if err := importSQL(database, user, pass, charset, outputDIR); err != nil {
		goto AUTH
	}
	fmt.Println()
}

func importSQL(database, user, pass, charset, outputDIR string) error {
	sql := snake.FS(outputDIR)
	files := sql.Find("*.sql")
	l := log.Start("将SQL文件导入数据库", len(files))
	for _, v := range files {
		if err := orm.ImportSQL(env.mysqlPATH, database, user, pass, charset, v); err != nil {
			l.Err(err)
			return err
		}
		l.Add()
	}
	l.Done()
	return nil
}
