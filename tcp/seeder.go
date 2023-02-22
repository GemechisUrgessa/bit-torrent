package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

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

func addMoreTorrents(client *torrent.Client) ([]*torrent.Torrent, error) {
	var torrents []*torrent.Torrent
	for {
		t, err := addTorrent(client)
		if err != nil {
			return nil, err
		}
		torrents = append(torrents, t)
		fmt.Print("Do you want to add another torrent file? (y/n) ")
		reader := bufio.NewReader(os.Stdin)
		add, _ := reader.ReadString('\n')
		add = strings.TrimSpace(add)
		if add == "n" {
			break
		}
	}
	return torrents, nil
}

func seedTorrents(torrents []*torrent.Torrent) error {
	fmt.Println("Starting seeding...")
	status := make(chan string)
	var wg sync.WaitGroup
	for _, t := range torrents {
		wg.Add(1)
		go func(t *torrent.Torrent) {
			t.DownloadAll()
			for t.BytesCompleted() < t.Info().TotalLength() {
				select {
				case <-t.GotInfo():
					percent := 100 * float64(t.BytesCompleted()) / float64(t.Info().TotalLength())
					status <- fmt.Sprintf("%s: %.2f%%", t.Name(), percent)
				}
			}
			status <- fmt.Sprintf("%s: 100%%", t.Name())
			wg.Done()
		}(t)
	}
	wg.Wait()
	close(status)
	for s := range status {
		fmt.Println(s)
	}
	return nil
}

func main() {
	client, err := newClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	torrents, err := addMoreTorrents(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = seedTorrents(torrents)
	if err != nil {
		fmt.Println(err)
		return
	}
}
