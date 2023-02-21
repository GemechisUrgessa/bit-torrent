package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/anacrolix/torrent"
)

func newClient() (*torrent.Client, error) {
	return torrent.NewClient(nil)
}

func addTorrent(client *torrent.Client) (*torrent.Torrent, error) {
	fmt.Print("Enter a torrent file path: ")
	reader := bufio.NewReader(os.Stdin)
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	return client.AddTorrentFromFile(path)
}
