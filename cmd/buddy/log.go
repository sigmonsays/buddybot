package main

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("buddy", func(newlog gologging.Logger) { log = newlog })
}
