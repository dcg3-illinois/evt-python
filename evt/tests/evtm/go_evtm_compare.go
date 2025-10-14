// go_evtm_compare.go
// CLI tool to expose Go EventManager scenario operations for Python equivalence testing
package main

import (
	"fmt"
	"os"

	"github.com/iti/evt/evtm"
	"github.com/iti/evt/vrtime"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go_evtm_compare <scenario>")
		os.Exit(1)
	}
	fn := os.Args[1]
	mgr := evtm.New()

	switch fn {
	case "scenario_basic":
		// Schedule events
		test_slice := []int{}
		mgr.Schedule(&test_slice, 1, handler, vrtime.CreateTime(10, 1))
		eid2, _ := mgr.Schedule(&test_slice, 2, handler, vrtime.CreateTime(5, 2))
		mgr.Schedule(&test_slice, 3, handler, vrtime.CreateTime(5, 1))
		eid4, _ := mgr.Schedule(&test_slice, 4, handler, vrtime.CreateTime(15, 1))
		// Modify eid4 to occur earlier
		mgr.EventList.UpdateTime(eid4, vrtime.CreateTime(7, 1))
		// Cancel eid2
		mgr.CancelEvent(eid2)
		mgr.Run(20)
		// Output executed events in order: eid3, eid4, eid1
		fmt.Println(test_slice)
	case "scenario_cancel":
		// Schedule events
		test_slice := []int{}
		mgr.Schedule(&test_slice, 1, handler, vrtime.CreateTime(2, 1))
		mgr.Schedule(&test_slice, 2, handler, vrtime.CreateTime(4, 1))
		eid3, _ := mgr.Schedule(&test_slice, 3, handler, vrtime.CreateTime(6, 1))
		mgr.CancelEvent(eid3)
		mgr.Run(10)
		// print the contents of test_array
		fmt.Println(test_slice)
	default:
		fmt.Println("unknown scenario")
		os.Exit(1)
	}
}

func handler(mgr *evtm.EventManager, context any, data any) any {
	// append data to array pointed to by context
	context_array := context.(*[]int)
	*context_array = append(*context_array, data.(int))
	return nil
}
