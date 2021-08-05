package info

import (
	"fmt"
	"time"

	"github.com/dedecms/dedetools/clif"
	"github.com/dedecms/snake"
)

func Copyright() string {
	now := snake.String(time.Now().Format("2006"))
	if now.Get() != "2021" {
		now = snake.String("2021 - ").Add(now.Get())
	}
	return snake.String("Copyright ").Add(now.Get()).Add(" 上海卓卓网络科技有限公司").Get()
}

func CallCopyright(c *clif.Command) {
	fmt.Println(Copyright())
}

func AppDesc() string {
	box := snake.String("* 内部版本，严禁外传 *").Ln(2).
		Add("DedeCMS Manage Tools 是由上海卓卓网络科技有限公司开发的\nDedeCMS系统管理工具集。").Ln(2).
		Add("官方网站：http://www.dedecms.com").Ln().
		Add("产品维护：织梦团队").Ln().
		Add("版权所有：").Add(Copyright()).Ln().
		DrawBox(65).Get()
	return snake.String().Ln().Add(box).Get()
}
