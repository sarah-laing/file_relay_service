package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/sarah-laing/filerelayservice/pkg/frs"
)

func main() {

	// Process command line arguments.
	if len(os.Args) != 3 {
		fmt.Println("usage: ", os.Args[0], " [relay-service] [file-name]")
		os.Exit(1)
	}
	relayServiceIP := os.Args[1]
	inFileName := os.Args[2]

	// Open file to send.
	inFile, fileError := os.Open(inFileName)
	frs.FatalCheck(fileError)
	defer inFile.Close()

	// Connect to relay service.
	tcpConnection, dialError := net.Dial("tcp", relayServiceIP)
	frs.FatalCheck(dialError)
	defer tcpConnection.Close()

	// Send SEND command to give file name to relay service and request a transfer handle.
	//tcpConnection.SetReadDeadline(time.Now().Add(1000 * time.Second))
	enc := json.NewEncoder(tcpConnection)
	encErr := enc.Encode(frs.RelayCommand{"SEND", inFileName})
	frs.FatalCheck(encErr)

	// Decode SEND command response from relay service.
	//tcpConnection.SetReadDeadline(time.Now().Add(1000 * time.Second))
	var rr frs.RelayResponse
	dec := json.NewDecoder(tcpConnection)
	decErr := dec.Decode(&rr)
	frs.FatalCheck(decErr)

	// Display transfer handle on stdout.
	fmt.Println(rr.Response) // must happen _before_ transfer, otherwise breaks large file transfers.

	// Transfer file data.
	nWritten, writeError := io.Copy(tcpConnection, inFile)

	log.Println("sender: ", nWritten, writeError) // Also display transfer diagnostics.

	// Close connection - signals relay service that transfer is complete.
	syncErr := tcpConnection.Close()
	frs.FatalCheck(syncErr)
}
