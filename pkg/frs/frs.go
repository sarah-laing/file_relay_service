package frs

import (
	"log"
)

func Check(e error) {
	if e != nil {
		log.Println(e.Error())
	}
}

func FatalCheck(e error) {
	if e != nil {
		log.Fatalln(e.Error())
	}
}

// RelayCommand - Command may be "SEND" or "RECEIVE"
// Name depends on Command:
// SEND 	- file name to send
// RECEIVE 	- handle of a sent file waiting for a receiver
type RelayCommand struct {
	Command string
	Name    string
}

// RelayResponse - server response to a RelayCommand
// SEND 	- handle to receive sent file
// RECEIVE 	- name of file associated with requested handle
type RelayResponse struct {
	Response string
}
