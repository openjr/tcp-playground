package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	address := "localhost:6001"
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	shutdown := make(chan struct{}, 1)

	go func(ctx context.Context) {
		log.Println("Listening for incoming connections")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := socket.AcceptTCP()
				if err != nil {
					log.Println("error accepting TCP connection", err)
					continue
				}

				go handleConn(ctx, conn)
			}

		}
	}(ctx)

	go startSignalHandler(shutdown)

	// wait until shutdown signal comes
	<-shutdown
	cancelFunc()
	log.Println("waiting to shutdown the server")
	time.Sleep(1 * time.Second)

}

func startSignalHandler(shutdown chan struct{}) {
	log.Println("Waiting for shutdown signals")
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for _ = range signalChan {
		shutdown <- struct{}{}
		return
	}
}

func handleConn(ctx context.Context, conn *net.TCPConn) {
	defer conn.Close()
	readBuffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			log.Println("Closing connection")
			return
		default:
			err := conn.SetReadDeadline(time.Now().Add(time.Second))
			if err != nil {
				log.Println("Couldn't set read deadline", err)
				return
			}

			n, err := conn.Read(readBuffer)
			if err != nil {
				var opErr *net.OpError
				if errors.As(err, &opErr) && opErr.Temporary() {
					log.Println("Temporary error, will retry", err)
					continue
				}
				log.Println("error reading from connection", err)
				return
			}
			log.Println("read data: ", string(readBuffer[:n]))
			n, err = conn.Write(readBuffer[:n])

			if rand.Int63n(100)+1 > 70 {
				log.Println("closing the connection on the write side")
				_ = conn.CloseWrite()
			}
		}
	}
}
