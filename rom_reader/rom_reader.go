package rom_reader

import (
	"github.com/dingdongg/pkmn-platinum-rom-parser/char_encoder"
	"github.com/dingdongg/pkmn-platinum-rom-parser/prng"
	"encoding/binary"
	"fmt"
)

/*

X[n+1] = (0x41C64E6D * X[n] + 0x6073)
To decrypt the data, given a function rand() which returns the
upper 16 bits
of consecutive results of the above given function:

1. Seed the PRNG with the checksum (let X[n] be the checksum).
2. Sequentially, for each 2-byte word Y from 0x08 to 0x87, apply the transformation: unencryptedByte = Y xor rand()
3. Unshuffle the blocks using the block shuffling algorithm above.

---

ex. ACDB -> ADBC
A -> move 0 to the right
B -> move 3 to the right (wrap around)
C -> move 2 to the right
D -> move 3 to the right (wrap around)

ACDB, [0, 3, 2, 3]
[0, 3, 1, 2], [0, 3, 2, 3]

to get block A (represented as ShuffledPos[0]), 
(ShuffledPos[0] + Displacements[0]) % 4 = 0 (un-shuffled position)

to get block B (ShuffledPos[1]),
(ShuffledPos[1] + Displacements[1]) % 4 = (3 + 3) % 4 = 2 (unshuffled position)

to get block C (ShuffledPos[2]),
(ShuffledPos[2] + Displacements[2]) % 4 = (1 + 2) % 4 = 3 (unshuffled position)

to get block D (ShuffledPos[3]),
(2 + 3) % 4 = 1

now we have the unshuffled positions [0, 2, 3, 1] -> ADBC
each block is 32 bytes long, so we can access any block in constant time with offset calculations

eg. get block C -> unshuffled[2] = 3 --> 0x8 + (3 * 0x20) = 0x63 (starting position for block C)

questions i still have:
- are the decrypted values just represented as little endian? or big? (hopefully LE, since that's what it seems like the ROM stuck to thus far)
	- RESOLVED: little endian seems to work
*/

type blockOrder struct {
	ShuffledPos [4]uint
	OriginalPos [4]uint
}

type EffortValues struct {
	Hp			uint
	Attack		uint
	Defense 	uint
	SpAttack	uint
	SpDefense	uint
	Speed		uint
}

type Pokemon struct {
	Name 		string
	Level		uint
	EffortValues
}

const (
	A uint = iota
	B uint = iota
	C uint = iota
	D uint = iota
)

const BLOCK_SIZE_BYTES uint = 32
const PARTY_POKEMON_SIZE uint = 236

// populated with results from the shuffler package!
var unshuffleTable [24]blockOrder = [24]blockOrder{
	{ [4]uint{A, B, C, D}, [4]uint{A, B, C, D} }, // ABCD ABCD
	{ [4]uint{A, B, D, C}, [4]uint{A, B, D, C} }, // ABDC ABDC
	{ [4]uint{A, C, B, D}, [4]uint{A, C, B, D} }, // ACBD ACBD
	{ [4]uint{A, C, D, B}, [4]uint{A, D, B, C} }, // ACDB ADBC
	{ [4]uint{A, D, B, C}, [4]uint{A, C, D, B} }, // ADBC ACDB
	{ [4]uint{A, D, C, B}, [4]uint{A, D, C, B} }, // ADCB ADCB
	{ [4]uint{B, A, C, D}, [4]uint{B, A, C, D} }, // BACD BACD
	{ [4]uint{B, A, D, C}, [4]uint{B, A, D, C} }, // BADC BADC
	{ [4]uint{B, C, A, D}, [4]uint{C, A, B, D} }, // BCAD CABD
	{ [4]uint{B, C, D, A}, [4]uint{D, A, B, C} }, // BCDA DABC
	{ [4]uint{B, D, A, C}, [4]uint{C, A, D, B} }, // BDAC CADB
	{ [4]uint{B, D, C, A}, [4]uint{D, A, C, B} }, // BDCA DACB
	{ [4]uint{C, A, B, D}, [4]uint{B, C, A, D} }, // CABD BCAD
	{ [4]uint{C, A, D, B}, [4]uint{B, D, A, C} }, // CADB BDAC
	{ [4]uint{C, B, A, D}, [4]uint{C, B, A, D} }, // CBAD CBAD
	{ [4]uint{C, B, D, A}, [4]uint{D, B, A, C} }, // CBDA DBAC
	{ [4]uint{C, D, A, B}, [4]uint{C, D, A, B} }, // CDAB CDAB
	{ [4]uint{C, D, B, A}, [4]uint{D, C, A, B} }, // CDBA DCAB
	{ [4]uint{D, A, B, C}, [4]uint{B, C, D, A} }, // DABC BCDA
	{ [4]uint{D, A, C, B}, [4]uint{B, D, C, A} }, // DACB BDCA
	{ [4]uint{D, B, A, C}, [4]uint{C, B, D, A} }, // DBAC CBDA
	{ [4]uint{D, B, C, A}, [4]uint{D, B, C, A} }, // DBCA DBCA
	{ [4]uint{D, C, A, B}, [4]uint{C, D, B, A} }, // DCAB CDBA
	{ [4]uint{D, C, B, A}, [4]uint{D, C, B, A} }, // DCBA DCBA
}

