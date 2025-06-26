//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
)

func setupSyslog() error {
	// Windows doesn't support syslog, fallback to file logging
	logFile, err := os.OpenFile("qtunnel.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.Println("Windows: Using file logging instead of syslog")
	return nil
}