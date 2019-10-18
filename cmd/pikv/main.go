package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv"
	"github.com/joway/pikv/db"
	"github.com/urfave/cli"
	"os"
	"os/signal"
)

var logger = loki.New("pikv:main")

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

		return startServer(cfg)
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("%v", err)
	}
}

func startServer(cfg Config) error {
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

	rpcServer, err := db.ListenRpcServer(database, rpcAddress)
	if err != nil {
		logger.Fatal("%v", err)
	}
	redisServer := db.ListenRedisProtoServer(database, externalAddress)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
		rpcServer.Stop()
		if err := redisServer.Close(); err != nil {
			return err
		}
	}
	return nil
}
