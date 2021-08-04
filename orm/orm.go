package orm

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os/exec"

	// mysql driver

	"github.com/dedecms/snake"
	_ "github.com/go-sql-driver/mysql"
	"github.com/i582/cfmt/cmd/cfmt"
)

var (
	Conf  = new(Config)
	dbDSN string
)

type Config struct {
	User     string
	Pass     string
	Host     string
	Port     string
	Database string
	Charset  string
}

func GetCommon(file string) error {
	if f, ok := snake.FS(file).Open(); ok {
		for _, v := range f.String().Lines() {
			arr := snake.String(v).Split("=")
			if len(arr) == 2 {
				k := snake.String(arr[0]).Remove("\\'", "\\;", "\\$").Trim(" ").Get()
				v := snake.String(arr[1]).Remove("\\'", "\\;", "\\$").Trim(" ").Get()
				switch k {
				case "cfg_dbname":
					Conf.Database = v
				case "cfg_dbhost":
					host := snake.String(v).Split(":")
					if len(host) == 1 {
						Conf.Host = host[0]
						Conf.Port = "3306"
					}
					if len(host) == 2 {
						Conf.Host = host[0]
						Conf.Port = host[1]
					}
				case "cfg_dbuser":
					Conf.User = v
				case "cfg_dbpwd":
					Conf.Pass = v
				case "cfg_db_language":
					Conf.Charset = v
				case "cfg_dbtype":
					if v != "mysql" {
						return fmt.Errorf("无法转换非Mysql/MariaDB数据库的数据。")
					}
				}
			}
		}
	}

	dbDSN = snake.String(Conf.User).
		Add(":").
		Add(Conf.Pass).
		Add("@tcp(").
		Add(Conf.Host).
		Add(":").
		Add(Conf.Port).
		Add(")/").
		Add(snake.String("?charset=").Add(Conf.Charset).Get()).
		Get()

	return nil

}

func GetSQL(mysqldumpPATH, table string) *snake.SnakeString {
	cmd := exec.Command(mysqldumpPATH, "--skip-comments", "--opt", "-h"+Conf.Host, "-P"+Conf.Port, "-u"+Conf.User, "-p"+Conf.Pass, Conf.Database, table)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	if err := cmd.Start(); err != nil {
		return nil
	}
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil
	}
	return snake.String(string(bytes))
}

func GetTables() ([]string, error) {
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return nil, cfmt.Errorf("{{错误: 无法链接数据库。}}::red(" + err.Error() + ")")
	}
	defer db.Close()
	_, err = db.Exec("USE " + Conf.Database)
	if err != nil {
		return nil, cfmt.Errorf("{{错误: %s数据库不存在。}}::red("+err.Error()+")", Conf.Database)
	}

	res, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, cfmt.Errorf("{{错误: 无法执行SHOW TABLES命令}}::red(" + err.Error() + ")")
	}
	tables := make([]string, 0)
	var table string

	for res.Next() {
		res.Scan(&table)
		tables = append(tables, table)
	}

	return tables, nil
}

func ImportSQL(mysqlPATH, database, user, pass, file string) error {
	dbDSN := snake.String(user).
		Add(":").
		Add(pass).
		Add("@tcp(").
		Add(Conf.Host).
		Add(":").
		Add(Conf.Port).
		Add(")/").
		Add(snake.String("?charset=utf8").Get()).
		Get()
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return cfmt.Errorf("{{错误: 无法链接数据库。}}::red(" + err.Error() + ")")
	}
	defer db.Close()
	_, err = db.Exec("USE " + database)
	if err != nil {
		_, err = db.Exec("CREATE DATABASE " + database)
		if err != nil {
			return cfmt.Errorf("{{错误: 无法创建数据库。}}::red(" + err.Error() + ")")
		}
	}
	cmd := exec.Command(mysqlPATH, fmt.Sprintf("-u%s", user), fmt.Sprintf("-p%s", pass), database, "-e", fmt.Sprintf("source %s", file))
	if err := cmd.Run(); err != nil {
		return cfmt.Errorf("{{错误: 数据库无法导入数据。}}::red(" + err.Error() + ")")
	}
	return nil
}
