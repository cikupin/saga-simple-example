package main

import (
	"os"
	"sort"

	"github.com/cikupin/saga-simple-example/payment"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Saga PAttern"
	app.Usage = "A simple saga pattern example in go"
	app.UsageText = "[global options] command [command options] [arguments...]"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		payment.Serve,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)

}
