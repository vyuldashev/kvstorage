package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "listen ip address")
	port := flag.Int("port", 12345, "listen port")
	filePath := flag.String("file", "./kvstorage.json", "storage file")

	flag.Parse()

	s := NewStore(filePath)

	srv := Server{
		IP:    *ip,
		Port:  *port,
		Store: s,
	}

	go srv.Start()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c

	s.Close()
}
