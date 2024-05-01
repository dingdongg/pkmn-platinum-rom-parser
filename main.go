package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"dingdongg/pkmn-platinum-rom-parser/char_encoder"
)

const CHUNK_SIZE = 1576

const MONEY_OFFSET = 0x7c
const MONEY_SIZE = 4

const GENDER_OFFSET = 0x80
const GENDER_SIZE = 1

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

	gender := uint8(buf[GENDER_OFFSET])

	if gender == 0 {
		fmt.Println("MALE")
	} else if gender == 1 {
		fmt.Println("FEMALE")
	}

	r := char_encoder.Char(0x003b)
	fmt.Printf("'%s'\n", r)
}