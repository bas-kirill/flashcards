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

func contains(array []string, word string) bool {
	for _, elem := range array {
		if elem == word {
			return true
		}
	}
	return false
}

func readTerm(reader *bufio.Reader, terms []string, idx int) string {
	fmt.Printf("The term for card #%d:\n", idx)
	term := readUserInput(reader)
	for contains(terms, term) {
		fmt.Printf("The term \"%s\" already exists. Try again:\n", term)
		term = readUserInput(reader)
	}
	return term
}

func readDefinition(reader *bufio.Reader, definitions []string, idx int) string {
	fmt.Printf("The definition for card #%d:\n", idx)
	definition := readUserInput(reader)
	for contains(definitions, definition) {
		fmt.Printf("The definition \"%s\" already exists. Try again:\n", definition)
		definition = readUserInput(reader)
	}
	return definition
}

func readAnswer(reader *bufio.Reader, term string) string {
	fmt.Printf("Print the definition of \"%s\":\n", term)
	ans := readUserInput(reader)
	return ans
}

func appliedDefToAnotherTerm(
	terms []string,
	definitions []string,
	userDef string,
) (bool, string) {
	for i, elem := range definitions {
		if userDef == elem {
			return true, terms[i]
		}
	}
	return false, ""
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	cards, _ := readCardsNumber(reader)
	terms := []string{}
	definitions := []string{}
	for i := 1; i <= cards; i++ {
		term := readTerm(reader, terms, i)
		definition := readDefinition(reader, definitions, i)
		terms = append(terms, term)
		definitions = append(definitions, definition)
	}

	for i := 0; i < cards; i++ {
		term, def := terms[i], definitions[i]
		userDef := readAnswer(reader, term)
		if userDef == def {
			fmt.Println("Correct!")
		} else {
			ok, anotherTerm := appliedDefToAnotherTerm(terms, definitions, userDef)
			if ok {
				fmt.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", def, anotherTerm)
			} else {
				fmt.Printf("Wrong. The right answer is \"%s\".\n", def)
			}
		}
	}
}
