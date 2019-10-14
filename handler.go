package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

type Handler struct {
	conn  net.Conn
	store *store
}

func (h Handler) Handle() error {
	r := bufio.NewReader(h.conn)
	w := bufio.NewWriter(h.conn)

	scanr := bufio.NewScanner(r)

	for {
		scanned := scanr.Scan()
		if !scanned {
			if err := scanr.Err(); err != nil {
				log.Printf("%v(%v)", err, h.conn.RemoteAddr())
				return err
			}
			break
		}

		input := scanr.Text()

		if strings.HasPrefix(input, "PUT") {
			h.put(input, w)
		} else if strings.HasPrefix(input, "READ") {
			h.read(input, w)
		} else if strings.HasPrefix(input, "DELETE") {
			h.delete(input, w)
		} else if strings.HasPrefix(input, "QUIT") {
			h.conn.Close()
		} else {
			w.WriteString("UNKNOWN COMMAND\n")
		}

		w.Flush()
	}
	return nil
}

func (h Handler) put(input string, w *bufio.Writer) {
	args := strings.Split(input, " ")[1:]

	if len(args) != 3 {
		w.WriteString("INVALID PUT FORMAT (example: PUT foo 1 bar) \n")

		return
	}

	key := args[0]
	ttl, err := strconv.Atoi(args[1])

	if err != nil || ttl < 1 {
		w.WriteString("INVALID TTL VALUE\n")

		return
	}

	val := args[2]

	h.store.Put(key, val, ttl)

	w.WriteString("OK\n")
}

func (h Handler) read(input string, w *bufio.Writer) {
	args := strings.Split(input, " ")[1:]

	key := args[0]

	val, err := h.store.Get(key)

	if err != nil {
		w.WriteString("KEY NOT FOUND\n")

		return
	}

	w.WriteString(*val + "\n")
}

func (h Handler) delete(input string, w *bufio.Writer) {
	args := strings.Split(input, " ")[1:]

	key := args[0]

	h.store.Forget(key)
	w.WriteString("DELETED\n")
}
