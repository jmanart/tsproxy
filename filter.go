package tsproxy

import (
	"bufio"
	"fmt"
)

// LengthPrintFilter implements screen print of
// length of ingoing and outgoing messages
type LengthPrintFilter struct{}

// FilterOutput prints message length
func (p LengthPrintFilter) FilterOutput(message []byte) {
	fmt.Printf(">>> %d\n", len(message))
}

// FilterInput prints message length
func (p LengthPrintFilter) FilterInput(message []byte) {
	fmt.Printf("<<< %d\n", len(message))
}

// BufWriteFilter implements bufio.writers dump of ingoing and
// outgoing messages
type BufWriteFilter struct {
	BufferIn  bufio.Writer
	BufferOut bufio.Writer
}

// FilterOutput writes message to BufferOut
func (bw BufWriteFilter) FilterOutput(message []byte) {
	bw.BufferOut.Write(message)
}

// FilterInput writes message to BufferIn
func (bw BufWriteFilter) FilterInput(message []byte) {
	bw.BufferIn.Write(message)
}
