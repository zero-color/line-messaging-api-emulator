package shortid

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var New = newID

// readableAlphabet is the alphabet used for the generated IDs.
// It excludes characters that are hard to distinguish such as 0/O and I/l/1.
const readableAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWYZabcdeghkmnpqrsuvwyz"

func newID() string {
	// https://zelark.github.io/nano-id-cc/
	return gonanoid.MustGenerate(readableAlphabet, 11)
}
