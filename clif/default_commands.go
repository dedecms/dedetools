package clif

import "fmt"

// NewHelpCommand returns the default help command
func NewHelpCommand() *Command {
	return NewCommand("help", "显示帮助", func(o *Command, out Output) error {
		if n := o.Argument("command").String(); n != "" {
			if cmd, ok := o.Cli.Commands[n]; ok {
				out.Printf(DescribeCommand(cmd))
			} else {
				out.Printf(DescribeCli(o.Cli))
				return fmt.Errorf("Unknown command \"%s\"", n)
			}
		} else {
			out.Printf(DescribeCommand(o))
		}
		return nil
	}).NewArgument("command", "显示帮助的命令", "", false, false)
}

// NewListCommand returns the default help command
func NewListCommand() *Command {
	return NewCommand("list", "可用命令列表", func(c *Cli, Command, out Output) {
		out.Printf(DescribeCli(c))
	})
}
