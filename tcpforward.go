package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	localAddr  = flag.String("l", ":9999", "host:port to listen on")
	remoteAddr = flag.String("r", ":9200", "host:port to forward to")
	prefix     = flag.String("p", "tcpforward: ", "String to prefix log output")
)

func forward(conn net.Conn) {
	client, err := net.Dial("tcp", *remoteAddr)
	if err != nil {
		log.Printf("Dial failed: %v", err)
		defer conn.Close()
		return
	}
	longBuf := make([]byte, 256)

	log.Printf("Connected to %v\n, read greeting", client.RemoteAddr())
	if _, err := io.ReadAtLeast(client, longBuf, 30); err != nil {
		fmt.Println("error:", err)
	}
	log.Printf("received: %s", longBuf)

	log.Printf("Sending EHLO test.com to %v\n", client.RemoteAddr())
	client.Write([]byte("EHLO test.com\r\n"))
	log.Printf("reading from client..")
	if _, err := io.ReadAtLeast(client, longBuf, 60); err != nil {
		fmt.Println("error:", err)
	}
	log.Printf("received: %s", longBuf)

	log.Printf("Sending TLSSTART to %v\n", client.RemoteAddr())
	client.Write([]byte("STARTTLS\r\n"))
	log.Printf("reading from client..")
	if _, err := io.ReadAtLeast(client, longBuf, 10); err != nil {
		fmt.Println("error:", err)
	}
	log.Printf("received: %s", longBuf)

	log.Printf("Forwarding from %v to %v\n", conn.LocalAddr(), client.RemoteAddr())
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func main() {
	flag.Parse()
	log.SetPrefix(*prefix + ": ")

	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}
		log.Printf("Accepted connection from %v\n", conn.RemoteAddr().String())
		go forward(conn)
	}
}
