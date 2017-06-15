package main

import(
    "github.com/mitchellh/mapstructure"
    r "gopkg.in/gorethink/gorethink.v3"
    "fmt"
    "time"
)

const (
    ChannelStop = iota
    UserStop
    MessageStop
)

func addChannel(client *Client, data interface{}) {
    var channel Channel
    err := mapstructure.Decode(data, &channel)
    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }
    go func() {
        err := r.Table("channel").
          Insert(channel).
          Exec(client.session)
        if err != nil {
            client.send <- Message{"error", err.Error()}
        }
    }()
}

func subscribeChannel(client *Client, data interface{}) {
    stop := client.NewStopChannel(ChannelStop)
    result := make(chan r.ChangeResponse)
    cursor, err := r.Table("channel").
        Changes(r.ChangesOpts{IncludeInitial: true}).
        Run(client.session)
    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }
    go func() {
        var change r.ChangeResponse
        for cursor.Next(&change) {
            result <- change
        }
    }()
    go func() {
        for {
            select {
                case <-stop:
                    cursor.Close()
                    return
                case change := <-result:
                    if change.NewValue != nil && change.OldValue == nil {
                        client.send <- Message{"channel add", change.NewValue}
                        fmt.Println("[Channel] Add")
                    }
            }
        }
    }()
}

func unsubscribeChannel(client *Client, data interface{}) {
    client.StopForKey(ChannelStop)
}

func userEdit(client *Client, data interface{}) {
    var user User
    err := mapstructure.Decode(data, &user)
    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }
    go func() {
        err := r.Table("user").
          Get(user.Id).
          Update(user).
          Exec(client.session)
        if err != nil {
            client.send <- Message{"error", err.Error()}
        }
        client.send <- Message{"user edit", user}
    }()
}

func subscribeUser(client *Client, data interface{}) {
    createAnonymousUser(client)
    stop := client.NewStopChannel(UserStop)
    result := make(chan r.ChangeResponse)
    cursor, err := r.Table("user").
      Changes(r.ChangesOpts{IncludeInitial: true}).
      Run(client.session)
    if err != nil {
        client.send <- Message{"Error", err.Error()}
    }
    go func() {
        var change r.ChangeResponse
        for cursor.Next(&change) {
            result <- change
        }
    }()
    go func(){
        for {
            select {
                case <-stop:
                    cursor.Close()
                    return
                case change := <-result:
                    if change.NewValue != nil && change.OldValue == nil {
                        client.send <- Message{"user add", change.NewValue}
                        fmt.Println("[User] Add")
                    } else if change.NewValue == nil && change.OldValue != nil {
                        client.send <- Message{"user remove", change.OldValue}
                        fmt.Println("[User] Remove")
                    } else if change.NewValue != nil && change.OldValue != nil {
                        client.send <- Message{"user edit", change.NewValue}
                        fmt.Println("[User] Edit")
                    }
            }
        }
    }()
}

func unsubscribeUser(client *Client, data interface{}) {
    client.StopForKey(UserStop)
}

func addMessage(client *Client, data interface{}) {
    var message MessageChannel
    err := mapstructure.Decode(data, &message)
    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }
    go func() {
        message.CreatedAt = time.Now()
        message.Author = getAuthorName(message.Author, client)
        err := r.Table("message").
          Insert(message).
          Exec(client.session)
        if err != nil {
            client.send <- Message{"error", err.Error()}
        }
    }()
}

func subscribeMessage(client *Client, data interface{}) {
    var subscriptionData map[string]string
    err := mapstructure.Decode(data, &subscriptionData)
    stop := client.NewStopChannel(MessageStop)
    result := make(chan r.ChangeResponse)

    query := r.Table("message").
        OrderBy(r.OrderByOpts{Index: r.Asc("createdAt")}).
        Filter(r.Row.Field("channelId").Eq(subscriptionData["channelId"]))

    response, err := query.Run(client.session)
    var messages []MessageChannel
    err = response.All(&messages)
    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }

    for _, message := range messages {
        client.send <- Message{"message add", message}
        fmt.Println("[Message] Add")
    }

    cursor, err := query.
        Changes(r.ChangesOpts{IncludeInitial: false}).
        Run(client.session)

    if err != nil {
        client.send <- Message{"error", err.Error()}
        return
    }
    go func(){
        var change r.ChangeResponse
        for cursor.Next(&change) {
            result <- change
        }
    }()
    go func(){
        for {
            select {
                case <-stop:
                    cursor.Close()
                    return
                case change := <-result:
                    if change.NewValue != nil && change.OldValue == nil {
                        client.send <- Message{"message add", change.NewValue}
                        fmt.Println("[Message] Add")
                    }
            }
        }
    }()
}

func unsubscribeMessage(client *Client, data interface{}) {
    client.StopForKey(MessageStop)
}

func createAnonymousUser(client *Client) {
    user := User{"", "Anonymous"}
    response, err := r.Table("user").
      Insert(user).
      RunWrite(client.session)
    if err != nil {
        client.send <- Message{"error", err.Error()}
    }
    user.Id = response.GeneratedKeys[0]
    client.userId = user.Id
    client.send <- Message{"user edit", user}
}

func getAuthorName(authorId string, client *Client) string {
    var author User
    response, err := r.Table("user").
      Get(authorId).
      Run(client.session)
    if err != nil {
        client.send <- Message{"error", err.Error()}
    }
    err = response.One(&author)
    return author.Name
}






