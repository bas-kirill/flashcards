package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Card struct {
	Term       string `json:"term"`
	Definition string `json:"def"`
}

func readUserInput(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	return line
}

func readCard(reader *bufio.Reader) (string, string) {
	fmt.Println("The card:")
	term, _ := reader.ReadString('\n')
	fmt.Println("The definition of the card:")
	def, _ := reader.ReadString('\n')
	return term, def
}

func addCard(termToDef map[string]string, term string, def string) bool {
	_, ok := termToDef[term]

	if ok {
		fmt.Printf("This <%s/%s> already exists. Try again:\n", term, def)
		return false
	} else {
		termToDef[term] = def
		fmt.Printf("The pair (\"%s\":\"%s\") has been added\n", term, def)
		return true
	}
}

func removeCard(termToDef map[string]string, term string) bool {
	_, ok := termToDef[term]
	if ok {
		delete(termToDef, term)
		fmt.Println("The card has been removed.")
		return true
	} else {
		fmt.Printf("Can't remove \"%s\": there is no such card.", term)
		return false
	}
}

func readFileName(reader *bufio.Reader) string {
	fmt.Println("File name:")
	fileName := readUserInput(reader)
	return fileName
}

func importCards(fileName string, termToDef map[string]string) (int, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		return 0, err
	}
	importedCardsCnt := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		card := Card{}
		err = json.Unmarshal(line, &card)
		if err != nil {
			log.Fatal(err)
		}
		termToDef[card.Term] = card.Definition
		importedCardsCnt++
	}
	return importedCardsCnt, nil
}

func exportCards(fileName string, termToDef map[string]string) (int, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0222)
	if err != nil {
		return 0, err
	}
	exportedCards := 0
	writer := bufio.NewWriter(file)
	for term, def := range termToDef {
		card := Card{Term: term, Definition: def}
		cardJSON, err := json.Marshal(card)
		if err != nil {
			return 0, err
		}
		_, err = fmt.Fprintln(writer, cardJSON)
		if err != nil {
			return 0, err
		}
		exportedCards++
	}
	return exportedCards, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	termToDef := map[string]string{}
	for cmd := readUserInput(reader); cmd != "exit"; {
		fmt.Println("Input the action (add, remove, import, export, ask, exit):")

		switch cmd {
		case "add":
			term, def := readCard(reader)
			for ok := addCard(termToDef, term, def); !ok; {
			}
		case "remove":
			fmt.Println("Which card?")
			term := readUserInput(reader)
			removeCard(termToDef, term)
		case "import":
			fileName := readFileName(reader)
			loadedCards, err := importCards(fileName, termToDef)
			if err != nil {
				fmt.Println("File not found.")
			} else {
				fmt.Printf("%d cards have been loaded.\n", loadedCards)
			}
		case "export":
			fileName := readFileName(reader)
			exportedCards, err := exportCards(fileName, termToDef)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%d cards have been saved.\n", exportedCards)
		case "ask":
			fmt.Println("How many times to ask?")
			var questions int
			_, err := fmt.Scan(&questions)
			if err != nil {
				log.Fatal(err)
			}
			// TODO: we need create navigation map for traverse through insertion order
			for i := 0; i < questions; i++ {
			}
		case "exit":
			break
		}
	}
}
