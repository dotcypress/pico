package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

var Version string = "0.0.0"
var Release bool = false
var log = logging.MustGetLogger("PICOCDN")

var store Store

func openStore(path string) error {
	s, err := NewStore(path)
	if err != nil {
		return err
	}
	store = s
	return nil
}

func main() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackend("")
	if err != nil {
		log.Fatal(err)
	}
	logging.SetBackend(logBackend, syslogBackend)
	logging.SetFormatter(logging.MustStringFormatter("%{color}PICO %{time:15:04:05.00} %{level:-9.9s} %{color:reset} %{message}"))

	if Release {
		logging.SetLevel(logging.ERROR, "PICOCDN")
	} else {
		logging.SetLevel(logging.DEBUG, "PICOCDN")
	}

	app := cli.NewApp()
	app.Name = "pico-cdn"
	app.Version = Version
	app.Usage = "Damn small CDN"
	app.Author = "SmartBoard team"
	app.Email = "root@smart-board.tv"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "network",
			Value:  ":8080",
			Usage:  "listening interface",
			EnvVar: "PICO_INTERFACE",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "master",
			ShortName: "m",
			Usage:     "Starts server as master node",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "store",
					Value:  "/data/pico-cdn",
					Usage:  "store path",
					EnvVar: "PICO_STORE",
				},

				cli.StringFlag{
					Name:   "uploadKey",
					Value:  "secret",
					Usage:  "upload authorization key",
					EnvVar: "PICO_UPLOAD_KEY",
				},
			},
			Action: func(c *cli.Context) {
				err := openStore(c.String("store"))
				if err != nil {
					log.Critical("Can't start server: %v", err)
					return
				}
				Serve(c.GlobalString("network"), c.String("uploadKey"), true)
			},
		},
	}
	app.Run(os.Args)
}
