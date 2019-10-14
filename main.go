package main

import (
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
)

var logger = loki.New("main")

func main() {
	serve()
}

func serve() {
	opt := storage.Options{Dir: "/tmp/badger"}
	store, err := storage.NewBadgerStorage(opt)
	if err != nil {
		logger.Fatal("%v", err)
	}
	defer store.Close()

	if err := redcon.ListenAndServe(fmt.Sprintf(":%d", 6370),
		func(conn redcon.Conn, cmd redcon.Command) {
			context := types.Context{
				Args:    cmd.Args,
				Storage: store,
			}
			out, action := parser.Parse(context)
			if len(out) > 0 {
				conn.WriteRaw(out)
			}
			if action == parser.Close {
				if err := conn.Close(); err != nil {
					logger.Fatal("Connection Close Failed:\n%v", err)
				}
			}
			if action == parser.Shutdown {
				logger.Fatal("Shutting server down, bye bye")
			}
		}, nil, nil); err != nil {
		logger.Fatal("%v", err)
	}
}
