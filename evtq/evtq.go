// Package evtq creates and manages event queues
//
// It depends upon [container/heap] to manage a heap structure
package evtq

import (
	"container/heap"
	"github.com/iti/evt/vrtime"
	"sync"
)

// InvalidEventID will never be returned from
// EventQueue.Insert()
const InvalidEventID = 0

// EventQueue represents the queue
type EventQueue struct {
	evtID    int           // monotonically increasing counter used for default secondary time in event Time
	itemHeap *itemHeapType // data structure holding items, see struct definition for item and itemHeapType
	lookup   map[int]*item // event identifier to event, used for marking events to be ignored
	MaxTime  vrtime.Time   // Largest vrtime.Time value pushed onto to the heap as yet
	mu       sync.Mutex    // used to support thread safety
}

// New is a constructor. Initializes an empty slice of events
func New() *EventQueue {
	return &EventQueue{
		evtID:    InvalidEventID,      // has to have an event id, so include an invalid one at initialization
		itemHeap: &itemHeapType{},     // event list is initialized to be empty of events
		lookup:   make(map[int]*item)} // map to support deletion of events is initially empty
}

// Len returns the number of elements in the queue.
func (p *EventQueue) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	rtn := p.itemHeap.Len()
	return rtn
}

// MinTime returns the Time associated with the next event.
func (p *EventQueue) MinTime() vrtime.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	rtn := (*p.itemHeap)[0].Time
	return rtn
}

// Insert inserts a new element into the queue. No action is performed on duplicate elements.
func (p *EventQueue) Insert(v any, time vrtime.Time) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.evtID++

	// update maximum time of inserted event
	if p.MaxTime.LT(time) {
		p.MaxTime = time
	}

	// if the priority in the time stamp is -1
	// change it to be something monotonically increasing
	// to try to give some determinism for events with the
	// same tick count when some other priority isn't given
	//
	if time.Pri() == -1 {
		time.SetPri(int64(p.evtID))
	}

	// create an item for insertion
	newItem := &item{
		itemID: p.evtID, // identifier for this event
		Value:  v,       // notice that v can be anything, what matters for ordering is time value
		Time:   time}

	heap.Push(p.itemHeap, newItem)
	p.lookup[p.evtID] = newItem
	rtn := p.evtID
	return rtn
}

// Pop removes the element with the least time from the queue and returns it.
// In case of an empty queue, an error is returned.
func (p *EventQueue) Pop() any {
	p.mu.Lock()
	defer p.mu.Unlock()

	popped := heap.Pop(p.itemHeap).(*item)
	delete(p.lookup, popped.itemID)
	rtn := popped.Value
	return rtn
}

// UpdateTime changes the priority of a given item.
// If the specified item is not present in the queue, no action is performed.
func (p *EventQueue) UpdateTime(evtID int, newTime vrtime.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	item, present := p.lookup[evtID]

	if !present {
		return
	}

	item.Time = newTime
	heap.Fix(p.itemHeap, item.index)
}

func (p *EventQueue) GetItem(evtID int) any {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, present := p.lookup[evtID]
	if !present {
		return nil
	}
	return p.lookup[evtID]
}


// Remove an element. Returns true on success.
func (p *EventQueue) Remove(evtID int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	element, present := p.lookup[evtID]
	if !present {
		return false
	}

	// we're going to push the element to be rid of to the top
	element.Time = vrtime.ZeroTime()
	heap.Fix(p.itemHeap, element.index)

	// pop it off
	popped := heap.Pop(p.itemHeap).(*item)
	delete(p.lookup, popped.itemID)
	return true
}

// itemHeapType is the type of the data structure written to satisfy the Heap interface
type itemHeapType []*item

// The item struct defines the item organized by Time value
type item struct {
	itemID int         // unique identifier with every event inserted into the queue
	Value  any         // completely general payload for the item
	Time   vrtime.Time // the field used to order the elements
	index  int         // the position of the item in the (heap-organized) slice of events
	Cancel bool        // has been marked for removal 
}

// Len, Less, Swap, Push, and Pop are funcs required for a
// a type to satisfy the Heap interface

// Len returns the number of items in the slice of events
func (ih *itemHeapType) Len() int {
	return len(*ih)
}

// Less with arguments i, j returns true if the item in the priority queue in position i
// has an earlier time-stamp than the item in position j
func (ih *itemHeapType) Less(i, j int) bool {
	return (*ih)[i].Time.LT((*ih)[j].Time)
}

// Swap with arguments i, j exchanges the items in positions i and j
func (ih *itemHeapType) Swap(i, j int) {
	(*ih)[i], (*ih)[j] = (*ih)[j], (*ih)[i]
	(*ih)[i].index = i
	(*ih)[j].index = j
}

// Push puts the item given as an argument at the end of the slice of events
// This is not necessarily its final position in the heap
func (ih *itemHeapType) Push(x any) {
	it := x.(*item)
	it.index = len(*ih)
	*ih = append(*ih, it)
}

// Pop removes the item in the heap which appears in the last position of the heap
func (ih *itemHeapType) Pop() any {
	old := *ih
	item := old[len(old)-1]
	*ih = old[0 : len(old)-1]
	return item
}
