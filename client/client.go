// Description: Client is a wrapper around a net.Conn that implements the BitTorrent protocol.
// It is used to communicate with peers.

package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"bit-torrent/bitfield"
	"bit-torrent/handshake"
	"bit-torrent/message"
	"bit-torrent/peers"
)

// this is a Client struct that contains the following fields:  Conn, Choked, Bitfield, Peer, infoHash, and peerID
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	Peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

// completeHandShake completes the handshake with the peer
// It returns the handshake response and an error if one occurred.
func completeHandShake(conn net.Conn, infohash, peerID [20]byte) (*handshake.HandShake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}
	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x",
			res.InfoHash, infohash)
	}

	return res, nil
}

// recvBitfield receives a bitfield message from the peer
// It returns the bitfield and an error if one occurred.
func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}
	return msg.Payload, nil
}

// New creates a new Client
// It returns the client and an error if one occurred.
func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	_, err = completeHandShake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}
	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		Peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

// Close closes the connection
func (c *Client) Close() error {
	return c.Conn.Close()
}

// Read reads and consumes a message from the connection
// It returns the message and an error if one occurred.
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// SendRequest sends a Request message to the peer
// It returns an error if one occurred.
func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// SendInterested sends an Interested message to the peer
// It returns an error if one occurred.
func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a NotInterested message to the peer
// It returns an error if one occurred.
func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an Unchoke message to the peer
// It returns an error if one occurred.
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendPiece senda a piece message to the peer
// It returns an error if one occurred.
func (c *Client) SendPiece(index, begin int, data []byte) error {
	msg := message.FormatPiece(index, begin, data)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a Have message to the peer
// It returns an error if one occurred.
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendKeepAlive sends a KeepAlive message to the peer
// It returns an error if one occurred.
func (c *Client) SendKeepAlive() error {
	msg := message.Message{}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
