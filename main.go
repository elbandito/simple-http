package main

import (
    . "github.com/elbandito/simple-http/web"
    "os"
    "os/signal"
    "syscall"
    "time"
    "context"
    "fmt"
)

func main() {
    server := NewServer()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        server.Start()
    }()

    // Wait for kill signal
    <-stop
    fmt.Println("Shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    server.Stop(ctx)
}
