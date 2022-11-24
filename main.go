package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readUserInput(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	//fmt.Println(line)
	return line
}

func readCardsNumber(reader *bufio.Reader) (int, error) {
	fmt.Println("Input the number of cards:")
	inp := readUserInput(reader)
	cards, err := strconv.ParseInt(inp, 10, 32)
	if err != nil {
		return -1, err
	}
	return int(cards), nil
}

func readTerm(reader *bufio.Reader, idx int) string {
	fmt.Printf("The term for card #%d:\n", idx)
	term := readUserInput(reader)
	return term
}

func readDefinition(reader *bufio.Reader, idx int) string {
	fmt.Printf("The definition for card #%d:\n", idx)
	definition := readUserInput(reader)
	return definition
}

func readAnswer(reader *bufio.Reader, term string) string {
	fmt.Printf("Print the definition of \"%s\":\n", term)
	ans := readUserInput(reader)
	return ans
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	cards, _ := readCardsNumber(reader)
	terms := []string{}
	definitions := []string{}
	for i := 1; i <= cards; i++ {
		term := readTerm(reader, i)
		definition := readDefinition(reader, i)
		terms = append(terms, term)
		definitions = append(definitions, definition)
	}

	for i := 0; i < cards; i++ {
		term, def := terms[i], definitions[i]
		ans := readAnswer(reader, term)
		if ans == definitions[i] {
			fmt.Println("Correct!")
		} else {
			fmt.Printf("Wrong. The right answer is \"%s\".\n", def)
		}
	}
}
