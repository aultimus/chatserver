package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/aultimus/chatserver"
	"github.com/cocoonlife/timber"
	"github.com/gorilla/websocket"
	"github.com/satori/uuid"
)

// It was not strictly necessary to write a client as we could use a 3rd party one.
// However, it is a good exericise in cementing our understanding
// of how the protocol works and allows us to have control of both sides

func main() {
	timber.AddLogger(timber.ConfigLogger{
		LogWriter: new(timber.ConsoleWriter),
		Level:     timber.DEBUG,
		Formatter: timber.NewPatFormatter("[%D %T] [%L] %s %M"),
	})

	timber.Infof("chatserver client started")

	var address = flag.String("address", "localhost:8080", "address to connect to")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *address, Path: "/ws"}
	timber.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		timber.Fatalf(err.Error())
	}
	defer c.Close()

	done := make(chan struct{})

	// TODO: close websocket properly on ctrl+c without sending ctrl+c to server
	go func() {
		defer close(done)
		for {
			_, b, err := c.ReadMessage()
			if err != nil {
				timber.Errorf(err.Error())
				return
			}
			var msg chatserver.Message
			err = json.Unmarshal(b, &msg)
			if err != nil {
				timber.Errorf(err.Error())
				return
			}

			fmt.Printf("[%s] %s", msg.Username, msg.Message)
		}
	}()

	uuid := uuid.NewV4().String()
	msg := chatserver.Message{Username: uuid, Email: uuid + "@gmail.com"}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for {
		select {
		case <-signalChan:
			return
		case <-done:
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if text == "\n" {
				continue
			}
			msg.Message = text
			b, err := json.Marshal(msg)
			if err != nil {
				panic(err.Error())
				return
			}
			err = c.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				panic(err.Error())
				return
			}
		}
	}
}
