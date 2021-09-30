package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}
	l, err := cfg.Listen(ctx, "tcp", "127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	stdin := make(chan string, 1)
	entering := make(chan net.Conn, 1)
	leaving := make(chan net.Conn, 1)
	go watchStdin(ctx, stdin)
	go broadcastServerMsg(ctx, entering, leaving, stdin)
	log.Println("I'm started!")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			} else {
				wg.Add(1)
				entering <- conn
				go sendTime(ctx, conn, wg, leaving)
			}
		}
	}()

	<-ctx.Done()
	log.Println("done")
	l.Close()
	wg.Wait()
	log.Println("exit")

}

func sendTime(ctx context.Context, conn net.Conn, wg *sync.WaitGroup, leaving chan<- net.Conn) {
	defer wg.Done()
	defer conn.Close()
	// каждую 1 секунду отправлять клиентам текущее время сервера
	tck := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-tck.C:
			_, err := fmt.Fprintf(conn, "now: %s\n", t)
			if err != nil {
				leaving <- conn
				return
			}
		}
	}
}

func broadcastServerMsg(ctx context.Context, entering <-chan net.Conn, leaving <-chan net.Conn, stdin <-chan string) {
	clients := make(map[net.Conn]bool)
	for {
		select {
		case <-ctx.Done():
			return
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
		case msg := <-stdin:
			for cli := range clients {
				fmt.Fprintf(cli, "MESSAGE FROM SERVER: %s", msg)
			}
		}
	}
}

func watchStdin(ctx context.Context, stdin chan<- string) {
	in := bufio.NewReader(os.Stdin)
	for {
		serverMsg, err := in.ReadString('\n')
		if err != nil {
			log.Println(err)
			close(stdin)
			return
		}
		stdin <- serverMsg
	}
}
