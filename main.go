package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

type Cards struct {
	TermToDef *OrderedMap[string, string]
	DefToTerm *OrderedMap[string, string]
}

func NewCards() *Cards {
	return &Cards{
		TermToDef: New[string, string](),
		DefToTerm: New[string, string](),
	}
}

type Card struct {
	Term       string `json:"term"`
	Definition string `json:"def"`
}

func readUserInput(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, "\n")
	return line
}

func tryAddCardTerm(cards *Cards, term string) bool {
	_, termPresent := cards.TermToDef.Get(term)
	if !termPresent {
		return true
	} else {
		fmt.Printf("The card \"%s\" already exists. Try again:\n", term)
		return false
	}
}

func tryAddCardDef(cards *Cards, def string) bool {
	_, defPresent := cards.DefToTerm.Get(def)
	if !defPresent {
		return true
	} else {
		fmt.Printf("The definition \"%s\" already exists. Try again:\n", def)
		return false
	}
}

func removeCard(cards *Cards, term string) bool {
	def, ok := cards.TermToDef.Get(term)
	if ok {
		cards.DefToTerm.Delete(def)
		cards.TermToDef.Delete(term)
		fmt.Println("The card has been removed.")
		return true
	} else {
		fmt.Printf("Can't remove \"%s\": there is no such card.\n", term)
		return false
	}
}

func importCards(file *os.File, cards *Cards) int {
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
		cards.DefToTerm.Set(card.Definition, card.Term)
		imported++
	}
	return imported
}

func exportCards(file *os.File, cards *Cards) int {
	exported := 0
	writer := bufio.NewWriter(file)
	for pair := cards.TermToDef.Oldest(); pair != nil; pair = pair.Next() {
		term, def := pair.Key, pair.Value
		card := Card{Term: term, Definition: def}
		cardJSON, err := json.Marshal(card)
		if err != nil {
			log.Fatal(err)
		}
		//_, _ = writer.WriteString(string(cardJSON))
		_, err = fmt.Fprintln(writer, string(cardJSON))
		if err != nil {
			log.Fatal(err)
		}
		//_ = writer.Flush()
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
		exported++
	}
	return exported
}

func readAsks() int {
	fmt.Println("How many times to ask?")
	var asks int
	_, err := fmt.Scan(&asks)
	if err != nil {
		log.Fatal(err)
	}
	return asks
}

func appliedDefToAnotherTerm(cards *Cards, userDef string) (bool, string) {
	for pair := cards.TermToDef.Oldest(); pair != nil; pair = pair.Next() {
		term, def := pair.Key, pair.Value
		if userDef == def {
			return true, term
		}
	}
	return false, ""
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	cards := NewCards()
	cmd := ""
	for cmd != "exit" {
		fmt.Println("Input the action (add, remove, import, export, ask, exit):")
		cmd = readUserInput(reader)

		switch cmd {
		case "add":
			fmt.Println("The card:")
			term := readUserInput(reader)
			termPresent := tryAddCardTerm(cards, term)
			for !termPresent {
				term = readUserInput(reader)
				termPresent = tryAddCardTerm(cards, term)
			}

			fmt.Println("The definition of the card:")
			def := readUserInput(reader)
			defPresent := tryAddCardDef(cards, def)
			for !defPresent {
				def = readUserInput(reader)
				defPresent = tryAddCardDef(cards, def)
			}

			cards.TermToDef.Set(term, def)
			cards.DefToTerm.Set(def, term)

			fmt.Printf("The pair (\"%s\":\"%s\") has been added.\n", term, def)
		case "remove":
			fmt.Println("Which card?")
			term := readUserInput(reader)
			removeCard(cards, term)
		case "import":
			fmt.Println("File name:")
			fileName := readUserInput(reader)
			file, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
			if err != nil {
				fmt.Println("File not found.")
				break
			}
			loadedCards := importCards(file, cards)
			fmt.Printf("%d cards have been loaded.\n", loadedCards)
		case "export":
			fmt.Println("File name:")
			fileName := readUserInput(reader)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			exportedCards := exportCards(file, cards)
			fmt.Printf("%d cards have been saved.\n", exportedCards)
		case "ask":
			asks := readAsks()
			idx := 0
			for pair := cards.TermToDef.Oldest(); idx < asks; pair, idx = pair.Next(), idx+1 {
				if pair == nil {
					pair = cards.TermToDef.Oldest()
				}
				term, def := pair.Key, pair.Value
				fmt.Printf("Print the definition of \"%s\":\n", term)
				userDef := readUserInput(reader)
				if userDef == def {
					fmt.Println("Correct!")
				} else {
					ok, anotherTerm := appliedDefToAnotherTerm(cards, userDef)
					if ok {
						fmt.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", def, anotherTerm)
					} else {
						fmt.Printf("Wrong. The right answer is \"%s\".\n", def)
					}
				}
			}
		case "exit":
			fmt.Print("Bye bye!")
			os.Exit(0)
		}

		fmt.Println()
	}
}
