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

