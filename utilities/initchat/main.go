package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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

func main() {
	var user User

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username (less than 14 characters): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if len(username) > 14 {
		fmt.Println("Error: username must be less than 14 characters")
		return
	}
	user.Username = username

	fmt.Print("Enter number (3 digits): ")
	numberStr, _ := reader.ReadString('\n')
	numberStr = strings.TrimSpace(numberStr)
	if len(numberStr) != 3 {
		fmt.Println("Error: number must be 3 digits")
		return
	}
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		fmt.Println("Error: invalid number")
		return
	}
	user.Number = number

	fmt.Print("Enter password (8 characters, less than 13): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if len(password) < 8 || len(password) > 13 {
		fmt.Println("Error: password must be 8 characters and less than 13")
		return
	}
	user.Password = password

	fmt.Print("Enter level (0 to 4): ")
	levelStr, _ := reader.ReadString('\n')
	levelStr = strings.TrimSpace(levelStr)
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 0 || level > 4 {
		fmt.Println("Error: invalid level")
		return
	}
	user.Level = level

	fmt.Println("User struct initialized:", user)

	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	err = initializeDB(db)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	err = saveUser(db, user)
	if err != nil {
		fmt.Println("Error saving user:", err)
		return
	}
	fmt.Println("User saved successfully.")
}

func saveUser(db *sql.DB, user User) error {
	query := `INSERT INTO users (username, number, password, level, channel) VALUES (?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.Username, user.Number, user.Password, user.Level, user.Channel)
	if err != nil {
		return err
	}
	return nil
}

func initializeDB(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(14) NOT NULL,
			number INTEGER NOT NULL,
			password VARCHAR(13) NOT NULL,
			level INTEGER NOT NULL,
			channel INTEGER NOT NULL
			);
			`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
