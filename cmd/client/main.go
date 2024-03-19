package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", "localhost:6001")
	if err != nil {
		log.Panicln("couldn't resolve address", err)
	}
	tcpConn, err := net.DialTCP("tcp4", nil, addr)

	recBuff := make([]byte, 1024)
	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("%d", i)
		_, err = tcpConn.Write([]byte(msg))
		if err != nil {
			log.Println("error writing to connection", err)
			break
		}
		log.Println("message sent", msg)

		n, err := tcpConn.Read(recBuff)
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Temporary() {
				log.Println("Temporary error, will retry")
				continue
			}
			log.Println("error reading the connection", err)
			_ = tcpConn.Close()
			break
		}
		log.Println("message received", string(recBuff[:n]))
		log.Println("sleeping for 5 seconds")
		time.Sleep(5 * time.Second)
	}
	log.Println("exiting!")
}
