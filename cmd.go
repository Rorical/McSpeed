package main

import "github.com/Rorical/McSpeed/server"

func main() {
	err := server.Loop()
	if err != nil {
		panic(err)
	}
}
