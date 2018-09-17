package chatserver

import (
	"net/http"
	"sync"
	"time"

	"github.com/cocoonlife/timber"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	ContentType    = "Content-Type"
	DefaultPortNum = "8080"
	MimeTypeJSON   = "application/json"
)

func NewClients() *Clients {
	return &Clients{members: make(map[*websocket.Conn]struct{})}
}

type Clients struct {
	members map[*websocket.Conn]struct{}
	sync.RWMutex
}

func (c *Clients) Add(conn *websocket.Conn) {
	c.Lock()
	defer c.Unlock()
	timber.Infof("adding client. %d -> %d clients", len(c.members), len(c.members)+1)
	c.members[conn] = struct{}{}
}

func (c *Clients) Delete(conn *websocket.Conn) {
	c.Lock()
	defer c.Unlock()
	timber.Infof("deleting client. %d -> %d clients", len(c.members), len(c.members)-1)
	delete(c.members, conn)
}

func (c *Clients) Broadcast(msg Message) {
	c.Lock()
	defer c.Unlock()
	for client := range c.members {
		err := client.WriteJSON(msg)
		if err != nil {
			timber.Errorf(err.Error())
			client.Close()
			delete(c.members, client)
		}
	}
}

type App struct {
	server        *http.Server
	broadcastChan chan Message
	clients       *Clients
}

func NewApp() *App {
	return &App{
		broadcastChan: make(chan Message),
		clients:       NewClients(),
	}
}

func (a *App) Init(portNum string) error {
	router := mux.NewRouter()

	router.HandleFunc("/",
		a.HealthHandler).Methods(http.MethodGet)

	router.HandleFunc("/ws",
		a.handleWebsocket).Methods(http.MethodGet)

	server := &http.Server{
		Addr:           ":" + portNum,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	a.server = server

	return nil
}

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	w.Header().Set(ContentType, MimeTypeJSON)
	if err != nil {
		timber.Errorf(err.Error())
		return
	}
	a.clients.Add(ws)
	defer a.clients.Delete(ws)

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			timber.Warnf(err.Error())
			return
		}
		timber.Infof("received message")
		a.broadcastChan <- msg
	}
}

func (a *App) broadcastMessages() {
	for msg := range a.broadcastChan {
		a.clients.Broadcast(msg)
	}
}

func (a *App) Run() error {
	timber.Infof("running server on %s", a.server.Addr)
	go a.broadcastMessages()
	return a.server.ListenAndServe()
}

func (a *App) HealthHandler(w http.ResponseWriter, r *http.Request) {
	timber.Infof("root handler")
}
