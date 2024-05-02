package prng

type PRNG struct {
	Checksum uint16
	Personality uint32
	PrevResult uint
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
