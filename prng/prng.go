package prng

import "encoding/binary"

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
*/

func Init(checksum uint16, personality uint32) PRNG {
	return PRNG{checksum, personality, 0}
}

func (prng *PRNG) Next() uint16 {
	result := 0x41C64E6D * prng.PrevResult + 0x6073
	result >>= 16
	prng.PrevResult = result
	return uint16(result)
}

func (prng *PRNG) Decrypt(word []byte) uint16 {
	uint16Word := binary.LittleEndian.Uint16(word)
	decrypted := uint16Word ^ prng.Next()
	return decrypted
}

func (prng *PRNG) DecryptPokemons(ciphertext []byte) {
	var plaintext_buf []byte

	// 1. XOR to get plaintext words
	for i := 0x8; i < 0x87; i += 0x2 {
		// ...
		word := binary.LittleEndian.Uint16(ciphertext[i:i + 2])
		plaintext := word ^ prng.Next()
		plaintext_buf = append(plaintext_buf, byte(plaintext & 0x00FF), byte(plaintext & 0xFF00))
	}

	// 2. de-shuffle

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
