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

type Version map[string]uint64

// TODO: make a struct to remove need for returning append result and allow possibility of more metadata
type Doc struct {
	items []CRDTItem

	version Version
}

// NewDoc creates a new Doc.
func NewDoc() Doc {
	return Doc{version: make(map[string]uint64)}
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

func (d *Doc) localInsertChar(char UnicodeCharacter, agent string, pos uint64) {
	seq := d.version[agent] + 1
	id := ID{agent, seq}

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

	d.Integrate(item)
}

func (d *Doc) LocalInsertText(text, agent string, pos uint64) {
	for _, r := range text {
		d.localInsertChar(UnicodeCharacter(r), agent, pos)
	}
}

func (d *Doc) RemoteInsertItem(item CRDTItem) {
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
	// Check for operation order
	lastSeenSeq := d.version[newItem.id.agent]
	newSeq := newItem.id.seq
	if newSeq != lastSeenSeq+1 {
		panic("Error: operations out of order")
	}

	// Update the document version
	d.version[newItem.id.agent] = newSeq

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
		// Also, if we reach the right index, there's no ambiguity from concurrent editing, so just insert
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

func (d *Doc) isInVersion(id ID) bool {
	highestSeq := d.version[id.agent]

	return highestSeq >= id.seq
}

func (d *Doc) canInsert(item CRDTItem) bool {
	// originLeft and originRight both need to be in the doc
	var leftExists, rightExists bool
	switch item.originLeft.(type) {
	case DocBeginning:
		leftExists = true
	case ID:
		originLeftID := item.originLeft.(ID)
		leftExists = d.isInVersion(originLeftID)
	}

	switch item.originRight.(type) {
	case DocEnding:
		rightExists = true
	case ID:
		originRightID := item.originRight.(ID)
		rightExists = d.isInVersion(originRightID)
	}

	// Can insert the first sequence from an agent, or if the previous sequence is in the doc
	isCorrectOrder := item.id.seq == 0 || d.isInVersion(ID{item.id.agent, item.id.seq - 1})

	// Can insert if the item is not already in the doc and all the prerequisite items are
	return !d.isInVersion(item.id) && isCorrectOrder && leftExists && rightExists
}

// Merge a document into a destination
// This function is idempotent
func (dest *Doc) MergeInto(src *Doc) error {
	toBeInserted := make(map[ID]CRDTItem)

	for _, item := range src.items {
		if !dest.isInVersion(item.id) {
			toBeInserted[item.id] = item
		}
	}

	numRemaining := len(toBeInserted)

	for numRemaining > 0 {
		// Try to merge something on every pass
		mergedOnThisPass := 0
		inserted := make(map[ID]bool)

		for id, item := range toBeInserted {
			// Keep going until we find something mergeable
			if !dest.canInsert(item) {
				continue
			}

			// Merge item
			dest.RemoteInsertItem(item)
			// Add to an inserted map to avoid modifying the toBeInserted map in the loop
			inserted[id] = true

			// Update control variables
			mergedOnThisPass += 1
			numRemaining -= 1
		}

		if mergedOnThisPass == 0 {
			return fmt.Errorf("Error: Not making progress")
		}

		for id := range inserted {
			delete(toBeInserted, id)
		}

	}

	return nil
}

func main() {
	doc1 := NewDoc()
	doc2 := NewDoc()

	doc1.LocalInsertText("A", "agent1", 0)
	doc2.LocalInsertText("B", "agent2", 0)

	doc1.MergeInto(&doc2)
	doc2.MergeInto(&doc1)

	fmt.Println("Doc1 String content: ", doc1.StringContent())
	fmt.Printf("Doc1 content: %v\n", doc1.items)

	doc2.MergeInto(&doc1)
	fmt.Println("Doc2 String content: ", doc2.StringContent())

}
