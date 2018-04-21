package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "svagent"
	app.Usage = "api gateway"
	app.Version = "1.0.0"
	log.SetFlags(log.Ltime | log.Lshortfile)

	app.Commands = []cli.Command{
		{
			Name:   "host",
			Usage:  "start host",
			Action: host_start,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "http",
					Value: "0.0.0.0:5555",
					Usage: "http listen",
				},
				cli.StringFlag{
					Name:  "server",
					Value: "0.0.0.0:6667",
					Usage: "host listen",
				},
			},
		},
		{
			Name:   "agent",
			Usage:  "start agent",
			Action: agent_start,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "local",
					Value: "127.0.0.1:8000",
					Usage: "local server",
				},
				cli.StringFlag{
					Name:  "remote",
					Value: "127.0.0.1:6667",
					Usage: "remote server",
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
