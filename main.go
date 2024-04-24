package main

import (
	"fmt"
)

func main() {
	// message := os.Args[1]
	dataChannel := make(chan string)
	
	go Request("hi", dataChannel)

	for data := range dataChannel {
		fmt.Print(data)
	}
}