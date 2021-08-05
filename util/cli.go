package util

import (
	"fmt"

	"github.com/dedecms/dedetools/clif"
	"github.com/dedecms/dedetools/orm"
	"github.com/dedecms/snake"
)

func Ask(t, d, o string, input clif.Input) string {
	message := ""
	title := snake.String(t).Add(" [默认值: ").Add(d).Add("]:").Get()
	if d == "" {
		title = snake.String(t).Add(":").Get()
	}
	input.Ask(title, func(v string) error {
		if len(v) > 0 {
			if input.Confirm(snake.String("是否使用当前值 [").Add(v).Add("]？ (y/n) :").Get()) {
				switch o {
				case "existdir":
					if err := checkexistdir(v); err != nil {
						return err
					}
				case "existfile":
					if err := checkexistfile(v); err != nil {
						return err
					}
				case "makedir":
					if err := checkmakedir(v, input); err != nil {
						return err
					}
				case "database":
					if v == orm.Conf.Database {
						return fmt.Errorf("安全警告: 不能直接覆盖原数据库。")
					}
				}
				message = v
				return nil
			} else {
				return fmt.Errorf("操作已取消, 请继续操作。")
			}
		} else if d != "" {
			if input.Confirm(snake.String("是否使用默认值 [").Add(d).Add("]？ (y/n) :").Get()) {
				switch o {
				case "existdir":
					if err := checkexistdir(d); err != nil {
						return err
					}
				case "existfile":
					if err := checkexistfile(d); err != nil {
						return err
					}
				case "makedir":
					if err := checkmakedir(d, input); err != nil {
						return err
					}

				}
				message = d
				return nil
			} else {
				return fmt.Errorf("操作已取消, 请继续操作。")
			}
		}
		return fmt.Errorf("输入不能为空。")

	})
	return message
}

func checkexistdir(d string) error {
	if !snake.FS(d).Exist() {
		return fmt.Errorf("错误: 目标不存在，请重新输入。")
	}

	if !snake.FS(d).IsDir() {
		return fmt.Errorf("错误: 目标不是目录，请重新输入。")
	}

	return nil
}
func checkexistfile(d string) error {
	if !snake.FS(d).Exist() {
		return fmt.Errorf("错误: 目标不存在，请重新输入。")
	}

	if !snake.FS(d).IsFile() {
		return fmt.Errorf("错误: 目标不是文件，请重新输入。")
	}

	return nil
}

func checkmakedir(d string, input clif.Input) error {
	if snake.FS(d).Exist() && !snake.FS(d).IsDir() {
		return fmt.Errorf("错误: 目标不是目录，请重新输入。")
	}
	if !snake.FS(d).Exist() {
		if input.Confirm(snake.String("目录不存在，是否要创建目录 [").Add(d).Add("]？ (y/n) :").Get()) {
			snake.FS(d).MkDir()
			return nil
		} else {
			return fmt.Errorf("操作已取消, 请继续操作。")
		}
	}
	return nil
}
