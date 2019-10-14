package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
	"github.com/urfave/cli"
	"os"
)

var logger = loki.New("main")

type Config struct {
	port    string
	dataDir string
}

func main() {
	app := cli.NewApp()
	app.Name = "pikv"
	app.Version = "0.0.1"
	app.Usage = ""
	cli.VersionFlag = cli.BoolFlag{
		Name: "version, v",
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Value: "6380",
		},
		cli.StringFlag{
			Name:  "dataDir, d",
			Value: "/tmp/pikv",
		},
	}
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		dataDir := c.String("dataDir")
		cfg := Config{
			port:    port,
			dataDir: dataDir,
		}
		serve(cfg)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("%v", err)
	}
}

func serve(cfg Config) {
	opt := storage.Options{
		Storage: storage.TypeMemory,
		Dir:     cfg.dataDir,
	}
	store, err := storage.NewStorage(opt)
	if err != nil {
		logger.Fatal("%v", err)
	}
	defer func() {
		loki.Info("Graceful Shutdown")
		store.Close()
	}()

	if err := redcon.ListenAndServe(fmt.Sprintf(":%s", cfg.port),
		func(conn redcon.Conn, cmd redcon.Command) {
			defer func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			}()
			context := types.Context{
				Out:     nil,
				Args:    cmd.Args,
				Storage: store,
			}
			out, action := parser.Parse(context)

			if len(out) > 0 {
				conn.WriteRaw(out)
			}

			if action == types.Close {
				if err := conn.Close(); err != nil {
					logger.Fatal("Connection Close Failed:\n%v", err)
				}
			}
			if action == types.Shutdown {
				logger.Fatal("Shutting server down, bye bye")
			}
		}, nil, nil); err != nil {
		logger.Fatal("%v", err)
	}
}