// `ciphertext` must be a slice with the first byte 
// referring to the first pokemon data structure
func GetPokemon(ciphertext []byte, partyIndex uint) Pokemon {
	offset := partyIndex * PARTY_POKEMON_SIZE

	personality := binary.LittleEndian.Uint32(ciphertext[offset:offset + 4])
	checksum := binary.LittleEndian.Uint16(ciphertext[offset + 6:offset + 8])

	rand := prng.Init(checksum, personality)
	return decryptPokemon(rand, ciphertext[offset:])
}

func (bo blockOrder) getUnshuffledPos(block uint) uint {
	metadataOffset := uint(0x8)
	startIndex := bo.OriginalPos[block]
	res := metadataOffset + (startIndex * BLOCK_SIZE_BYTES)
	return res
} 

func getPokemonBlock(buf []byte, block uint, personality uint32) []byte {
	shiftValue := ((personality & 0x03E000) >> 0x0D) % 24
	unshuffleInfo := unshuffleTable[shiftValue]
	startAddr := unshuffleInfo.getUnshuffledPos(block)
	blockStart := buf[startAddr:startAddr + BLOCK_SIZE_BYTES]

	return blockStart
}

// first block of ciphertext points to offset 0x88 in a party pokemon block
// TODO: needs some validation/testing
func getPokemonLevel(ciphertext []byte, personality uint32) uint {
	bsprng := prng.InitBattleStatPRNG(personality)

	var decrypted uint16

	for i := 0; i < 6; i += 2 {
		decrypted = bsprng.Next() ^ binary.LittleEndian.Uint16(ciphertext[i:i + 2])
	}

	fmt.Printf("% x\n", decrypted)
	return uint(decrypted & 0xFF)
}

func decryptPokemon(prng prng.PRNG, ciphertext []byte) Pokemon {
	plaintext_buf := ciphertext[:8]

	// 1. XOR to get plaintext words
	for i := 0x8; i < 0x87; i += 0x2 {
		word := binary.LittleEndian.Uint16(ciphertext[i:i + 2])
		plaintext := word ^ prng.Next()
		littleByte := byte(plaintext & 0x00FF)
		bigByte := byte((plaintext >> 8) & 0x00FF)
		plaintext_buf = append(plaintext_buf, littleByte, bigByte)
	}

	// 2. un-shuffle
	blockA := getPokemonBlock(plaintext_buf, A, prng.Personality)
	blockC := getPokemonBlock(plaintext_buf, C, prng.Personality)

	pokemonNameLength := 22
	name := ""

	for i := 0; i < pokemonNameLength; i += 2 {
		charIndex := binary.LittleEndian.Uint16(blockC[i:i + 2])
		str, err := char_encoder.Char(charIndex)
		if err != nil { 
			break
		}
		name += str
	}

	fmt.Printf("Pokemon: '%s'\n", name)

	level := getPokemonLevel(ciphertext[0x88:], prng.Personality)

	hpEVOffset := 0x10
	attackEVOffset := 0x11
	defenseEVOffset := 0x12
	speedEVOffset := 0x13
	specialAtkEVOffset := 0x14
	specialDefEVOffset := 0x15

	fmt.Printf(
		"Stats:\n\t- HP:  %d\n\t- ATK: %d\n\t- DEF: %d\n\t- SpA: %d\n\t- SpD: %d\n\t- SPE: %d\n",
		blockA[hpEVOffset], blockA[attackEVOffset],
		blockA[defenseEVOffset], blockA[specialAtkEVOffset],
		blockA[specialDefEVOffset], blockA[speedEVOffset],
	)

	evSum := 0
	for i := 0; i < 6; i++ {
		evSum += int(blockA[hpEVOffset + i])
	}

	fmt.Printf("Total EV Spenditure: %d / 510\n", evSum)

	return Pokemon{
		name,
		level,
		EffortValues{
			uint(blockA[hpEVOffset]),
			uint(blockA[attackEVOffset]),
			uint(blockA[defenseEVOffset]),
			uint(blockA[specialAtkEVOffset]),
			uint(blockA[specialDefEVOffset]),
			uint(blockA[speedEVOffset]),
		},
	}
}
