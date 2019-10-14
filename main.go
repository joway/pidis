package main

import (
	"flag"
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
)

var logger = loki.New("main")

func main() {
	var isVersion bool
	flag.BoolVar(&isVersion, "v", false, "-v Show version")
	flag.Parse()
	if isVersion {
		fmt.Println("0.0.1")
		return
	}

	serve()
}

func serve() {
	opt := storage.Options{
		Storage: storage.TypeMemory,
		//Storage: storage.TypeBadger,
		Dir: "/tmp/badger",
	}
	store, err := storage.NewStorage(opt)
	if err != nil {
		logger.Fatal("%v", err)
	}
	defer func() {
		loki.Info("Graceful Shutdown")
		store.Close()
	}()

	if err := redcon.ListenAndServe(fmt.Sprintf(":%d", 6380),
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
