package tsproxy

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	// TCPBufferReadSize buffer size when reading tcp message
	TCPBufferReadSize int = 4096
)

type Filter interface {
	// FilterInput
	FilterInput(message []byte)
	FilterOutput(message []byte)
}

// TSProxy implements a simple proxy that listens for connections on inPort
// and creates outgoing connections to outAddress while passing all through
// message through filterList
type TSProxy struct {
	inPort     int
	outAddress string
	filterList []Filter
}

// NewTSProxy returns a new TSProxy initialized with the given arguments
func NewTSProxy(inPort int, outAddress string, filterList []Filter) *TSProxy {
	tsp := new(TSProxy)
	tsp.inPort = inPort
	tsp.outAddress = outAddress
	tsp.filterList = filterList
	return tsp
}

// Run runs in an infinite loop accepting connections
func (tsp TSProxy) Run() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", tsp.inPort))
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %e\n", err)
			return err
		}
		go tsp.handleConnection(conn)
	}
	return nil
}

// handleConnection for the incoming connection create an outgoing connection
// and handle input and output and filter calling
func (tsp TSProxy) handleConnection(conn net.Conn) {
	var wg sync.WaitGroup
	defer conn.Close()

	wg.Add(6)

	inChan := make(chan []byte)
	resendChan := make(chan []byte)
	backChan := make(chan []byte)
	outChan := make(chan []byte)

	fBufRdr := bufio.NewReader(conn)
	fBufWrt := bufio.NewWriter(conn)
	frontBuf := bufio.NewReadWriter(fBufRdr, fBufWrt)

	// Connect to endpoint
	repConn, err := net.Dial("tcp", tsp.outAddress)
	if err != nil {
		return
	}
	defer repConn.Close()

	// Define ReadWriter for endpoint
	bBufRdr := bufio.NewReader(repConn)
	bBufWrt := bufio.NewWriter(repConn)
	backBuf := bufio.NewReadWriter(bBufRdr, bBufWrt)

	// PROXY
	go func() {
		for {
			in, ok := <-inChan
			if !ok {
				break
			}
			for _, f := range tsp.filterList {
				f.FilterInput(in)
			}
			resendChan <- in
		}
		wg.Done()
	}()
	go func() {
		for {
			back, ok := <-backChan
			if !ok {
				break
			}
			for _, f := range tsp.filterList {
				f.FilterOutput(back)
			}
			outChan <- back
		}
		wg.Done()
	}()

	go func() {
		res := make([]byte, TCPBufferReadSize)
		for {
			read, err := frontBuf.Read(res)
			if err != nil {
				break
			}
			inChan <- res[:read]
		}
		close(inChan)
		close(outChan)
		repConn.Close()
		wg.Done()
	}()
	go func() {
		res := make([]byte, TCPBufferReadSize)
		for {
			read, err := backBuf.Read(res)
			if err != nil {
				break
			}
			backChan <- res[:read]
		}
		close(backChan)
		close(resendChan)
		conn.Close()
		wg.Done()
	}()
	go func() {
		for {
			rep, ok := <-resendChan
			if !ok {
				break
			}
			_, err := backBuf.Write(rep)
			if err != nil {
				continue
			}
			err = backBuf.Flush()
			if err != nil {
				continue
			}
		}
		wg.Done()
	}()
	go func() {
		for {
			rep, ok := <-outChan
			if !ok {
				break
			}
			_, err := frontBuf.Write(rep)
			if err != nil {
				continue
			}
			err = frontBuf.Flush()
			if err != nil {
				continue
			}
		}
		wg.Done()
	}()
	wg.Wait()
	return
}
