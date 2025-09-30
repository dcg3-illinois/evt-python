// go_vrtime_compare.go
// This Go program exposes vrtime.go functions for CLI comparison with Python
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iti/evt/vrtime"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: go_vrtime_compare <function> <value> [<priority>]")
		os.Exit(1)
	}
	fn := os.Args[1]
	val := os.Args[2]
	switch fn {
	case "TicksToSeconds":
		ticks, _ := strconv.ParseInt(val, 10, 64)
		fmt.Printf("%.12f", vrtime.TicksToSeconds(ticks))
	case "SecondsToTicks":
		sec, _ := strconv.ParseFloat(val, 64)
		fmt.Printf("%d", vrtime.SecondsToTicks(sec))
	case "MuSecondsToTicks":
		mus, _ := strconv.ParseFloat(val, 64)
		fmt.Printf("%d", vrtime.MuSecondsToTicks(mus))
	case "SetTicksPerSecond":
		tps, _ := strconv.ParseInt(val, 10, 64)
		ok := vrtime.SetTicksPerSecond(tps)
		if !ok {
			fmt.Println("invalid TicksPerSecond value")
			os.Exit(1)
		}
		fmt.Printf("%d", vrtime.TicksPerSecond)
	case "TimeToSeconds":
		// usage: go_vrtime_compare TimeToSeconds <tick> <priority>
		if len(os.Args) < 4 {
			fmt.Println("need <priority> argument")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(val, 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%.12f", vrtime.TimeToSeconds(t))
	case "CreateTime":
		// usage: go_vrtime_compare CreateTime <tick> <priority>
		if len(os.Args) < 4 {
			fmt.Println("need <priority> argument")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(val, 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.CreateTime(tick, pri)
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "InfinityTime":
		t := vrtime.InfinityTime()
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "ZeroTime":
		t := vrtime.ZeroTime()
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "SecondsToTime":
		sec, _ := strconv.ParseFloat(val, 64)
		t := vrtime.SecondsToTime(sec)
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "SecondsToTimePri":
		if len(os.Args) < 4 {
			fmt.Println("need <priority> argument")
			os.Exit(1)
		}
		sec, _ := strconv.ParseFloat(val, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.SecondsToTimePri(sec, pri)
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "CmpTime":
		// usage: go_vrtime_compare CmpTime <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		cmp := 0
		if t1.LT(t2) {
			cmp = -1
		} else if t1.GT(t2) {
			cmp = 1
		} else if t1.EQ(t2) {
			cmp = 0
		}
		fmt.Printf("%d", cmp)
	case "EQ":
		// usage: go_vrtime_compare EQ <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.EQ(t2))
	case "GE":
		// usage: go_vrtime_compare GE <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.GE(t2))
	case "GT":
		// usage: go_vrtime_compare GT <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.GT(t2))
	case "LE":
		// usage: go_vrtime_compare LE <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.LE(t2))
	case "LT":
		// usage: go_vrtime_compare LT <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.LT(t2))
	case "NEQ":
		// usage: go_vrtime_compare NEQ <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		fmt.Printf("%t", t1.NEQ(t2))
	case "Plus":
		// usage: go_vrtime_compare Plus <tick1> <pri1> <tick2> <pri2>
		if len(os.Args) < 6 {
			fmt.Println("need <tick1> <pri1> <tick2> <pri2>")
			os.Exit(1)
		}
		tick1, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri1, _ := strconv.ParseInt(os.Args[3], 10, 64)
		tick2, _ := strconv.ParseInt(os.Args[4], 10, 64)
		pri2, _ := strconv.ParseInt(os.Args[5], 10, 64)
		t1 := vrtime.Time{TickCnt: tick1, Priority: pri1}
		t2 := vrtime.Time{TickCnt: tick2, Priority: pri2}
		t3 := t1.Plus(t2)
		fmt.Printf("%d,%d", t3.TickCnt, t3.Priority)
	case "Pri":
		// usage: go_vrtime_compare Pri <tick> <pri>
		if len(os.Args) < 4 {
			fmt.Println("need <tick> <pri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%d", t.Pri())
	case "Seconds":
		// usage: go_vrtime_compare Seconds <tick> <pri>
		if len(os.Args) < 4 {
			fmt.Println("need <tick> <pri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%.12f", t.Seconds())
	case "SecondsStr":
		// usage: go_vrtime_compare SecondsStr <tick> <pri>
		if len(os.Args) < 4 {
			fmt.Println("need <tick> <pri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%s", t.SecondsStr())
	case "SetPri":
		// usage: go_vrtime_compare SetPri <tick> <pri> <newpri>
		if len(os.Args) < 5 {
			fmt.Println("need <tick> <pri> <newpri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		newpri, _ := strconv.ParseInt(os.Args[4], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		t.SetPri(newpri)
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "SetTicks":
		// usage: go_vrtime_compare SetTicks <tick> <pri> <newtick>
		if len(os.Args) < 5 {
			fmt.Println("need <tick> <pri> <newtick>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		newtick, _ := strconv.ParseInt(os.Args[4], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		t.SetTicks(newtick)
		fmt.Printf("%d,%d", t.TickCnt, t.Priority)
	case "Ticks":
		// usage: go_vrtime_compare Ticks <tick> <pri>
		if len(os.Args) < 4 {
			fmt.Println("need <tick> <pri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%d", t.Ticks())
	case "TimeStr":
		// usage: go_vrtime_compare TimeStr <tick> <pri>
		if len(os.Args) < 4 {
			fmt.Println("need <tick> <pri>")
			os.Exit(1)
		}
		tick, _ := strconv.ParseInt(os.Args[2], 10, 64)
		pri, _ := strconv.ParseInt(os.Args[3], 10, 64)
		t := vrtime.Time{TickCnt: tick, Priority: pri}
		fmt.Printf("%s", t.TimeStr())
	default:
		fmt.Println("unknown function")
		os.Exit(1)
	}
}
