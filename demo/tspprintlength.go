// An output filter example. It will print for every incoming and outgoing
// message the direction and length of tcp data received.
package main

import (
	"fmt"
	"os"
	"strconv"
	"tsproxy"
)

func main() {
	if len(os.Args) < 2 {
		panic("Invalid number of arguments")
	}
	inAdd, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("First argument must be port number")
	}
	outAdd := os.Args[2]

	filterList := []tsproxy.Filter{}
	filterList = append(filterList, tsproxy.LengthPrintFilter{})

	tsp := tsproxy.TSProxy{
		InPort:     inAdd,
		OutAddress: outAdd,
		FilterList: filterList,
	}
	err = tsp.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}
