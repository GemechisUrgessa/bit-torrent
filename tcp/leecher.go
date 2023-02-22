package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
)

func main() {
	// Get magnet link from terminal
	fmt.Print("Enter magnet link: ")
	var magnetLink string
	fmt.Scanln(&magnetLink)

	// Create a new client
	client, err := torrent.NewClient(nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Add the magnet link to the client
	t, err := client.AddMagnet(magnetLink)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait for the torrent to be ready
	<-t.GotInfo()

	// Select the first file to download
	file := t.Files()[0]

	// create the nested folder if not exists
	path := "downloads/" + filepath.Dir(file.Path())
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	// Create a new file to save the downloaded data
	outFile, err := os.Create("downloads/" + file.Path())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Download the file
	_, err = io.Copy(outFile, file.NewReader())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Close the file
	outFile.Close()

	// Close the client
	client.Close()
}
