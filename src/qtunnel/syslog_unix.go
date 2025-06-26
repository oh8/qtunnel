//go:build !windows
// +build !windows

package main

import (
	"log"
	"log/syslog"
)

func setupSyslog() error {
	w, err := syslog.New(syslog.LOG_INFO, "qtunnel")
	if err != nil {
		return err
	}
	log.SetOutput(w)
	return nil
}