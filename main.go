package main

import (
	"os"
	"sort"

	"github.com/cikupin/saga-simple-example/item"
	"github.com/cikupin/saga-simple-example/orchestrator"
	"github.com/cikupin/saga-simple-example/order"
	"github.com/cikupin/saga-simple-example/payment"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Saga Pattern"
	app.Usage = "A simple saga pattern example in go"
	app.UsageText = "go run main.go [command]"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		orchestrator.Serve,
		item.Serve,
		order.Serve,
		payment.Serve,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}
