package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sarah-laing/filerelayservice/pkg/frs"
)

func main() {

	// Process command line arguments.
	if len(os.Args) != 4 {
		fmt.Println("usage: ", os.Args[0], " [relay-service] [transfer-handle] [output-path]")
		os.Exit(1)
	}
	relayAddr := os.Args[1]
	transferHandle := strings.TrimSpace(os.Args[2])
	receivedFilesPath := os.Args[3]

	// connect to relay service
	tcpConnection, dialError := net.Dial("tcp", relayAddr)
	defer tcpConnection.Close()
	frs.Check(dialError)

	// Send RECEIVE command with transferHandle.
	enc := json.NewEncoder(tcpConnection)
	encErr := enc.Encode(frs.RelayCommand{Command: "RECEIVE", Name: transferHandle})
	frs.Check(encErr)

	// Receive and decode relay service RECEIVE command response.
	var rr frs.RelayResponse
	dec := json.NewDecoder(tcpConnection)
	decErr := dec.Decode(&rr)
	frs.Check(decErr)

	// Extract file name from RECEIVE command response,
	outFileName := filepath.Join(receivedFilesPath, rr.Response)
	log.Println("Receiving: ", outFileName)

	// and create local file in the output directory.
	outFile, fileError := os.Create(outFileName)
	frs.Check(fileError)
	defer outFile.Close()

	// Download file.
	nWritten, writeError := io.Copy(outFile, tcpConnection)
	log.Println("Received: ", nWritten, writeError)
	syncErr := outFile.Sync()
	frs.Check(syncErr)
	outFile.Close()
}
