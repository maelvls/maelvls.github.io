---
title: Writing a good Go SDK
description: ""
date: 2020-05-08T15:00:27.000+02:00
url: "/writing-a-good-go-sdk"
images:
  - writing-a-good-go-sdk/cover-writing-a-good-go-sdk.png
tags: []
author: Maël Valais
devtoId: "0"
devtoPublished: false
draft: true
---

The typical way of building a client in Go is to pass a `*http.Client` in order to let the caller set timeouts and so on. See [godo](https://github.com/digitalocean/godo/blob/97ac73b1d53e23afa2700e9f97d5eeb1f3641e3f/godo.go#L153-L188) (Digital Ocean API client) for a good way of doing it.

Another good practice is to avoid hidden network calls. For example, most Go users do not expect a `NewClient` function to do a network call. Prefer using a dedicated function such as `client.Auth`.

In the \[docker\]\[\] client

\[docker\]: https://github.com/moby/moby/blob/f5bb374a0c6260721ac551b232e8eac02b7d2674/client/client.go#L146-L150)

> `NewClient` should probably only take an `*http.Client` and work from there.

I think `NewClient` should only take a token and we also offer the user a way to give the token:

```go
func NewClient(c *http.Client, token string) Client {
}

func GetToken(
```

**Content:**

1. [Unleash](#unleash)
2. [Stripe](#stripe)
3. [Redis](#redis)
4. [Saltstack client (r3labs/go-salt)](#saltstack-client-r3labsgo-salt)
5. [Docker Engine API client](#docker-engine-api-client)
6. [Heroku](#heroku)
7. [Slack](#slack)
8. [Scaleway](#scaleway)

## Unleash

```go
import "github.com/Unleash/unleash-client-go"

func init() {
    unleash.Initialize(
        unleash.WithListener(&unleash.DebugListener{}),
        unleash.WithAppName("my-application"),
        unleash.WithUrl("http://unleash.herokuapp.com/api/"),
    )
}

func main() {
    // Example of API call.
    unleash.IsEnabled("app.ToggleX")
}
```

## Stripe

```go
import (
    "net/http"
    stripe "github.com/stripe/stripe-go"
    "github.com/stripe/stripe-go/client"
)
func main() {
    httpClient := &http.Client{Timeout: 10 * time.Second}
    backends := stripe.NewBackends(httpClient)
    client := client.NewClient("token", backends)
}
```

## Redis

Doesn't use HTTP anyway. But you can pass a Dialer function that creates a `net.Conn`. Since Redis is using TCP directly, the Redis client has to offer everything itself (e.g. DialTimeout).

```go
import "github.com/go-redis/redis"

func main() {
    opts := &redis.Options{Addr: "localhost:6379",Password: ""}
    client := redis.NewClient(opts)
}
```

## Saltstack client (r3labs/go-salt)

The [NewClient](https://github.com/r3labs/go-salt/blob/e6bcc1482122fbfbb41c8c5d7204e067e97a4266/client.go#L18) does a network call. That's pretty uncommon to do that in Go.

There is no way to pass a `*http.Client` in `NewClient`. But you can by going through some hoops.

Also, the authentication should create a new `*http.Client` instead of mutating the `conn.AuthToken` value.

```go
import (
    "net/http"
    "github.com/r3labs/go-salt"
)

func main() {
    // Using NewClient
    opts := salt.Config{Host: "", Username: "", Password: ""}
    client, err := salt.NewClient(opts)

    // With a timeout using *http.Client
    httpClient := &http.Client{Timeout: 10 * time.Second}
    conn := salt.Connector{Client: httpClient, AuthToken: "token"}
    conn.Authenticate()
    client := &salt.Client{Connector: conn}
    job, err := client.Job("foo")
}
```

## Docker Engine API client

The [dockerengine.NewClient](https://github.com/moby/moby/blob/f5bb374a0c6260721ac551b232e8eac02b7d2674/client/client.go#L119) does return an error, which is kind of unusual for NewClient in Go. The only thing they do is to check that the transport field (`http.Transport`) is in fact a `http.RoundTripper`.

```go
import (
    "net/http"
    "github.com/moby/moby/client"
)
func main() {
    httpClient := &http.Client{Timeout: 10 * time.Second}
    client, err := client.NewClientWithOpts(client.WithHTTPClient(httpClient))
    if err != nil {
        panic(err)
    }
```

## Heroku

Very clean! Many globals, but it's just for ease of use: everything can be used without using the global state (`heroku.DefaultTransport`). [https://github.com/heroku/heroku-go/blob/master/v5/transport.go](https://github.com/heroku/heroku-go/blob/master/v5/transport.go)

```go
import (
    heroku "github.com/heroku/heroku-go/v5"
)
func main() {
    c := &http.Client{
        Timeout: 10 * time.Second,
        Transport: heroku.Transport{BearerToken: "token"},
    }
    h := heroku.NewService(c)
    addons, err := h.AddOnList(context.TODO(), &heroku.ListRange{Field: "name"})
}
```

## Slack

Very good!!! [https://github.com/nlopes/slack/blob/e5749f13b5af3c139165ab4180a95bb06a60128b/slack.go#L68](https://github.com/nlopes/slack/blob/e5749f13b5af3c139165ab4180a95bb06a60128b/slack.go#L68)

```go
import (
    "github.com/nlopes/slack"
)

func main() {
    httpClient := &http.Client{Timeout: 10 * time.Second}
    api := slack.New("token", slack.OptionHTTPClient(httpClient))
    groups, err := api.GetGroups(false)
}
```

## Scaleway

[https://github.com/scaleway/scaleway-sdk-go](https://github.com/scaleway/scaleway-sdk-go)

And they use an HTTP recorder "à la Jest Snapshot"!
