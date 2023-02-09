package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func readIniFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	settings := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := strings.TrimSpace(line[equal+1:])
				settings[key] = value
			}
		} else {
			key := strings.TrimSpace(line)
			settings[key] = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}

func writeIniFile(filename string, settings map[string]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for key, value := range settings {
		if value != "" {
			_, err := fmt.Fprintf(w, "%s = %s\n", key, value)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintln(w, key)
			if err != nil {
				return err
			}
		}
	}
	return w.Flush()
}

func modifySetting(settings map[string]string, key string, value string) {
	settings[key] = value
}

func showHelp() {
	fmt.Println("Usage: chat-ini [OPTION]... [KEY=VALUE]...")
	fmt.Println("Generate or modify an ini file for the chat system.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println(" -f, --file\t\tspecify the ini file to be generated or modified")
	fmt.Println(" -h, --help\t\tshow this help message")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println(" chat-ini -f chat.ini AES_KEY_HERE=newkey PORT=2021")
}

func main() {
	var filename string
	var showHelpFlag bool
	settings, err := readIniFile(filename)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-f, --file":
			i++
			if i < len(os.Args) {
				filename = os.Args[i]
			} else {
				fmt.Println("ERROR: filename not specified after -f or --file option")
				os.Exit(1)
			}
		case "-h", "--help":
			showHelpFlag = true
		default:
			if strings.Contains(arg, "=") {
				if equal := strings.Index(arg, "="); equal >= 0 {
					if key := strings.TrimSpace(arg[:equal]); len(key) > 0 {
						value := strings.TrimSpace(arg[equal+1:])
						modifySetting(settings, key, value)
					}
				}
			} else {
				fmt.Printf("ERROR: Invalid option or argument: %s\n", arg)
				os.Exit(1)
			}
		}
	}
	if showHelpFlag || filename == "" {
		showHelp()
		os.Exit(0)
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.Contains(arg, "=") {
			if equal := strings.Index(arg, "="); equal >= 0 {
				if key := strings.TrimSpace(arg[:equal]); len(key) > 0 {
					value := strings.TrimSpace(arg[equal+1:])
					modifySetting(settings, key, value)
				}
			}
		}
	}

	err = writeIniFile(filename, settings)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
}
