package char_encoder

import (
	"encoding/json"
	"log"
	"os"
)

func Char(index uint16) string {
	file, err := os.ReadFile("char_encoder/table.json")
	if err != nil {
		log.Fatal("Error parsing char table file: ", err)
	}
	
	var chars []string
	err = json.Unmarshal(file, &chars)
	if err != nil {
		log.Fatal("oops ?? ", err)
	}
	
	return chars[index]
}