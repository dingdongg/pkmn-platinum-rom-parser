package char_encoder

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	// "path/filepath"
)

const END_OF_STRING uint16 = 0xFFFF
const NULL_CHAR uint16 = 0x0

func Char(index uint16) (string, error) {
	// dir, err := os.Getwd()
	// if err != nil {
	// 	log.Fatal("failed fetching pwd: ", err)
	// }
	// fmt.Println(dir)
	// fpath := filepath.Clean(dir + "/../pkmn-platinum-rom-parser/char_encoder/table.json")
	fpath := "char_encoder/table.json"
	file, err := os.ReadFile(fpath)
	
	if err != nil {
		log.Fatal("Error parsing char table file: ", err)
	}
	
	var chars []string
	err = json.Unmarshal(file, &chars)
	if err != nil {
		log.Fatal("oops ?? ", err)
	}

	if index == END_OF_STRING || index == NULL_CHAR || index >= uint16(len(chars)) {
		// end of string
		return "", errors.New("invalid index")
	}
	
	return chars[index], nil
}