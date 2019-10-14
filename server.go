package main

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	IP    string
	Port  int
	Store *store
}

func (srv *Server) Start() {
	addr := fmt.Sprintf("%s:%d", srv.IP, srv.Port)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalln("Error: ", err.Error())
	}

	defer listener.Close()

	log.Printf("Listening on %s", addr)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("Accepted connection from %v", conn.RemoteAddr())

		go func() {
			h := Handler{
				conn:  conn,
				store: srv.Store,
			}

			defer func() {
				log.Printf("closing connection from %v", h.conn.RemoteAddr())
				h.conn.Close()
			}()

			h.Handle()
		}()
	}
}

func (srv *Server) CloseAllConnections() {

}
