package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sarah-laing/file_relay_service/pkg/frs"
)

type connectionInfo struct {
	sender   *net.Conn
	fileName string
}

type connectionMap map[string]connectionInfo

var cm connectionMap
var cmMutex sync.RWMutex

func handleSend(tcpConnection net.Conn, command frs.RelayCommand) {

	// Add a connection record for this file sender to the map,
	connectionHandle := fmt.Sprintf("%X", &tcpConnection)
	cmMutex.Lock()
	cm[connectionHandle] = connectionInfo{&tcpConnection, command.Name}
	cmMutex.Unlock()

	// and return handle to sender.
	rr := frs.RelayResponse{connectionHandle}
	enc := json.NewEncoder(tcpConnection)
	encErr := enc.Encode(rr)
	frs.Check(encErr)
}

func handleReceive(tcpConnection net.Conn, command frs.RelayCommand) {

	// Find sender's connection record by handle in the connections map.
	defer tcpConnection.Close()
	senderConnectionHandle := command.Name
	cmMutex.RLock()
	senderConnectionInfo, haveConnection := cm[senderConnectionHandle]
	cmMutex.RUnlock()

	if haveConnection {

		// Send RECEIVE response to tell receiver name of file to receive.
		rr := frs.RelayResponse{senderConnectionInfo.fileName}
		enc := json.NewEncoder(tcpConnection)
		encErr := enc.Encode(rr)
		frs.Check(encErr)

		// Relay file content from sender's connection to receiver's.
		time.Sleep(100 * time.Millisecond) // Delay while receiver prepares to receive file - fragile and slow!
		_, writeErr := io.Copy(tcpConnection, *senderConnectionInfo.sender)
		frs.Check(writeErr)
		tcpConnection.Close()
		(*senderConnectionInfo.sender).Close()

		// After finishing transfer and closing connections, remove the sender's record from the connections map.
		cmMutex.Lock()
		delete(cm, senderConnectionHandle)
		cmMutex.Unlock()

	} else {
		log.Println("Unable to find sender for handle: ", senderConnectionHandle)
	}
}

func acceptConnection(tcpConnection net.Conn) {

	// Decode client's command,
	var command frs.RelayCommand
	dec := json.NewDecoder(tcpConnection)
	decErr := dec.Decode(&command)
	frs.Check(decErr)

	// and asynch dispatch to appropriate command handler.
	switch command.Command {
	case "SEND":
		handleSend(tcpConnection, command)
	case "RECEIVE":
		handleReceive(tcpConnection, command)
	default:
		tcpConnection.Close() // drop connection if client does not open with SEND or RECEIVE command.
	}
}

func main() {

	// Process command line arguments.
	if len(os.Args) != 2 {
		fmt.Println("usage: ", os.Args[0], " [listen-port]")
		os.Exit(1)
	}
	listenPort := os.Args[1]

	// Initialize a globally shared map of connected clients.
	cm = make(connectionMap)

	// Listen for new client connections.
	tcpListener, listenError := net.Listen("tcp", listenPort)
	frs.Check(listenError)

	for {
		tcpConnection, connectionError := tcpListener.Accept()
		go frs.Check(connectionError)      // asynch any console i/o from inside listen loop
		go acceptConnection(tcpConnection) // set up new connection asynch
	}
}
