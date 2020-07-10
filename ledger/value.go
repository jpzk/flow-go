package ledger

import "encoding/hex"

// Value holds the value part of a ledger key value pair
type Value []byte

// Size returns the value size
func (v Value) Size() int {
	return len(v)
}

func (v Value) String() string {
	return hex.EncodeToString(v)
}