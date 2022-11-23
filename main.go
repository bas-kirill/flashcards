package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readUserInput(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	return line[2:]
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	_ = readUserInput(reader)
	definition := readUserInput(reader)
	answer := readUserInput(reader)
	if definition == answer {
		fmt.Println("right")
	} else {
		fmt.Println("wrong")
	}
}
