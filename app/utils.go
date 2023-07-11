package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func AutoSu() {
	if os.Geteuid() != 0 {
		exe, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		exeName := filepath.Base(exe)
		log.Println("The application " + exeName + " requires elevated privileges. Attempting to acquire them.")

		cmd := exec.Command("sudo", exe)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}
