package main

import (
    "fmt"
    "encoding/json"
    "github.com/mitchellh/mapstructure"
)

func main() {
    recRawMsg := []byte(`{"name":"channel add",` +
        `"data":{"name":"Hardare Support"}}`)

    var recMessage Message
    err := json.Unmarshal(recRawMsg, &recMessage)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("%#v\n", recMessage)

    if recMessage.Name == "channel add" {
        channel, err := addChannel(recMessage.Data)
        var sendMessage Message
        sendMessage.Name = "channel add"
        sendMessage.Data = channel
        sendRawMsg, err := json.Marshal(sendMessage)
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Printf(string(sendRawMsg)) 
    }
}



