package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "prism"
	app.Usage = "api gateway"
	app.Version = "2.0.0"
	log.SetFlags(log.Ltime | log.Lshortfile)

	app.Commands = []cli.Command{
		{
			Name:   "host",
			Usage:  "start host",
			Action: host_start,
		},
		{
			Name:   "agent",
			Usage:  "start agent",
			Action: agent_start,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
