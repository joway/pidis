package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/db"
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
	app.Version = pikv.VERSION
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

	if err := app.Run(os.Args); err != nil {
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
			out, err := database.Exec(cmd.Args)
			switch err {
			case nil:
				conn.WriteRaw(out)
			case common.ErrUnknownCommand:
				conn.WriteRaw(util.MessageError(fmt.Sprintf(
					"ERR unknown command '%s'",
					cmd.Args[0],
				)))
			case common.ErrUnknown:
				conn.WriteRaw(util.MessageError(common.ErrUnknown.Error()))
			case common.ErrRuntimeError:
				conn.WriteRaw(util.MessageError(common.ErrRuntimeError.Error()))
			case common.ErrInvalidNumberOfArgs:
				conn.WriteRaw(util.MessageError(common.ErrInvalidNumberOfArgs.Error()))
			case common.ErrSyntaxError:
				conn.WriteRaw(util.MessageError(common.ErrSyntaxError.Error()))
			case common.ErrCloseConn:
				if err := conn.Close(); err != nil {
					logger.Error("connection close Failed:\n%v", err)
				}
			case common.ErrShutdown:
				logger.Fatal("shutting server down, bye bye")
			default:
				logger.Error("Unhandled error : %v", err)
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
