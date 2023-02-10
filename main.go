package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/PatrickRudolph/telnet"
)

const (
	SystemName = "VariDial 1.0"
	Rodent     = "()"
	Normal     = "[]"
	CoSysop    = "<]"
	SysOp      = "<>"
)

type User struct {
	ID         int
	LineNumber int
	Username   string
	Number     int
	Password   string
	Level      int
}

var (
	clients          map[string]*telnet.Connection
	takenLineNumbers = make(map[int]bool)
	lineNumbers      = make(map[string]int)
)

func broadcastMessage(message string) {
	for _, client := range clients {
		client.Write([]byte(message + "\r\n"))
	}
}

func readLine(conn *telnet.Connection) (string, error) {
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(line)), nil
}

func processCommand(conn *telnet.Connection, username string, message string, userlevel int) {
	if message[0] == '/' {
		parts := strings.SplitN(message[1:], " ", 2)
		command := parts[0]
		var args string
		if len(parts) == 2 {
			args = parts[1]
		}
		switch command {
		case "q":
			conn.Close()
			delete(clients, username)
			broadcastMessage(fmt.Sprintf("\r\n%s left the chat\n", username))
		case "s":
			var usernames []string
			for u, ln := range lineNumbers {
				usernames = append(usernames, fmt.Sprintf("#%d[T1: %s: \r\n", ln, u))
			}
			conn.Write([]byte(fmt.Sprintf("\r\n->.\r\n    Online\r\n    ------------\r\n    %s\r\n", strings.Join(usernames, ", "))))
		case "p":
			split := strings.SplitN(args, " ", 2)
			if len(split) < 2 {
				conn.Write([]byte("Invalid private message format. Use /p# message.\r\n"))
				break
			}
			toLineNumberStr := split[0][1:]
			toLineNumber, err := strconv.Atoi(toLineNumberStr)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Invalid line number: %s\r\n", toLineNumberStr)))
				break
			}
			privateMessage := split[1]
			sendPrivateMessageByLineNumber(username, toLineNumber, privateMessage)
		case "i":
			conn.Write([]byte(fmt.Sprintf("\r\n->.\r\n    %s\r\n", SystemName)))
		case "?":
			conn.Write([]byte("\r\nCommands:\r\n  /q - Quit\r\n  /s - show online users\r\n  /p # message - Send private message\r\n  /i - system info\r\n  /? - Help\r\n"))
		default:
			conn.Write([]byte(fmt.Sprintf("Unknown command: %s\n", command)))
		}
	}
}

func login(number int, password string, db *sql.DB) *User {
	var user User
	err := db.QueryRow(`SELECT username, number, level FROM users WHERE number = ? AND password = ?`, number, password).Scan(&user.Username, &user.Number, &user.Level)
	if err != nil {
		return nil
	}
	return &user
}

func sendPrivateMessageByLineNumber(fromUsername string, toLineNumber int, message string) {
	toUsername := getUsernameByLineNumber(toLineNumber)
	if toUsername == "" {
		return
	}
	toConn, ok := clients[toUsername]
	if !ok {
		return
	}
	toConn.Write([]byte(fmt.Sprintf("\r\nP[T1:%s] ( %s )\r\n", fromUsername, message)))
}

func getUsernameByLineNumber(toLineNumber int) string {
	for username, lineNumber := range lineNumbers {
		if lineNumber == toLineNumber {
			return username
		}
	}
	return ""
}

func getNextAvailableLineNumber() int {
	for i := 1; i <= 99; i++ {
		if !takenLineNumbers[i] {
			takenLineNumbers[i] = true
			return i
		}
	}
	return -1
}

func showFile(conn *telnet.Connection, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error opening file: %s\r\n", filename)))
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		conn.Write([]byte(fmt.Sprintf("%s\r\n", line)))
	}
}

func formMessage(line int, uname string, message string, ulevel int) string {
	if ulevel == 0 {
		return fmt.Sprintf("#%d( %s ): %s\r\n", line, uname, message)
	} else if ulevel == 1 {
		return fmt.Sprintf("#%d[ %s ): %s\r\n", line, uname, message)
	} else if ulevel == 2 {
		return fmt.Sprintf("#%d< %s ): %s\r\n", line, uname, message)
	} else if ulevel == 3 {
		return fmt.Sprintf("#%d< %s ]: %s\r\n", line, uname, message)
	} else if ulevel == 4 {
		return fmt.Sprintf("#%d< %s >: %s\r\n", line, uname, message)
	}
	return ("Error: No user Level")
}

func freeLineNumber(lineNumber int) {
	takenLineNumbers[lineNumber] = false
}

func handleConnection(conn *telnet.Connection, db *sql.DB) {
	showFile(conn, "login.txt")
	conn.Write([]byte(SystemName + "\r\n"))
	conn.Write([]byte("Enter your number: "))
	numberStr, err := readLine(conn)
	if err != nil {
		conn.Close()
		return
	}
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		conn.Close()
		return
	}
	conn.Write([]byte("\r\nEnter your password: "))
	password, err := readLine(conn)
	if err != nil {
		conn.Close()
		return
	}
	user := login(number, password, db)
	if user == nil {
		conn.Close()
		return
	}
	lineNumber := getNextAvailableLineNumber()
	if lineNumber == -1 {
		conn.Close()
		return
	}
	conn.Write([]byte("\r\n/? for help\r\n"))

	user.LineNumber = lineNumber
	clients[user.Username] = conn
	lineNumbers[user.Username] = lineNumber
	broadcastMessage(fmt.Sprintf("\r\n->\r\n   +#%d:%s\r\n", lineNumber, user.Username))
	for {
		message, err := readLine(conn)
		if err != nil {
			break
		}
		if len(message) > 0 && message[0] == '/' {
			processCommand(conn, user.Username, message, user.Level)
		} else {
			sendAllMessage := formMessage(lineNumber, user.Username, message, user.Level)
			broadcastMessage(sendAllMessage)
		}
	}
	conn.Close()
	delete(clients, user.Username)
	freeLineNumber(lineNumber)
	delete(lineNumbers, user.Username)
	broadcastMessage(fmt.Sprintf("\r\n-> \r\n   %s left the chat\r\n", user.Username))
}

func main() {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	clients = make(map[string]*telnet.Connection)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Chat server started on port 8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(telnet.NewConnection(conn, []telnet.Option{}), db)
	}
}
