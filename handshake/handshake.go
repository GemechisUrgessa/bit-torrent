// Description: Handshake message
// Package handshake implements the handshake message used to identify a peer.

package handshake

import (
	"fmt"
	"io"
)


// A Handshake is a special message that a peer uses to identify itself 
// to another peer. 
// It contains the following fields: Pstr, InfoHash, and PeerID 
type HandShake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte 
} 
 
// New Creates a new handshake with the standard pstr and the given infohash and peerid. 
// It returns a pointer to the new handshake. 

func New(infoHash, peerID [20]byte) *HandShake {
 	return &HandShake{
 		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}


// Serialize serializes the handshake into a byte slice. 
// It returns the serialized handshake. 
// It returns an error if one occurred. 

func (h *HandShake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}



// Read reads a handshake from the given reader.
// It returns a pointer to the handshake and an error if one occurred.
// It returns an error if the pstrlen is 0.

func Read(r io.Reader) (*HandShake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	h := HandShake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
