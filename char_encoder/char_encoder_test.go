package char_encoder

import "testing"

func TestCharOutOfBoundsIndex(t *testing.T) {
	_, err := Char(1000)

	if err == nil {
		t.Fatalf("Out of bounds index not handled properly\n")
	}
}

func TestCharEOSCharacter(t *testing.T) {
	_, err := Char(0xFFFF)

	if err == nil {
		t.Fatalf("Null-terminating character not handled properly\n")
	}
}

func TestCharNullCharacter(t *testing.T) {
	_, err := Char(0x0)

	if err == nil {
		t.Fatalf("Null character not handled properly\n")
	}
}