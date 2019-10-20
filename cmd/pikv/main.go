package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv"
	"github.com/joway/pikv/db"
	"github.com/soheilhy/cmux"
	"github.com/tidwall/redcon"
	"github.com/urfave/cli"
	"io"
	"net"
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

func matchNotGrpc() cmux.Matcher {
	m := cmux.HTTP2HeaderField("content-type", "application/grpc")
	return func(r io.Reader) bool { return !m(r) }
}

func startServer(cfg Config) error {
	database, err := db.New(db.Options{
		DBDir: cfg.dir,
	})
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("graceful shutdown")
		database.Close()
	}()
	database.Run()

	rdsAddr := fmt.Sprintf(":%s", cfg.port)
	rpcAddr := fmt.Sprintf(":%s", cfg.rpcPort)
	rdsLis, err := net.Listen("tcp", rdsAddr)
	logger.Info("running pikv server at: %s", rdsAddr)
	if err != nil {
		logger.Fatal("%v", err)
	}
	rpcLis, err := net.Listen("tcp", rpcAddr)
	logger.Info("running redis server at: %s", rpcAddr)
	if err != nil {
		logger.Fatal("%v", err)
	}

	rpcServer := db.NewRpcServer(database)
	go func() {
		if err := rpcServer.Serve(rpcLis); err != nil {
			logger.Fatal("%v", err)
		}
	}()

	redisServer := redcon.NewServer(
		"",
		db.GetRedisCmdHandler(database),
		nil,
		nil,
	)
	go func() {
		if err := redisServer.Serve(rdsLis); err != nil {
			logger.Fatal("%v", err)
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
		if err := redisServer.Close(); err != nil {
			logger.Error("%v", err)
		}
		rpcServer.GracefulStop()

	}
	return nil
}
