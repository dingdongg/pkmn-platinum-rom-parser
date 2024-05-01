package prng

import (
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

*/

func (usi UnshuffleInfo) UnshuffledPos(block int) uint {
	metadataOffset := uint(0x8)
	startAddr := uint((usi.ShuffledPos[block] + usi.Displacements[block]) % 4)
	return metadataOffset + (startAddr * BLOCK_SIZE_BYTES)
} 

func Init(checksum uint16, personality uint32) PRNG {
	return PRNG{checksum, personality, 0}
}

func (prng *PRNG) Next() uint16 {
	result := 0x41C64E6D * prng.PrevResult + 0x6073
	result >>= 16
	prng.PrevResult = result
	return uint16(result)
}

func getPokemonBlock(buf []byte, block int, personality int) []byte {
	shiftValue := ((personality & 0x3E000) >> 0xD) % 24
	unshuffleInfo := unshuffleTable[shiftValue]
	startAddr := unshuffleInfo.UnshuffledPos(block)
	blockStart := buf[startAddr:startAddr + BLOCK_SIZE_BYTES]

	return blockStart
}

func (prng *PRNG) DecryptPokemons(ciphertext []byte) {
	var plaintext_buf []byte

	// 1. XOR to get plaintext words
	for i := 0x8; i < 0x87; i += 0x2 {
		// ...
		word := binary.BigEndian.Uint16(ciphertext[i:i + 2])
		plaintext := word ^ prng.Next()
		plaintext_buf = append(plaintext_buf, byte(plaintext & 0x00FF), byte(plaintext & 0xFF00))
	}

	plaintext_buf = append(ciphertext[:8], plaintext_buf...)

	fmt.Printf("% x\n", plaintext_buf)

	// 2. de-shuffle
	personalityIndex := ((prng.Personality & 0x3E000) >> 0xD) % 24
	unshuffleInfo := unshuffleTable[personalityIndex]

	startA := unshuffleInfo.UnshuffledPos(A)
	blockA := plaintext_buf[startA:startA + BLOCK_SIZE_BYTES]

	// fmt.Printf("start addr for block A: %x\n", startA)
	// fmt.Printf("% x\n", blockA)

	natPokedexId := binary.LittleEndian.Uint16(blockA[:0x2])

	fmt.Printf("national dex ID: %d\n", natPokedexId)
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
