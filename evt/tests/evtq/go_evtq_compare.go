// go_evtq_compare.go
// CLI tool to expose Go EventQueue operations for Python equivalence testing
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iti/evt/evtq"
	"github.com/iti/evt/vrtime"
)

var queue *evtq.EventQueue

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go_evtq_compare <function> [args...]")
		os.Exit(1)
	}
	fn := os.Args[1]
	if queue == nil {
		queue = evtq.New()
	}
	switch fn {
	case "New":
		queue = evtq.New()
		fmt.Println("ok")
	case "Insert":
		if len(os.Args) < 5 {
			fmt.Println("usage: go_evtq_compare Insert <value> <timeTicks> <timePriority>")
			os.Exit(1)
		}
		val := os.Args[2]
		ticks, _ := strconv.ParseInt(os.Args[3], 10, 64)
		priority, _ := strconv.ParseInt(os.Args[4], 10, 64)
		time := vrtime.CreateTime(ticks, priority)
		evtID := queue.Insert(val, time)
		fmt.Printf("%d", evtID)
	case "Len":
		testLen()
	case "MinTime":
		testMinTime()
	case "Pop":
		testPop()
	case "UpdateTime":
		testUpdateTime()
	case "Remove":
		testRemove()
	case "GetItem":
		testGetItem()
	default:
		fmt.Println("unknown function")
		os.Exit(1)
	}

}

func testLen() {
	// Edge: empty
	q := evtq.New()
	fmt.Printf("empty:%d\n", q.Len())
	// Edge: single
	q = evtq.New()
	q.Insert("foo", vrtime.CreateTime(42, 7))
	fmt.Printf("single:%d\n", q.Len())
	// Normal: multiple (non-linear times)
	q = evtq.New()
	q.Insert("a", vrtime.CreateTime(15, 2))
	q.Insert("b", vrtime.CreateTime(3, 99))
	q.Insert("c", vrtime.CreateTime(27, 0))
	fmt.Printf("multiple:%d\n", q.Len())
}

func testMinTime() {
	// Edge: single
	q := evtq.New()
	q.Insert("foo", vrtime.CreateTime(42, 7))
	fmt.Printf("single:%d\n", q.MinTime().TickCnt)
	// Normal: multiple (non-linear times)
	q = evtq.New()
	q.Insert("a", vrtime.CreateTime(15, 2))
	q.Insert("b", vrtime.CreateTime(3, 99))
	q.Insert("c", vrtime.CreateTime(27, 0))
	fmt.Printf("multiple:%d\n", q.MinTime().TickCnt)
}

func testPop() {
	// Edge: single
	q := evtq.New()
	q.Insert("foo", vrtime.CreateTime(42, 7))
	val := q.Pop()
	fmt.Printf("single:%v\n", val)
	// Normal: multiple (non-linear times)
	q = evtq.New()
	q.Insert("a", vrtime.CreateTime(15, 2))
	q.Insert("b", vrtime.CreateTime(3, 99))
	q.Insert("c", vrtime.CreateTime(27, 0))
	val = q.Pop()
	fmt.Printf("multiple:%v\n", val)
}

func testUpdateTime() {
	q := evtq.New()
	_, idB, _ := func() (int, int, int) {
		idA := q.Insert("a", vrtime.CreateTime(15, 2))
		idB := q.Insert("b", vrtime.CreateTime(3, 99))
		idC := q.Insert("c", vrtime.CreateTime(27, 0))
		return idA, idB, idC
	}()
	// Update time of second item
	q.UpdateTime(idB, vrtime.CreateTime(99, 0))
	item := q.GetItem(idB)
	fmt.Printf("updated:%d\n", item)
	// Edge: update non-existent
	q.UpdateTime(9999, vrtime.CreateTime(42, 0))
	fmt.Printf("nonexistent:%v\n", q.GetItem(9999))
}

func testRemove() {
	q := evtq.New()
	_, idB, _ := func() (int, int, int) {
		idA := q.Insert("a", vrtime.CreateTime(15, 2))
		idB := q.Insert("b", vrtime.CreateTime(3, 99))
		idC := q.Insert("c", vrtime.CreateTime(27, 0))
		return idA, idB, idC
	}()
	ok := q.Remove(idB)
	fmt.Printf("removed:%t\n", ok)
	fmt.Printf("length after remove:%d\n", q.Len())
	ok = q.Remove(9999)
	fmt.Printf("nonexistent:%t\n", ok)
	fmt.Printf("length after nonexistent remove:%d\n", q.Len())
}

func testGetItem() {
	q := evtq.New()
	_, idB, _ := func() (int, int, int) {
		idA := q.Insert("a", vrtime.CreateTime(15, 2))
		idB := q.Insert("b", vrtime.CreateTime(3, 99))
		idC := q.Insert("c", vrtime.CreateTime(27, 0))
		return idA, idB, idC
	}()
	item := q.GetItem(idB)
	fmt.Printf("found:%v\n", item)
	item = q.GetItem(9999)
	fmt.Printf("notfound:%v\n", item)
}
