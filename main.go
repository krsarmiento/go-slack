package main

import (
    "net/http"
    r "gopkg.in/gorethink/gorethink.v3"
    "log"
)

type Channel struct {
    Id   string `json:"id" gorethink:"id,omitempty"`
    Name string `json:"name" gorethink:"name"`
}

type User struct {
    Id string `json:"id" orethink:"id,omitempty"`
    Name string `json:"name" gorethink:"name"`
}


func main() {
    session, err := r.Connect(r.ConnectOpts{
        Address: "localhost:28015",
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
    http.Handle("/", router)
    http.ListenAndServe(":4000", nil)
}
