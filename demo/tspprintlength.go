// An output filter example. It will print for every incoming and outgoing
// message the direction and length of tcp data received.
package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/jmanart/tsproxy"
)

func usage() {
	fmt.Println("Usage: tspprintlength <Listening Port> <Outgoing Address>")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		return
	}
	inAdd, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		panic("First Argument must be a number")
	}
	outAdd := flag.Arg(1)

	filterList := []tsproxy.Filter{}
	filterList = append(filterList, tsproxy.LengthPrintFilter{})

	tsp := tsproxy.NewTSProxy(inAdd, outAdd, filterList)
	err = tsp.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}
