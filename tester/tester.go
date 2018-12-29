package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/sarah-laing/filerelayservice/pkg/frs"
)

const relayServiceIP = "127.0.0.1:25504"

type fileSend struct {
	fileName string
	handle   string
	result   string
}

type sendChan chan (fileSend)

func sendFile(sent sendChan, file, sender, inDir, receiver, outDir string) {

	log.Println("sendFile( ", file, ")")

	// Setup sender exec command.
	sendCommand := exec.Command(sender, relayServiceIP, file)
	sendCommand.Dir = inDir
	stdout, pipeErr := sendCommand.StdoutPipe()
	frs.FatalCheck(pipeErr)

	sendErr := sendCommand.Start() // starts sender in background, call returns immediately
	frs.FatalCheck(sendErr)

	// Grab transfer handle from sender's stdout.
	connectionHandle := make([]byte, 10)
	_, readErr := io.ReadAtLeast(stdout, connectionHandle, 10)
	frs.FatalCheck(readErr)

	// Exec receiver with the captured transfer handle.
	log.Println("Receiving: ", string(connectionHandle))
	receiveCommand := exec.Command(receiver, relayServiceIP, string(connectionHandle), outDir)
	receiveResult, receiveErr := receiveCommand.Output()
	log.Println(string(receiveResult))
	frs.FatalCheck(receiveErr)

	// Wait for sender/receier pair to finish, signal sent channel that transfer for this file is complete.
	sendCommand.Wait()
	sent <- fileSend{file, string(connectionHandle), string(receiveResult)}
}

func sendFolder(sender, inDir, receiver, outDir string) int {

	sent := make(sendChan) // wait channel counts each file transfer

	// Launch local asynch sender and receiver process pairs for each file in the folder.
	files, err := ioutil.ReadDir(inDir)
	frs.FatalCheck(err)
	nFiles := len(files)
	log.Println("nFiles = ", nFiles)
	for iSend, file := range files {
		go sendFile(sent, file.Name(), sender, inDir, receiver, outDir)
		log.Println("iSend = ", iSend)
	}

	// Wait for all the transfers to finish.
	for wait := 0; wait < nFiles; wait++ {
		log.Println("wait = ", wait)
		result := <-sent
		log.Println("sendFile() -> ", result)
	}
	log.Println("Done!!!!")
	return nFiles
}

func main() {

	// Process command line arguments.
	if len(os.Args) != 5 {
		fmt.Println("usage: ", os.Args[0], " [sender] [send-files-path] [receiver] [received-files-path")
		os.Exit(1)
	}
	sender := os.Args[1]
	inDir := os.Args[2]
	receiver := os.Args[3]
	outDir := os.Args[4]

	time.Sleep(10 * time.Second) // wait for relay server to load

	nSent := sendFolder(sender, inDir, receiver, outDir)
	log.Println("Sent: ", nSent)
}
