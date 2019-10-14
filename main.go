package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
	"github.com/urfave/cli"
	"os"
)

type Config struct {
	port string
	dir  string
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
			Name:  "dir, d",
			Value: "/tmp/pikv",
		},
	}
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		dir := c.String("dir")
		cfg := Config{
			port: port,
			dir:  dir,
		}
		serve(cfg)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		loki.Fatal("%v", err)
	}
}

func serve(cfg Config) {
	database, err := db.NewDatabase(db.Options{
		DBDir: cfg.dir,
	})
	if err != nil {
		loki.Fatal("%v", err)
	}
	defer func() {
		loki.Info("Graceful Shutdown")
		database.Close()
	}()

	if err := redcon.ListenAndServe(fmt.Sprintf(":%s", cfg.port),
		func(conn redcon.Conn, cmd redcon.Command) {
			defer func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			}()
			context := types.Context{
				Out:  nil,
				Args: cmd.Args,
				DB:   database,
			}
			fmt.Printf("%s", cmd.Args)
			if err := database.Record(cmd.Args); err != nil {
				loki.Error("%v", err)
				return
			}
			fmt.Println("sdasda\n")
			out, action := parser.Parse(context)
			fmt.Printf("%s", out)

			if len(out) > 0 {
				conn.WriteRaw(out)
			}

			if action == types.Close {
				if err := conn.Close(); err != nil {
					loki.Fatal("Connection Close Failed:\n%v", err)
				}
			}
			if action == types.Shutdown {
				loki.Fatal("Shutting server down, bye bye")
			}
		}, nil, nil); err != nil {
		loki.Fatal("%v", err)
	}
}
