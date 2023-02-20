// Description: Message package for the BitTorrent client implementation.
// This is a package that contains the following:  messageID, Message, and ParseRequest functions and types
// it is used to parse messages from the peer
package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

// this is a messageID type that is an unsigned 8-bit integer (uint8)
type messageID uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageID = 0

	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageID = 1

	// MsgInterested expresses interest in receiving data
	MsgInterested messageID = 2

	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageID = 3

	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageID = 4

	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5

	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6

	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7

	// MsgCancel cancels a request
	MsgCancel messageID = 8
)


// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// ParseRequst parses a Request message and returns the index, begin, and length
// of the requested piece
// this is a function that takes a pointer to a Message and returns 3 integers and an error
func ParseRequest(msg *Message) (index, begin, length int, err error) {
	if msg.ID != MsgRequest {
		return 0, 0, 0, fmt.Errorf("Invalid message ID for ParseRequest: %d", msg.ID)
	}

	if len(msg.Payload) != 12 {
		return 0, 0, 0, fmt.Errorf("Invalid payload length for ParseRequest: %d", len(msg.Payload))
	}

	index = int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	begin = int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	length = int(binary.BigEndian.Uint32(msg.Payload[8:12]))

	return index, begin, length, nil
}

// FormatPiece create a Piece message
// FormatPiece create a Piece message
// this is a function that takes 3 integers and a byte slice and returns a pointer to a Message
func FormatPiece(index, begin int, data []byte) *Message {
	payload := make([]byte, 8+len(data))

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], data)
	return &Message{ID: MsgPiece, Payload: payload}
}

// FormatRequest creates a REQUEST message
// this is a function that takes 3 integers and returns a pointer to a Message
// returns a pointer to a Message

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}


// FormatHave creates a HAVE message
// this is a function that takes an integer and returns a pointer to a Message
// returns a pointer to a Message

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}


// ParsePiece parses a PIECE message and copies its payload into a buffer
// this is a function that takes an integer, a byte slice, and a pointer to a Message and returns an integer and an error
// returns an integer and an error

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {

	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}

	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}

	copy(buf[begin:], data)
	return len(data), nil
}

// ParseHave parses a HAVE message
func ParseHave(msg *Message) (int, error) {

	if msg.ID != MsgHave {
		return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
	}

	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}


	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	
	return buf
}


// Read parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		// fmt.Println("Error while reading the legthBuf")
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		// fmt.Println("Error while reading the message buf")
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}

func (m *Message) name() string {
	if m == nil {
		return "KeepAlive"
	}


	switch m.ID {
	case MsgChoke:
		return "Choke"

	case MsgUnchoke:
		return "Unchoke"

	case MsgInterested:
		return "Interested"

	case MsgNotInterested:
		return "NotInterested"

	case MsgHave:
		return "Have"

	case MsgBitfield:
		return "Bitfield"

	case MsgRequest:
		return "Request"

	case MsgPiece:
		return "Piece"

	case MsgCancel:
		return "Cancel"

	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}


func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}


