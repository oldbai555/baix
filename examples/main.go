package main

import (
	"github.com/oldbai555/lbtool/log"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "baiX"
	app.Version = "v0.0.1"
	app.Description = "lb tcp server"
	app.Action = serve
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("berr:%v", err)
		return
	}
}

func serve(c *cli.Context) error {
	return nil
}
