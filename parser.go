package parser

import (
	"github.com/dingdongg/pkmn-platinum-rom-parser/rom_reader"
	"io"
	"log"
	"os"
	"path/filepath"
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

func Parse() {
	savefile := make([]byte, CHUNK_SIZE)

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("failed fetching pwd: ", err)
	}

	fpath := filepath.Clean(dir + "/../pkmn-platinum-rom-parser/savefiles/Pt_savefile-v2")
	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal("bruh, ", err)
	}
	defer file.Close()

	io.ReadFull(file, savefile)

	for i := uint(0); i < 6; i++ {
		rom_reader.GetPokemon(savefile[PERSONALITY_OFFSET:], i)
	}
}