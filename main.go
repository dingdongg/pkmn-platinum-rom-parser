package main

import (
	"io"
	"log"
	"os"

	"dingdongg/pkmn-platinum-rom-parser/prng"
)

const CHUNK_SIZE = 1576

const MONEY_OFFSET = 0x7c
const MONEY_SIZE = 4

const GENDER_OFFSET = 0x80
const GENDER_SIZE = 1

// these offsets/sizes are unique to every party pokemon
const PERSONALITY_OFFSET = 0xa0
const PERSONALITY_SIZE = 4

const CHECKSUM_OFFSET = PERSONALITY_OFFSET + 0x6
const CHECKSUM_SIZE = 2

func main() {
	buf := make([]byte, CHUNK_SIZE)

	file, err := os.Open("./Plat savefile")
	if err != nil {
		log.Fatal("bruh")
	}
	defer file.Close()

	io.ReadFull(file, buf)

	for i := uint(0); i < 6; i++ {
		prng.GetPokemon(buf[PERSONALITY_OFFSET:], i)
	}
}