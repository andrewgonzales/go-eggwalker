package main

import (
	"fmt"
)

// Unique identifier per session (or device), not user
type ID struct {
	agent string
	seq   uint64
}

// UnicodeCharacter represents a single Unicode character.
type UnicodeCharacter string

// Special markers for beginning and end of document
type DocBeginning struct{}
type DocEnding struct{}

// OriginLeft is an interface for types that can be on the left side of an item.
type OriginLeft interface {
	isOriginLeft()
	isOrigin()
}

// OriginRight is an interface for types that can be on the right side of an item.
type OriginRight interface {
	isOriginRight()
	isOrigin()
}

type Origin interface {
	isOrigin()
}

func (ID) isOriginLeft()           {}
func (DocBeginning) isOriginLeft() {}

func (ID) isOriginRight()        {}
func (DocEnding) isOriginRight() {}

func (ID) isOrigin()           {}
func (DocBeginning) isOrigin() {}
func (DocEnding) isOrigin()    {}

type CRDTItem struct {
	content UnicodeCharacter

	id ID

	originLeft  OriginLeft
	originRight OriginRight

	deleted bool
}

// TODO: make a struct to remove need for returning append result and allow possibility of more metadata
type Doc struct {
	items []CRDTItem
}

// NewDoc creates a new Doc.
func NewDoc() Doc {
	return Doc{}
}

// StringContent returns the string content of the doc, omitting deleted contents.
func (d *Doc) StringContent() string {
	var result string
	for _, item := range d.items {
		if !item.deleted {
			result += string(item.content)
		}
	}

	return result
}

func (d *Doc) localInsertChar(char UnicodeCharacter, id ID, pos uint64) {
	originLeft := OriginLeft(DocBeginning{})
	if pos > 0 && pos-1 < uint64(len(d.items)) {
		originLeft = d.items[pos-1].id
	}

	originRight := OriginRight(DocEnding{})
	if pos < uint64(len(d.items)) {
		originRight = d.items[pos].id
	}

	item := CRDTItem{
		content:     char,
		id:          id,
		originLeft:  originLeft,
		originRight: originRight,
		deleted:     false,
	}

	fmt.Printf("item: %v\n", item)
	d.Integrate(item)
}

func (d *Doc) LocalInsertText(text string, id ID, pos uint64) {
	for _, r := range text {
		d.localInsertChar(UnicodeCharacter(r), id, pos)
	}
}

func (d *Doc) RemoteInsertChar(char UnicodeCharacter, item CRDTItem) {
	d.Integrate(item)
}

func idsEqual(id1, id2 ID) bool {
	return id1 == id2 || (id1.agent == id2.agent && id1.seq == id2.seq)
}

// For a real/long document, this should use a more efficient approach (e.g. B-tree)
func (d *Doc) findIndexById(id ID) int {
	for i, item := range d.items {
		if idsEqual(item.id, id) {
			return i
		}
	}
	return -1
}

func (d *Doc) findOriginIndex(origin Origin) int {
	switch o := origin.(type) {
	case DocBeginning:
		return -1
	case DocEnding:
		return len(d.items)
	case ID:
		return d.findIndexById(o)
	default:
		fmt.Println("Unknown type for origin: ", o)
		return -1
	}
}

func (d *Doc) Integrate(newItem CRDTItem) {
	newLeft := d.findOriginIndex(newItem.originLeft)
	newRight := d.findOriginIndex(newItem.originRight)
	destIndex := newLeft + 1
	// Scanning is when we might have a potential item to insert after, but need to look ahead to verify
	scanning := false

	// Look through the document to find the insertion point
	for i := destIndex; i <= newRight; i++ {
		if !scanning {
			destIndex = i
		}

		// If we reach the end of the document, just insert.
		// Also, if we reach the right index, there's no ambiguity frim concurrent editing, so just insert
		if i == newRight || i == len(d.items) {
			break
		}

		existingItem := d.items[i]

		existingLeft := d.findOriginIndex(existingItem.originLeft)
		existingRight := d.findOriginIndex(existingItem.originRight)

		if existingLeft < newLeft {
			// Base case, insert immediately
			break
		} else if existingLeft == newLeft {
			if existingRight < newRight {
				// This is tricky. We're looking at an item we *might* insert after - but we can't tell yet!
				scanning = true
				continue
			} else if existingRight == newRight {
				// Direct conflict. Use agent id as tiebreaker
				// TODO: could we add a timestamp to the item and use first-write-wins?
				if newItem.id.agent < existingItem.id.agent {
					break
				} else {
					scanning = false
					continue
				}

			} else {
				// existingRight > newRight
				// Reset scan and keep looking
				scanning = false
				continue
			}
		} else {
			// existingLeft > newLeft
			// Keep looking
			continue
		}
	}

	// We've found the position. Insert here.
	d.items = append(d.items[:destIndex], append([]CRDTItem{newItem}, d.items[destIndex:]...)...)
}

func main() {
	doc := NewDoc()

	doc.LocalInsertText("A", ID{"agent1", 1}, 0)
	doc.LocalInsertText("B", ID{"agent1", 2}, 1)
	doc.LocalInsertText("C", ID{"agent1", 3}, 0)

	fmt.Println("String content: ", doc.StringContent()) // Output: A
	fmt.Printf("Doc content: %v\n", doc.items)           // Output: A

}
