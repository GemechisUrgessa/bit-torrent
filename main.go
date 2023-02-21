// Description: This is the entry point for the program
package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"bit-torrent/seeder"
	"bit-torrent/torrent"
)

// main is the entry point for the program
// It takes in two arguments: the path to the .torrent file and the path to the file to be downloaded to (the file must not exist)
// It connects to peers and downloads the file
// It then starts seeding the file to the peers that are connected to it and waits for the user to press enter to exit
func main() {
	// Check if the correct number of arguments are passed in
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <path to .torrent file> <path to file to download to>")
		return
	}

	// Get the paths to the .torrent file and the file to download to from the arguments
	inPath := os.Args[1]
	outPath := os.Args[2]

	// Open the dot torrent file
	tf, err := torrent.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	// Get the torrent struct
	tor, err := tf.GetTorrent()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to peers and download file and start seeding
	fmt.Println("Connecting to peers...")
	keepAliveChan := make(chan bool)
	clients, err := torrent.ConnectToPeers(tor, keepAliveChan)
	fmt.Printf("Number of clients is %d\n", len(clients))
	if err != nil {
		log.Fatal(err)
	}
	// Start a goroutine to send keep alive messages to the peers
	go func() {
		for {
			select {
			case <-keepAliveChan:
				for _, c := range clients {
					c.SendKeepAlive()
				}
			}
		}
	}()

	// Download file and start seeding
	fmt.Println("Downloading file....")
	err = tf.DownloadToFile(outPath, tor, clients)
	if err != nil {
		log.Fatal(err)
	}
	//
	var wg sync.WaitGroup
	// Add one to the wait group
	wg.Add(1)
	// Start seeding the file

	go func() {
		defer wg.Done()
		fmt.Println("Starting to seed file...")
		seeder.SeedFile(clients, tor, outPath)
	}()
	// Wait for user to press enter to exit
	fmt.Println("Leeching and seeding complete. Press enter to exit")
	wg.Wait()
	fmt.Println("Exiting...")

}
