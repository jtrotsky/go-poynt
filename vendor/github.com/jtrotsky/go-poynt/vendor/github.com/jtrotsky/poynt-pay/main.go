package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jtrotsky/poynt-pay/server"
)

func init() {
	// Create new or open existing logfile.
	var logFile *os.File
	if _, err := os.Stat("./poynt-pay.log"); os.IsNotExist(err) {
		// Logfile does not exist, create one.
		fmt.Println("Creating logfile")
		logFile, err = os.Create("./poynt-pay.log")
		if err != nil {
			fmt.Println("Error creating logfile:", err)
			panic(0)
		}
	} else {
		// Found existing logfile.
		fmt.Println("Logfile found")
		// Open file for writing.
		logFile, err = os.OpenFile("./poynt-pay.log", os.O_RDWR|os.O_APPEND,
			os.FileMode(0666))
		if err != nil {
			fmt.Println("Error opening logfile:", err)
		}
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Starting logfile.")
}

func main() {
	// Run the webserver on port 8000.
	server.Run()
}
