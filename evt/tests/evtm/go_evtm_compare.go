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
		eid1, _ := mgr.Schedule("ctx1", "dat1", handler, vrtime.CreateTime(10, 1))
		eid2, _ := mgr.Schedule("ctx2", "dat2", handler, vrtime.CreateTime(5, 2))
		eid3, _ := mgr.Schedule("ctx3", "dat3", handler, vrtime.CreateTime(5, 1))
		eid4, _ := mgr.Schedule("ctx4", "dat4", handler, vrtime.CreateTime(15, 1))
		// Modify eid4 to occur earlier
		mgr.EventList.UpdateTime(eid4, vrtime.CreateTime(7, 1))
		// Cancel eid2
		mgr.CancelEvent(eid2)
		mgr.Run(20)
		// Output executed events in order: eid3, eid4, eid1
		for _, eid := range []int{eid3, eid4, eid1} {
			item := mgr.EventList.GetItem(eid)
			if item != nil {
				evt := item.(*evtm.Event)
				fmt.Printf("%s,%s\n", evt.Context, evt.Data)
			}
		}
	case "scenario_cancel":
		// Schedule events
		eid1, _ := mgr.Schedule("A", "X", handler, vrtime.CreateTime(2, 1))
		eid2, _ := mgr.Schedule("B", "Y", handler, vrtime.CreateTime(4, 1))
		eid3, _ := mgr.Schedule("C", "Z", handler, vrtime.CreateTime(6, 1))
		mgr.CancelEvent(eid3)
		mgr.Run(10)
		// Output executed events in order: eid2, eid1
		for _, eid := range []int{eid2, eid1} {
			item := mgr.EventList.GetItem(eid)
			if item != nil {
				evt := item.(*evtm.Event)
				fmt.Printf("%s,%s\n", evt.Context, evt.Data)
			}
		}
		// Output cancellation status for eid3
		item3 := mgr.EventList.GetItem(eid3)
		if item3 != nil {
			evt := item3.(*evtm.Event)
			fmt.Printf("cancelled:%v\n", evt.Cancel)
		}
	default:
		fmt.Println("unknown scenario")
		os.Exit(1)
	}
}

func handler(mgr *evtm.EventManager, context any, data any) any {
	return nil
}
