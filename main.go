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
	Channel    int
}

var (
	clients          map[string]*telnet.Connection
	takenLineNumbers = make(map[int]bool)
	lineNumbers      = make(map[string]User)
)

func broadcastMessage(message string, sender User) {
	for username, conn := range clients {
		if user, ok := lineNumbers[username]; ok && user.Channel == sender.Channel {
			conn.Write([]byte(message + "\r\n"))
		}
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
func sendAll(message string) {
	for _, conn := range clients {
		conn.Write([]byte(message + "\r\n"))
	}
}

/* fix this with the following:
input := "p12 message stuff"

// Split the input string into two parts, separated by a space
parts := strings.SplitN(input, " ", 2)

// The first part will be the first character and the number
firstPart := parts[0]

// The second part will be the rest of the message
secondPart := parts[1]

// Extract the number from the first part
number, err := strconv.Atoi(firstPart[1:])
if err != nil {
    // Handle the error, for example by logging it
    fmt.Println("Error converting number:", err)
}

fmt.Println("Number:", number)
fmt.Println("Message:", secondPart)
*/

func processCommand(conn *telnet.Connection, userchannel int, username string, message string, userlevel int, db *sql.DB) {
	if message[0] == '/' {
		parts := strings.SplitN(message[1:], " ", 2)
		command := parts[0]
		var args string
		fmt.Println("Command:", command)
		if len(parts) == 2 {
			args = parts[1]
		}
		switch command {
		case "q":
			conn.Close()
			delete(clients, username)
			sendAll(fmt.Sprintf("\r\n%s left the chat\n", username))
		case "s":
			var usernames []string
			for username, user := range lineNumbers {
				usernames = append(usernames, fmt.Sprintf("#%d[T1: %s: \r\n", user.LineNumber, username))
			}
			conn.Write([]byte(fmt.Sprintf("\r\n->.\r\n    Online\r\n    ------------\r\n    %s\r\n", strings.Join(usernames, ", "))))
		case "p":
			split := strings.SplitN(args, " ", 2)
			if len(split) < 2 {
				conn.Write([]byte("Invalid private message format. Use /p# message.\r\n"))
				break
			}
			toLineNumberStr := split[0]
			toLineNumber, err := strconv.Atoi(toLineNumberStr)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Invalid line number: %s\r\n", toLineNumberStr)))
				break
			}
			privateMessage := split[1]
			fmt.Println("Sending private message from", username, "to", toLineNumber, ":", privateMessage)
			sendPrivateMessageByLineNumber(userchannel, username, toLineNumber, privateMessage)

		case "t":
			channelStr := strings.TrimSpace(args)
			channel, err := strconv.Atoi(channelStr)
			if err != nil || channel < 1 || channel > 4 {
				conn.Write([]byte("Error: invalid channel. Must be a number between 1 and 4.\r\n"))
				break
			}
			user, ok := lineNumbers[username]
			if !ok {
				conn.Write([]byte("Error: user not found.\r\n"))
				break
			}
			user.Channel = channel
			conn.Write([]byte(fmt.Sprintf("Changed to channel %d.\r\n", channel)))
			updateUser(username, channel, db)
		case "i":
			conn.Write([]byte(fmt.Sprintf("\r\n->.\r\n    %s\r\n", SystemName)))
		case "?":
			conn.Write([]byte("\r\nCommands:\r\n  /q - Quit\r\n  /s - show online users\r\n  /p # message - Send private message\r\n  /i - system info\r\n  /? - Help\r\n"))
		default:
			conn.Write([]byte(fmt.Sprintf("Unknown command: %s\n", command)))
		}
	}
}

func updateUser(username string, channel int, db *sql.DB) error {
	query := `UPDATE users SET channel = ? WHERE number, username = ?,?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(channel, username)
	if err != nil {
		return err
	}
	return nil
}

func login(number int, password string, db *sql.DB) *User {
	var user User
	err := db.QueryRow(`SELECT username, number, level FROM users WHERE number = ? AND password = ?`, number, password).Scan(&user.Username, &user.Number, &user.Level)
	if err != nil {
		return nil
	}
	user.Channel = 1
	return &user
}

func sendPrivateMessageByLineNumber(fromChannel int, fromUsername string, toLineNumber int, message string) {
	toUsername := getUsernameByLineNumber(toLineNumber)
	if toUsername == "" {
		return
	}
	toConn, ok := clients[toUsername]
	if !ok {
		return
	}
	toConn.Write([]byte(fmt.Sprintf("\r\nP[T%d:%s] ( %s )\r\n", fromChannel, fromUsername, message)))
}

func getUsernameByLineNumber(toLineNumber int) string {
	for username, user := range lineNumbers {
		if user.LineNumber == toLineNumber {
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

func formMessage(line int, channel int, uname string, message string, ulevel int) string {
	if ulevel == 0 {
		return fmt.Sprintf("#%d(T%d:%s ): %s\r\n", line, channel, uname, message)
	} else if ulevel == 1 {
		return fmt.Sprintf("#%d[T%d:%s ): %s\r\n", line, channel, uname, message)
	} else if ulevel == 2 {
		return fmt.Sprintf("#%d<T%d:%s ): %s\r\n", line, channel, uname, message)
	} else if ulevel == 3 {
		return fmt.Sprintf("#%d<T%d:%s ]: %s\r\n", line, channel, uname, message)
	} else if ulevel == 4 {
		return fmt.Sprintf("#%d<T%d:%s >: %s\r\n", line, channel, uname, message)
	}
	return ("Error: No user Level")
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
	lineNumbers[user.Username] = *user
	broadcastMessage(fmt.Sprintf("\r\n->\r\n +#%d:%s\r\n", lineNumber, user.Username), *user)
	for {
		message, err := readLine(conn)
		if err != nil {
			break
		}
		if len(message) > 0 && message[0] == '/' {
			processCommand(conn, user.Channel, user.Username, message, user.Level, db)
		} else {
			sendAllMessage := formMessage(lineNumber, user.Channel, user.Username, message, user.Level)
			broadcastMessage(sendAllMessage, *user)
		}
	}
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
