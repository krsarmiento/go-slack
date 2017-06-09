package main

import (
	r "gopkg.in/gorethink/gorethink.v3"
	"log"
	"net/http"
)

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type MessageChannel struct {
	Id        string `json:"id" gorethink:"id,omitempty"`
	Author    string `json:"author" gorethink:"author"`
	CreatedAt string `json:"createdAt" gorethink:"createdAt"`
	Body      string `json:"body" gorethink:"body"`
	ChannelId string `json:"channelId" gorethink:"channelId"`
}

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "rtsupport",
	})

	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)
	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)
	router.Handle("user edit", userEdit)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("message add", addMessage)
	router.Handle("message subscribe", subscribeMessage)
	router.Handle("message unsubscribe", subscribeMessage)
	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
