package prng

import (
	"dingdongg/pkmn-platinum-rom-parser/char_encoder"
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

*/

type PRNG struct {
	Checksum uint16
	Personality uint32
	PrevResult uint
}

type UnshuffleInfo struct {
	ShuffledPos [4]int
	Displacements [4]int
}

const (
	A = iota
	B = iota
	C = iota
	D = iota
)

const BLOCK_SIZE_BYTES uint = 32
const POKEMON_STRUCTURE_SIZE uint = 236

// populated with results from the shuffler package!
var unshuffleTable [24]UnshuffleInfo = [24]UnshuffleInfo{
	{ [4]int{A, B, C, D}, [4]int{A, B, C, D} }, // ABCD ABCD
	{ [4]int{A, B, D, C}, [4]int{A, B, D, C} }, // ABDC ABDC
	{ [4]int{A, C, B, D}, [4]int{A, C, B, D} }, // ACBD ACBD
	{ [4]int{A, C, D, B}, [4]int{A, D, B, C} }, // ACDB ADBC
	{ [4]int{A, D, B, C}, [4]int{A, C, D, B} }, // ADBC ACDB
	{ [4]int{A, D, C, B}, [4]int{A, D, C, B} }, // ADCB ADCB
	{ [4]int{B, A, C, D}, [4]int{B, A, C, D} }, // BACD BACD
	{ [4]int{B, A, D, C}, [4]int{B, A, D, C} }, // BADC BADC
	{ [4]int{B, C, A, D}, [4]int{C, A, B, D} }, // BCAD CABD
	{ [4]int{B, C, D, A}, [4]int{D, A, B, C} }, // BCDA DABC
	{ [4]int{B, D, A, C}, [4]int{C, A, D, B} }, // BDAC CADB
	{ [4]int{B, D, C, A}, [4]int{D, A, C, B} }, // BDCA DACB
	{ [4]int{C, A, B, D}, [4]int{B, C, A, D} }, // CABD BCAD
	{ [4]int{C, A, D, B}, [4]int{B, D, A, C} }, // CADB BDAC
	{ [4]int{C, B, A, D}, [4]int{C, B, A, D} }, // CBAD CBAD
	{ [4]int{C, B, D, A}, [4]int{D, B, A, C} }, // CBDA DBAC
	{ [4]int{C, D, A, B}, [4]int{C, D, A, B} }, // CDAB CDAB
	{ [4]int{C, D, B, A}, [4]int{D, C, A, B} }, // CDBA DCAB
	{ [4]int{D, A, B, C}, [4]int{B, C, D, A} }, // DABC BCDA
	{ [4]int{D, A, C, B}, [4]int{B, D, C, A} }, // DACB BDCA
	{ [4]int{D, B, A, C}, [4]int{C, B, D, A} }, // DBAC CBDA
	{ [4]int{D, B, C, A}, [4]int{D, B, C, A} }, // DBCA DBCA
	{ [4]int{D, C, A, B}, [4]int{C, D, B, A} }, // DCAB CDBA
	{ [4]int{D, C, B, A}, [4]int{D, C, B, A} }, // DCBA DCBA
}

/*

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

func (usi UnshuffleInfo) UnshuffledPos(block int) uint {
	metadataOffset := uint(0x8)
	startIndex := uint(usi.Displacements[block] % 4)
	res := metadataOffset + (startIndex * BLOCK_SIZE_BYTES)
	return res
} 

func Init(checksum uint16, personality uint32) PRNG {
	return PRNG{checksum, personality, uint(checksum)}
}

func (prng *PRNG) Next() uint16 {
	result := 0x041C64E6D * prng.PrevResult + 0x06073
	prng.PrevResult = result
	result >>= 16
	// return the upper 16 bits only for external use; internally, all bits should be held for future calls
	return uint16(result & 0xFFFF) 
}

func getPokemonBlock(buf []byte, block int, personality uint32) []byte {
	shiftValue := ((personality & 0x03E000) >> 0x0D) % 24
	unshuffleInfo := unshuffleTable[shiftValue]
	startAddr := unshuffleInfo.UnshuffledPos(block)
	blockStart := buf[startAddr:startAddr + BLOCK_SIZE_BYTES]

	return blockStart
}

func (prng *PRNG) DecryptPokemons(ciphertext []byte) {
	var plaintext_buf []byte

	// 1. XOR to get plaintext words
	for i := 0x8; i < 0x87; i += 0x2 {
		// word := binary.BigEndian.Uint16(ciphertext[i:i + 2])
		word := binary.LittleEndian.Uint16(ciphertext[i:i + 2])
		plaintext := word ^ prng.Next()
		littleByte := byte(plaintext & 0x00FF)
		bigByte := byte((plaintext >> 8) & 0x00FF)
		plaintext_buf = append(plaintext_buf, littleByte, bigByte)
		// plaintext_buf = append(plaintext_buf, byte((plaintext >> 8) & 0x00FF), byte(plaintext & 0x00FF))
	}

	plaintext_buf = append(ciphertext[:8], plaintext_buf...)

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
}

// `ciphertext` must be a slice with the first byte 
// referring to the first pokemon data structure
func GetPokemon(ciphertext []byte, partyIndex uint) {
	offset := partyIndex * POKEMON_STRUCTURE_SIZE

	personality := binary.LittleEndian.Uint32(ciphertext[offset:offset + 4])
	checksum := binary.LittleEndian.Uint16(ciphertext[offset + 6:offset + 8])

	prng := Init(checksum, personality)
	prng.DecryptPokemons(ciphertext[offset:])
}
/*

shuffled			unshuffled
ABCD				ABCD
ADCB				ADCB

Blocks A, B, C, D are 32 BYTES long 
- address offsets ~ [0x8, 0x87]

DECRYPTION ALGORITHM
1. seed PRNG with checksum 
2. for every 2-byte word (w) in the address offset, apply: w ^ prng.Next()
3. unshuffle according to shuffling table found on the project pokemon website
*/
