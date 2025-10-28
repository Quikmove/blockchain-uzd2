package filetolist

import (
	"fmt"
	"log"
	"os"
)

func FileToList(path string) []string {
	data, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer data.Close()

	// read by line and append
	var list []string
	for {
		var line string
		_, err := fmt.Fscanln(data, &line)
		if err != nil {
			break
		}
		list = append(list, line)
	}
	return list
}
