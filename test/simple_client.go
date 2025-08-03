package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	commands := []string{
		"*1\r\n$4\r\nPING\r\n",
		"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		"*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
		"*2\r\n$6\r\nEXISTS\r\n$3\r\nfoo\r\n",
		"*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n",
	}

	for i, cmd := range commands {
		fmt.Printf("Command %d: %s", i+1, cmd)
		
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}

		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}

		fmt.Printf("Response: %s\n", response[:n])
		fmt.Println("---")
	}
}