package main

import (
	"flag"

	"github.com/aultimus/chatserver"
	"github.com/cocoonlife/timber"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	timber.AddLogger(timber.ConfigLogger{
		LogWriter: new(timber.ConsoleWriter),
		Level:     timber.DEBUG,
		Formatter: timber.NewPatFormatter("[%D %T] [%L] %s %M"),
	})

	timber.Infof("chatserver started")

	var portNum = flag.String("port", chatserver.DefaultPortNum,
		"TCP/IP port that this program listens on")

	flag.Parse()

	// pprof
	go func() {
		timber.Errorf(http.ListenAndServe(":6060", nil))
	}()

	app := chatserver.NewApp()
	err := app.Init(*portNum)
	if err != nil {
		timber.Fatalf(err.Error())
	}
	err = app.Run()
	if err != nil {
		timber.Fatalf(err.Error())
	}
}
