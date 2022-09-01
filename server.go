package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const password string = "qwerty"

func ValidatePassword(pwd string) bool {
	if pwd == password {
		return true
	}

	return false
}

func main() {
	const maxUsers = 2

	users := make(map[net.Conn]string)
	newConnection := make(chan net.Conn)
	addedUser := make(chan net.Conn)
	deadUsers := make(chan net.Conn)

	messages := make(chan string)

	server, err := net.Listen("tcp", ":6000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for{
		    conn, err := server.Accept()
		    if err != nil {
			panic(err)
		    }

		    if len(users) < maxUsers {
			newConnection <- conn
		    } else {
			io.WriteString(conn, "Server is full!")
		    }
		}
	}()

	for {
		select {
		case newConn := <-newConnection:
			go func(conn net.Conn) {
				reader := bufio.NewReader(conn)
				io.WriteString(conn, "Write the password: ")
				password, _ := reader.ReadString('\n')
				password = strings.Trim(password, "\r\n")
				if ValidatePassword(password) {
					io.WriteString(conn, "Write your name: ")
					username, _ := reader.ReadString('\n')
					username = strings.Trim(username, "\r\n")

					log.Println("Accepted new user: %s", username)

					messages <- fmt.Sprintf("Accepted new user: [%s]\n\n", username)

					users[conn] = username
					addedUser <- conn
				} else {
					io.WriteString(conn, "Error! Try again password\n\n")
					newConnection <- conn
				}

			}(newConn)
		case addUsr := <-addedUser:
			go func(conn net.Conn, username string) {
				reader := bufio.NewReader(conn)
				for {
					newMessage, err := reader.ReadString('\n')
					newMessage = strings.Trim(newMessage, "\r\n")
					if err != nil {
						break
					}

					messages <- fmt.Sprintf(">%s: %s \a\n\n", username, newMessage)
				}

				deadUsers <- conn
				messages <- fmt.Sprintf("%s disconnected", username)
			}(addUsr, users[addUsr])
		case msg := <-messages:
			for conn, _ := range users {
				go func(conn net.Conn, message string) {
					_, err := io.WriteString(conn, msg)
					if err != nil {
						deadUsers <- conn
					}
				}(conn, msg)
				log.Println("New message: %s", msg)
				log.Println("Message sent to %s users", len(users))
			}
		case conn := <-deadUsers:
			log.Println("%s disconnected", users[conn])
			delete(users, conn)
		}
	}

}
