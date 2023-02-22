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
	inPath := os.Args[1]
	outPath := os.Args[2]

	// Open the dot torrent file
	tf, err := torrent.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	tor, err := tf.GetTorrent()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to peers and download file
	keepAliveChan := make(chan bool)
	clients, err := torrent.ConnectToPeers(tor, keepAliveChan)
	fmt.Printf("Number of clients is %d\n", len(clients))
	if err != nil {
		log.Fatal(err)
	}
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
	err = tf.DownloadToFile(outPath, tor, clients)
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		seeder.SeedFile(clients, tor, outPath)
	}()

	wg.Wait()
	fmt.Println("Main function completed")
}
