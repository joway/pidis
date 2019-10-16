package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/executor"
	"github.com/joway/pikv/rpc"
	"github.com/joway/pikv/util"
	"github.com/tidwall/redcon"
	"github.com/urfave/cli"
	"net"
	"os"
)

var logger = loki.New("main")

type Config struct {
	port    string
	rpcPort string
	dir     string
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
			Name:  "rpcPort",
			Value: "6381",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: "/tmp/pikv",
		},
	}
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		rpcPort := c.String("rpcPort")
		dir := c.String("dir")
		cfg := Config{
			port:    port,
			rpcPort: rpcPort,
			dir:     dir,
		}

		return setup(cfg)
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("%v", err)
	}
}

func setup(cfg Config) error {
	database, err := db.New(db.Options{
		DBDir: cfg.dir,
	})
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("Graceful ActionShutdown")
		database.Close()
	}()
	database.Run()

	externalAddress := fmt.Sprintf(":%s", cfg.port)
	rpcAddress := fmt.Sprintf(":%s", cfg.rpcPort)

	go func() {
		if err := rpcServer(database, rpcAddress); err != nil {
			logger.Fatal("%v", err)
		}
	}()

	return redisServer(database, externalAddress)
}

func redisServer(database *db.Database, address string) error {
	if err := redcon.ListenAndServe(
		address,
		func(conn redcon.Conn, cmd redcon.Command) {
			defer func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			}()
			out, action, err := database.Exec(cmd.Args)
			if err != nil {
				logger.Error("%v", err)
				return
			}
			switch action {
			case executor.ActionNone:
				if len(out) > 0 {
					conn.WriteRaw(out)
				}
			case executor.ActionUnknown:
				conn.WriteRaw(util.MessageError(common.ErrUnknown))
			case executor.ActionRuntimeError:
				conn.WriteRaw(util.MessageError(common.ErrRuntimeError))
			case executor.ActionInvalidNumberOfArgs:
				conn.WriteRaw(util.MessageError(common.ErrInvalidNumberOfArgs))
			case executor.ActionInvalidSyntax:
				conn.WriteRaw(util.MessageError(common.ErrSyntaxError))
			case executor.ActionClose:
				if err := conn.Close(); err != nil {
					logger.Fatal("Connection ActionClose Failed:\n%v", err)
				}
			case executor.ActionShutdown:
				logger.Fatal("Shutting server down, bye bye")
			}
		},
		nil,
		nil,
	); err != nil {
		return err
	}
	return nil
}

func rpcServer(database *db.Database, address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := rpc.NewRpcServer(database)
	if err := server.Serve(listen); err != nil {
		return err
	}
	return nil
}
