package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

const CHUNK_SIZE = 1576

const MONEY_OFFSET = 0x7c
const MONEY_SIZE = 4

func main() {
	fmt.Println("HELLO WORLD")

	buf := make([]byte, CHUNK_SIZE)

	file, err := os.Open("./Plat savefile")
	if err != nil {
		log.Fatal("bruh")
	}
	defer file.Close()

	io.ReadFull(file, buf)
	
	// get da money info (should be 133440 in decimal)
	money := binary.LittleEndian.Uint32(buf[MONEY_OFFSET:MONEY_OFFSET + MONEY_SIZE])

	fmt.Println(money)
}