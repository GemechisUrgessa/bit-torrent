// Description: This file contains the main logic for the peer2peer package.
// It contains the Torrent struct, which holds the data required to download a torrent.
// It also contains the main download function, which starts a worker for each peer and distributes work to them.
// Finally, it contains the logic for downloading a piece from a peer.
package peer2peer

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"bit-torrent/client"
	"bit-torrent/message"
	"bit-torrent/peers"
)

const MaxBlockSize = 16384

const MaxBacklog = 5

// Torrent holds data required to download a torrent from a list of peers
// It contains the following fields: Peers, PeerID, InfoHash, PieceHashes, PieceLength, Length, and Name  (all of which are of type []byte)
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// this struct contains the following fields: index, hash, and length
type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

// this struct contains the following fields: index, buf
type pieceResult struct {
	index int
	buf   []byte
}

// this struct contains the following fields: index, client, buf, downloaded, requested, and backlog
type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}
// startDownloadWorker starts a worker that downloads pieces from a peer and puts them on the results queue when done downloading them (or when an error occurs)


func (t *Torrent) startDownloadWorker(c *client.Client, workQueue chan *pieceWork,
	results chan *pieceResult) {
	// c, err := client.New(peer, t.PeerID, t.InfoHash)
	// if err != nil {
	// 	log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
	// 	return
	// }
	// defer c.Conn.Close()
	// log.Printf("Completed handshake with %s\n", peer.IP)
	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // Put piece
			continue
		}
		// Download the piece

		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}
		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

// readMessage reads a message from the peer and updates the pieceProgress struct accordingly (if the message is a piece message)
func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() // this call blocks
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

// attemptDownloadPiece attempts to download a piece from a peer and returns the piece data (or an error if it fails)
func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

// checkIntegrity checks if the downloaded piece matches the hash in the torrent file and returns an error if it doesn't
func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}

// calculateBoundsForPiece calculates the begin and end byte offsets for a piece with the given index in the torrent file and returns them as a tuple
func (t *Torrent) calculateBoundsForPiece(index int) (begin, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

// calculatePieceSize calculates the size of a piece with the given index in the torrent file and returns it as an integer
func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

// download downloads the torrent file and returns the file data (or an error if it fails)
func (t *Torrent) Download(clients []*client.Client) ([]byte, error) {
	log.Println("Starting download for", t.Name)
	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start worker
	for _, client := range clients {
		go t.startDownloadWorker(client, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // substrat one main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workQueue)

	return buf, nil
}
