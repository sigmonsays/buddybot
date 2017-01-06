package main

import (
	"fmt"
	"os"
)

func CheckError(err error, s string, args ...interface{}) {
	if err == nil {
		return
	}
	fmt.Printf(s+"\n", args...)
	os.Exit(1)
}

func StartupError(s string, args ...interface{}) {
	fmt.Printf(s+"\n", args...)
	os.Exit(1)
}
