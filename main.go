package main

import (
    . "github.com/elbandito/simple-http/web"
    "os"
    "os/signal"
    "syscall"
    "fmt"
)

func main() {
    server := NewServer()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        server.Start()
    }()

    <-stop
    server.Stop()

    fmt.Println("server has gracefully shutdown")
    fmt.Println()
}
