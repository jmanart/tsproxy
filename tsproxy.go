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

// TSProxy implements a simple proxy that listens for connections on InPort
// and creates outgoing connections to OutAddress while passing all through
// message through FilterList
type TSProxy struct {
	InPort     int
	OutAddress string
	FilterList []Filter
}

// Run runs in an infinite loop accepting connections
func (tsp TSProxy) Run() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", tsp.InPort))
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

	in_chan := make(chan []byte)
	resend_chan := make(chan []byte)
	back_chan := make(chan []byte)
	out_chan := make(chan []byte)

	fBufRdr := bufio.NewReader(conn)
	fBufWrt := bufio.NewWriter(conn)
	frontBuf := bufio.NewReadWriter(fBufRdr, fBufWrt)

	// Connect to endpoint
	rep_conn, err := net.Dial("tcp", tsp.OutAddress)
	if err != nil {
		return
	}
	defer rep_conn.Close()

	// Define ReadWriter for endpoint
	bBufRdr := bufio.NewReader(rep_conn)
	bBufWrt := bufio.NewWriter(rep_conn)
	backBuf := bufio.NewReadWriter(bBufRdr, bBufWrt)

	// PROXY
	go func() {
		for {
			in, ok := <-in_chan
			if !ok {
				break
			}
			for _, f := range tsp.FilterList {
				f.FilterInput(in)
			}
			resend_chan <- in
		}
		wg.Done()
	}()
	go func() {
		for {
			back, ok := <-back_chan
			if !ok {
				break
			}
			for _, f := range tsp.FilterList {
				f.FilterOutput(back)
			}
			out_chan <- back
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
			in_chan <- res[:read]
		}
		close(in_chan)
		close(out_chan)
		rep_conn.Close()
		wg.Done()
	}()
	go func() {
		res := make([]byte, TCPBufferReadSize)
		for {
			read, err := backBuf.Read(res)
			if err != nil {
				break
			}
			back_chan <- res[:read]
		}
		close(back_chan)
		close(resend_chan)
		conn.Close()
		wg.Done()
	}()
	go func() {
		for {
			rep, ok := <-resend_chan
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
			rep, ok := <-out_chan
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
