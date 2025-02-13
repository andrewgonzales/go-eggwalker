package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// Unique identifier per session (or device), not user
type ID struct {
	agent string
	seq   uint64
}

// UnicodeCharacter represents a single Unicode character.
type UnicodeCharacter string

// NewUnicodeCharacter creates a new UnicodeCharacter if the input is a single character.
func NewUnicodeCharacter(s string) (UnicodeCharacter, error) {
	if utf8.RuneCountInString(s) != 1 {
		return "", errors.New("input must be a single Unicode character")
	}
	return UnicodeCharacter(s), nil
}

// Special markers for beginning and end of document
type DocBeginning struct{}
type DocEnding struct{}

// OriginLeft is an interface for types that can be on the left side of an item.
type OriginLeft interface {
	isOriginLeft()
}

// OriginRight is an interface for types that can be on the right side of an item.
type OriginRight interface {
	isOriginRight()
}

func (ID) isOriginLeft()           {}
func (DocBeginning) isOriginLeft() {}

func (ID) isOriginRight()        {}
func (DocEnding) isOriginRight() {}

type CRDTItem struct {
	content UnicodeCharacter

	id ID

	originLeft  OriginLeft
	originRight OriginRight

	deleted bool
}

// TODO: make a struct to remove need for returning append result and allow possibility of more metadata
type Doc []CRDTItem

// NewDoc creates a new Doc.
func NewDoc() Doc {
	return make(Doc, 0)
}

// StringContent returns the string content of the doc, omitting deleted contents.
func (d Doc) StringContent() string {
	var result string
	for _, item := range d {
		if !item.deleted {
			result += string(item.content)
		}
	}
	return result
}

func main() {
	doc := NewDoc()

	char, err := NewUnicodeCharacter("A")
	if err != nil {
		fmt.Println("Error creating UnicodeCharacter:", err)
		panic(err)
	}

	doc = append(doc, CRDTItem{
		content: char,
		id:      ID{agent: "agent1", seq: 1},
	})

	fmt.Println("Content: ", doc.StringContent()) // Output: A

}
