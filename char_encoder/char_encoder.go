package char_encoder

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

const END_OF_STRING uint16 = 0xFFFF
const NULL_CHAR uint16 = 0x0

func Char(index uint16) (string, error) {
	file, err := os.ReadFile("char_encoder/table.json")
	if err != nil {
		log.Fatal("Error parsing char table file: ", err)
	}
	
	var chars []string
	err = json.Unmarshal(file, &chars)
	if err != nil {
		log.Fatal("oops ?? ", err)
	}

	if index == END_OF_STRING || index == NULL_CHAR {
		// end of string
		return "", errors.New("invalid index")
	}
	
	return chars[index], nil
}