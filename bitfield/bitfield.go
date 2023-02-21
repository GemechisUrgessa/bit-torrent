// Description: Bitfield is a bit array that represents the pieces that a peer has.
// It is used to determine which pieces to request from a peer.
package bitfield

type Bitfield []byte

// New creates a new bitfield of the given length.
func New(length int) Bitfield {
	return make(Bitfield, (length+7)/8)
}

// Len returns the length of the bitfield.
func (bf Bitfield) Len() int {
	return len(bf) * 8
}

// HasPiece returns true if the bitfield has the given piece.
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	return bf[byteIndex]>>(7-offset)&1 != 0
}

// SetPiece sets the given piece in the bitfield.
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// silently discard invalid bounded index
	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}
	bf[byteIndex] |= 1 << (7 - offset)
}
