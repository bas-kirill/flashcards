package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type List[T any] struct {
	root Element[T] // sentinel list element, only &root, root.prev, and root.next are used
	len  int        // current list length excluding (this) sentinel element
}

// Element is an element of a linked list.
type Element[T any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[T]

	// The list to which this element belongs.
	list *List[T]

	// The value stored with this element.
	Value T
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V

	element *Element[*Pair[K, V]]
}

type OrderedMap[K comparable, V any] struct {
	pairs map[K]*Pair[K, V]
	list  *List[*Pair[K, V]]
}

func NewList[T any]() *List[T] { return new(List[T]).Init() }

// New creates a new OrderedMap.
func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		pairs: make(map[K]*Pair[K, V]),
		list:  NewList[*Pair[K, V]](),
	}
}

func (l *List[T]) Init() *List[T] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func (l *List[T]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List[T]) insert(e, at *Element[T]) *Element[T] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *List[T]) insertValue(v T, at *Element[T]) *Element[T] {
	return l.insert(&Element[T]{Value: v}, at)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[T]) PushBack(v T) *Element[T] {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// Get looks for the given key, and returns the value associated with it,
// or V's nil value if not found. The boolean it returns says whether the key is present in the map.
func (om *OrderedMap[K, V]) Get(key K) (val V, present bool) {
	if pair, present := om.pairs[key]; present {
		return pair.Value, true
	}

	return
}

// Set sets the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Set`.
func (om *OrderedMap[K, V]) Set(key K, value V) (val V, present bool) {
	if pair, present := om.pairs[key]; present {
		oldValue := pair.Value
		pair.Value = value
		return oldValue, true
	}

	pair := &Pair[K, V]{
		Key:   key,
		Value: value,
	}
	pair.element = om.list.PushBack(pair)
	om.pairs[key] = pair

	return
}

// remove removes e from its list, decrements l.len
func (l *List[T]) remove(e *Element[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *List[T]) Remove(e *Element[T]) T {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// Delete removes the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Delete`.
func (om *OrderedMap[K, V]) Delete(key K) (val V, present bool) {
	if pair, present := om.pairs[key]; present {
		om.list.Remove(pair.element)
		delete(om.pairs, key)
		return pair.Value, true
	}
	return
}

func listElementToPair[K comparable, V any](element *Element[*Pair[K, V]]) *Pair[K, V] {
	if element == nil {
		return nil
	}
	return element.Value
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[T]) Front() *Element[T] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[T]) Back() *Element[T] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// Oldest returns a pointer to the oldest pair. It's meant to be used to iterate on the ordered map's
// pairs from the oldest to the newest, e.g.:
// for pair := orderedMap.Oldest(); pair != nil; pair = pair.Next() { fmt.Printf("%v => %v\n", pair.Key, pair.Value) }
func (om *OrderedMap[K, V]) Oldest() *Pair[K, V] {
	return listElementToPair(om.list.Front())
}

// Next returns the next list element or nil.
func (e *Element[T]) Next() *Element[T] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Next returns a pointer to the next pair.
func (p *Pair[K, V]) Next() *Pair[K, V] {
	return listElementToPair(p.element.Next())
}

type TermError struct {
	Term   string
	Errors int
}

type Cards struct {
	TermToDef *OrderedMap[string, string]
	DefToTerm *OrderedMap[string, TermError]
}

func NewCards() *Cards {
	return &Cards{
		TermToDef: New[string, string](),
		DefToTerm: New[string, TermError](),
	}
}

type Card struct {
	Term       string `json:"term"`
	Definition string `json:"def"`
	ErrorCount int    `json:"errors"`
}

var logger *List[string]

func ReadUserInput(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, "\n")
	return line
}

func TryAddCardTerm(cards *Cards, term string) bool {
	_, termPresent := cards.TermToDef.Get(term)
	if !termPresent {
		return true
	} else {
		fmt.Printf("The card \"%s\" already exists. Try again:\n", term)
		logger.PushBack(fmt.Sprintf("The card \"%s\" already exists. Try again:\n", term))
		return false
	}
}

func TryAddCardDef(cards *Cards, def string) bool {
	_, defPresent := cards.DefToTerm.Get(def)
	if !defPresent {
		return true
	} else {
		fmt.Printf("The definition \"%s\" already exists. Try again:\n", def)
		//cards.DefToTerm.Set(def, TermError{termErr.Term, termErr.Errors + 1})
		logger.PushBack(fmt.Sprintf("The definition \"%s\" already exists. Try again:\n", def))
		return false
	}
}

func RemoveCard(cards *Cards, term string) bool {
	def, ok := cards.TermToDef.Get(term)
	if ok {
		cards.DefToTerm.Delete(def)
		cards.TermToDef.Delete(term)
		fmt.Println("The card has been removed.")
		logger.PushBack("The card has been removed.")
		return true
	} else {
		fmt.Printf("Can't remove \"%s\": there is no such card.\n", term)
		logger.PushBack(fmt.Sprintf("Can't remove \"%s\": there is no such card.\n", term))
		return false
	}
}

func ImportCards(file *os.File, cards *Cards) int {
	imported := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		card := Card{}
		err := json.Unmarshal(line, &card)
		if err != nil {
			log.Fatal(err)
		}
		cards.TermToDef.Set(card.Term, card.Definition)
		//fmt.Println(card.Term, card.Definition, card.ErrorCount)
		cards.DefToTerm.Set(card.Definition, TermError{card.Term, card.ErrorCount})
		imported++
	}
	return imported
}

