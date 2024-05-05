package parser

import (
	"github.com/dingdongg/pkmn-platinum-rom-parser/rom_reader"
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

func Parse(savefile []byte) []rom_reader.Pokemon {
	// TODO: savefile size/format validation 
	// for format validation, could probably use checksums

	// TODO: only edit the most recent savefiel
	var res []rom_reader.Pokemon

	for i := uint(0); i < 6; i++ {
		res = append(res, rom_reader.GetPokemon(savefile[PERSONALITY_OFFSET:], i))
	}

	return res
}