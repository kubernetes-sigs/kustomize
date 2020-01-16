package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("The usage of the command is: \n\t%s <log-file.json>", os.Args[0])
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	closeFile := func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}
	defer closeFile(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var entry interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Printf("failed to unmarshal a log entry: %s\n", line)
		}

		m := entry.(map[string]interface{})
		if payload, ok := m["textPayload"]; ok {
			// use fmt.Printf here instead of log.Printf to avoid the time and code location info the log package provides
			fmt.Printf("%s", payload)
		} else {
			log.Printf("the log entry does not have the `textPayload` field: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
