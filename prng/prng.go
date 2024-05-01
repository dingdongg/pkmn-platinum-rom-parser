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
	for i := 0x8; i < 0x87; i += 0x2 {
		// ...
	}
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
