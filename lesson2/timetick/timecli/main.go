package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT) //nolint

	d := net.Dialer{
		Timeout:   time.Second,
		KeepAlive: time.Minute,
	}
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(os.Stdout, conn); err == nil {
		fmt.Println("Connection closed")
	} else if err != nil {
		fmt.Printf("Error occured: %v\n", err)
	}
}
