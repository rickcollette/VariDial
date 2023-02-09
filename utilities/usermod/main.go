package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			active INT NOT NULL,
			username TEXT NOT NULL,
			number INT NOT NULL,
			password TEXT NOT NULL,
			level INT NOT NULL
		)
	`)
	return err
}

func addUser(db *sql.DB, active int, username string, number int, password string, level int) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO users (active, username, number, password, level)
		VALUES (?, ?, ?, ?, ?)
	`, active, username, number, password, level)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

func getUser(db *sql.DB, id int) (int, int, string, int, string, int, error) {
	var active int
	var username string
	var number int
	var password string
	var level int
	row := db.QueryRow(`
		SELECT active, username, number, password, level
		FROM users
		WHERE id = ?
	`, id)
	err := row.Scan(&active, &username, &number, &password, &level)
	if err != nil {
		return 0, 0, "", 0, "", 0, err
	}
	return id, active, username, number, password, level, nil
}

func updateUser(db *sql.DB, id int, active int, password string) (int64, error) {
	res, err := db.Exec(`
		UPDATE users
		SET active = ?, password = ?
		WHERE id = ?
	`, active, password, id)
	if err != nil {
		return 0, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affect, nil
}

func deleteUser(db *sql.DB, id int) (int64, error) {
	res, err := db.Exec(`
		DELETE FROM users
		WHERE id = ?
	`, id)
	if err != nil {
		return 0, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affect, nil
}

func listUsers(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT id, active, username, number, password, level
		FROM users
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var active int
		var username string
		var number int
		var password string
		var level int
		if err := rows.Scan(&id, &active, &username, &number, &password, &level); err != nil {
			return err
		}
		fmt.Printf("id: %d, active: %d, username: %s, number: %d, password: %s, level: %d\n", id, active, username, number, password, level)
	}
	return rows.Err()
}

func showHelp() {
	fmt.Println("Usage: go run main.go <command> <arguments>")
	fmt.Println("\nCommands:")
	fmt.Println("  create\t\tCreate the users table")
	fmt.Println("  add\t\t\tAdd a new user to the users table")
	fmt.Println("  get\t\t\tGet a user from the users table")
	fmt.Println("  update\t\tUpdate a user in the users table")
	fmt.Println("  delete\t\tDelete a user from the users table")
	fmt.Println("  list\t\tList all users in the users table")
	fmt.Println("  help\t\tShow this help message")
	fmt.Println("\nArguments:")
	fmt.Println("  <command> create\t\tNo arguments needed")
	fmt.Println("  <command> add\t\tActive (1/0), username, number, password, level")
	fmt.Println("  <command> get\t\tID of the user to get")
	fmt.Println("  <command> update\t\tID of the user to update, active (1/0), password")
	fmt.Println("  <command> delete\t\tID of the user to delete")
	fmt.Println("  <command> list\t\tNo arguments needed")
	fmt.Println("  <command> help\t\tNo arguments needed")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Expected at least one command-line argument")
	}
	db, err := sql.Open("sqlite3", "./crud.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	switch os.Args[1] {
	case "list":
		err := listUsers(db)
		if err != nil {
			log.Fatalf("Error listing users: %v", err)
		}
	case "help":
		showHelp()
	case "create":
		if len(os.Args) < 7 {
			log.Fatalf("Expected 6 arguments for creating user")
		}
		active, _ := strconv.Atoi(os.Args[2])
		username := os.Args[3]
		number, _ := strconv.Atoi(os.Args[4])
		password := os.Args[5]
		level, _ := strconv.Atoi(os.Args[6])
		lastID, err := addUser(db, active, username, number, password, level)
		if err != nil {
			log.Fatalf("Error adding user: %v", err)
		}
		fmt.Println("User added with ID:", lastID)
	case "read":
		if len(os.Args) < 3 {
			log.Fatalf("Expected 2 arguments for reading user")
		}
		id, _ := strconv.Atoi(os.Args[2])
		_, active, username, number, password, level, err := getUser(db, id)
		if err != nil {
			log.Fatalf("Error getting user: %v", err)
		}
		fmt.Println("User data:")
		fmt.Println("ID:", id)
		fmt.Println("Active:", active)
		fmt.Println("Username:", username)
		fmt.Println("Number:", number)
		fmt.Println("Password:", password)
		fmt.Println("Level:", level)
	case "update":
		if len(os.Args) < 5 {
			log.Fatalf("Expected 4 arguments for updating user")
		}
		id, _ := strconv.Atoi(os.Args[2])
		active, _ := strconv.Atoi(os.Args[3])
		password := os.Args[4]
		affect, err := updateUser(db, id, active, password)
		if err != nil {
			log.Fatalf("Error updating user: %v", err)
		}
		fmt.Println("Number of rows affected:", affect)
	case "delete":
		if len(os.Args) < 3 {
			log.Fatalf("Expected 2 arguments for deleting user")
		}
		id, _ := strconv.Atoi(os.Args[2])
		affect, err := deleteUser(db, id)
		if err != nil {
			log.Fatalf("Error deleting user: %v", err)
		}
		fmt.Println("Number of rows affected:", affect)
	default:
		log.Fatalf("Unsupported command-line argument")
		showHelp()
	}
}
