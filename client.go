package main

import (
    "github.com/gorilla/websocket"
    r "gopkg.in/gorethink/gorethink.v3"
    "fmt"
)

type FindHandler func(string) (Handler, bool)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Client struct {
    send         chan Message
    socket       *websocket.Conn
    findHandler  FindHandler
    session      *r.Session
    stopChannels map[int]chan bool
    userId       string
}

func (client *Client) NewStopChannel(stopKey int) chan bool {
    client.StopForKey(stopKey)
    stop := make(chan bool)
    client.stopChannels[stopKey] = stop
    return stop
}

func (client *Client) StopForKey(key int) {
    if ch, found := client.stopChannels[key]; found {
        ch <- true
        delete(client.stopChannels, key)
    }
}

func (client *Client) Read() {
    var message Message
    for {
        if err := client.socket.ReadJSON(&message); err != nil {
            break
        }
        if handler, found := client.findHandler(message.Name); found {
            handler(client, message.Data)
        }
    }
    client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.send {
        if err := client.socket.WriteJSON(msg); err != nil {
            break
        }
	}
    client.socket.Close()
}

func (client *Client) Close() {
    fmt.Println("Closing all subscription channels")
    for _, ch := range client.stopChannels {
        ch <- true
    }
    fmt.Println("Closing send channel")
    close(client.send)
    fmt.Println("[User Leaving] " + client.userId)
    _, _ = r.Table("user").
             Get(client.userId).
             Delete().
             Run(client.session)
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
    return &Client {
        send: make(chan Message),
        socket: socket,
        findHandler: findHandler,
        session: session,
        stopChannels: make(map[int]chan bool),
        userId: "",
    }
}