func ExportCards(file *os.File, cards *Cards) int {
	defer file.Close()
	exported := 0
	writer := bufio.NewWriter(file)
	for pair := cards.TermToDef.Oldest(); pair != nil; pair = pair.Next() {
		term, def := pair.Key, pair.Value
		errors, _ := cards.DefToTerm.Get(def)
		card := Card{Term: term, Definition: def, ErrorCount: errors.Errors}
		cardJSON, err := json.Marshal(card)
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Fprintln(writer, string(cardJSON))
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
		exported++
	}
	return exported
}

func ReadAsks() int {
	fmt.Println("How many times to ask?")
	logger.PushBack("How many times to ask?")
	var asks int
	_, err := fmt.Scan(&asks)
	if err != nil {
		log.Fatal(err)
	}
	return asks
}

func ApplyDefToAnotherTerm(cards *Cards, userDef string) (bool, string) {
	for pair := cards.TermToDef.Oldest(); pair != nil; pair = pair.Next() {
		term, def := pair.Key, pair.Value
		if userDef == def {
			return true, term
		}
	}
	return false, ""
}

func SaveLog(file *os.File) {
	fmt.Println("kek")
	writer := bufio.NewWriter(file)
	for elem := logger.Front(); elem != logger.Back().next; elem = elem.next {
		fmt.Println(elem.Value)
		_, err := fmt.Fprintln(writer, elem.Value)
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func HardestCard(cards *Cards) string {
	term := ""
	mxErr := -1
	var terms []string
	for pair := cards.DefToTerm.Oldest(); pair != nil; pair = pair.Next() {
		termError := pair.Value
		if termError.Errors > mxErr {
			mxErr = termError.Errors
			term = termError.Term
			terms = []string{term}
		} else if termError.Errors == mxErr {
			terms = append(terms, term)
		}
	}

	if mxErr == 0 || cards.DefToTerm.list.len == 0 {
		return "There are no cards with errors."
	} else if len(terms) == 1 {
		return fmt.Sprintf("The hardest card is \"%s\". You have %d errors answering it", term, mxErr)
	} else if len(terms) > 1 {
		ans := ""
		first := true
		for t := range terms {
			if !first {
				ans += ", "
			}
			first = false
			ans += fmt.Sprintf("\"%s\"", t)
		}
		return fmt.Sprintf("The hardest cards are %s", ans)
	}
	return "-1"
}

func main() {
	logger = NewList[string]()
	reader := bufio.NewReader(os.Stdin)
	cards := NewCards()
	cmd := ""
	for cmd != "exit" {
		fmt.Println("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
		logger.PushBack("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")

		cmd = ReadUserInput(reader)
		logger.PushBack(cmd)

		switch cmd {
		case "add":
			fmt.Println("The card:")
			logger.PushBack("The card:")

			term := ReadUserInput(reader)
			logger.PushBack(term)

			termPresent := TryAddCardTerm(cards, term)
			for !termPresent {
				term = ReadUserInput(reader)
				logger.PushBack(term)
				termPresent = TryAddCardTerm(cards, term)
			}

			fmt.Println("The definition of the card:")
			logger.PushBack("The definition of the card:")

			def := ReadUserInput(reader)
			logger.PushBack(def)
			defPresent := TryAddCardDef(cards, def)
			for !defPresent {
				def = ReadUserInput(reader)
				logger.PushBack(def)
				defPresent = TryAddCardDef(cards, def)
			}

			cards.TermToDef.Set(term, def)
			cards.DefToTerm.Set(def, TermError{term, 0})

			fmt.Printf("The pair (\"%s\":\"%s\") has been added.\n", term, def)
			logger.PushBack(fmt.Sprintf("The pair (\"%s\":\"%s\") has been added.", term, def))
		case "remove":
			fmt.Println("Which card?")
			logger.PushBack("Which card?")
			term := ReadUserInput(reader)
			logger.PushBack(term)
			RemoveCard(cards, term)
		case "import":
			fmt.Println("File name:")
			logger.PushBack("File name:")
			fileName := ReadUserInput(reader)
			logger.PushBack(fileName)
			file, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
			if err != nil {
				fmt.Println("File not found.")
				logger.PushBack("File not found.")
				break
			}
			loadedCards := ImportCards(file, cards)
			fmt.Printf("%d cards have been loaded.\n", loadedCards)
			logger.PushBack(fmt.Sprintf("%d cards have been loaded.", loadedCards))
		case "export":
			fmt.Println("File name:")
			logger.PushBack("File name:")
			fileName := ReadUserInput(reader)
			logger.PushBack(fileName)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			exportedCards := ExportCards(file, cards)
			fmt.Printf("%d cards have been saved.\n", exportedCards)
			logger.PushBack(fmt.Sprintf("%d cards have been saved.", exportedCards))
		case "ask":
			asks := ReadAsks()
			logger.PushBack(strconv.FormatInt(int64(asks), 10))
			idx := 0
			for pair := cards.TermToDef.Oldest(); idx < asks; pair, idx = pair.Next(), idx+1 {
				if pair == nil {
					pair = cards.TermToDef.Oldest()
				}
				term, def := pair.Key, pair.Value
				fmt.Printf("Print the definition of \"%s\":\n", term)
				logger.PushBack(fmt.Sprintf("Print the definition of \"%s\":", term))

				userDef := ReadUserInput(reader)
				logger.PushBack(userDef)

				if userDef == def {
					fmt.Println("Correct!")
					logger.PushBack("Correct!")
				} else {
					ok, anotherTerm := ApplyDefToAnotherTerm(cards, userDef)
					if ok {
						fmt.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", def, anotherTerm)
						logger.PushBack(fmt.Sprintf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".", def, anotherTerm))
					} else {
						fmt.Printf("Wrong. The right answer is \"%s\".\n", def)
						logger.PushBack(fmt.Sprintf("Wrong. The right answer is \"%s\".", def))
					}
					termErr, _ := cards.DefToTerm.Get(def)
					cards.DefToTerm.Set(def, TermError{termErr.Term, termErr.Errors + 1})
				}
			}
		case "exit":
			fmt.Print("Bye bye!")
			logger.PushBack("Bye bye!")
			os.Exit(0)
		case "log":
			fmt.Println("File name:")
			logger.PushBack("File name:")
			fileName := ReadUserInput(reader)
			logger.PushBack(fileName)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("The log has been saved.")
			logger.PushBack("The log has been saved.")
			SaveLog(file)
		case "hardest card":
			ans := HardestCard(cards)
			fmt.Println(ans)
			logger.PushBack(ans)
		case "reset stats":
			for pair := cards.DefToTerm.Oldest(); pair != nil; pair = pair.Next() {
				cards.DefToTerm.Set(pair.Key, TermError{Term: pair.Value.Term, Errors: 0})
			}
			fmt.Println("Card statistics have been reset.")
			logger.PushBack("Card statistics have been reset.")
		}

		fmt.Println()
		logger.PushBack("")
	}
}
