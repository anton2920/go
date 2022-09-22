package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	proto, addr := "tcp", "0.0.0.0:1500"
	listener, err := net.Listen(proto, addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s://%s", proto, addr)

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("Accepted from %s", conn.RemoteAddr().String())
		go handleConn(conn)
	}
}

type client struct {
	pipe chan<- string
	name string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func broadcaster() {
	clients := make(map[client]struct{})
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.pipe <- msg
			}
		case cli := <-entering:
			if len(clients) == 0 {
				cli.pipe <- "You are the only one in the chat"
			} else {
				cli.pipe <- "Currently online: "
				for user := range clients {
					cli.pipe <- user.name + " "
				}
			}
			clients[cli] = struct{}{}
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.pipe)
		}
	}
}

func handleConn(conn net.Conn) {
	/* Get name from client */
	input := bufio.NewScanner(conn)
	fmt.Fprint(conn, "Enter your name: ") /* NOTE: ignoring network errors */
	input.Scan()
	who := input.Text()
	who += " (" + conn.RemoteAddr().String() + ")"

	ch := make(chan string) /* outgoing client messages */
	go clientWriter(conn, ch)

	ch <- "You are " + who

	messages <- who + " has arrived"
	entering <- client{ch, who}

	/* Read from network peer and broadcast to others.
	 * This blocks until client it disconnected.
	 */
	for input.Scan() {
		message := who + ": " + input.Text()
		log.Println(message)
		messages <- message
	}
	/* NOTE: ignoring potential errors from input.Err() */

	leaving <- client{ch, who}
	messages <- who + " has left"
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s disconnected", who)
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, _ = fmt.Fprintln(conn, msg) /* NOTE: ignoring network errors */
	}
}
