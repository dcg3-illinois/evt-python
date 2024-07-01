// Package vrtime defines and manages virtual time inside a simulator.
// Time is tracked as an integral number of ticks since
// the epoch, along with a secondary sort value to provide for
// deterministic order among simultaneous events.
package vrtime

import (
	"fmt"
	"math"
)

// Time is represented by a pair of int64s.  The primary one is
// a tick counter, reporting the number of discrete 'ticks' since
// the start of the simulation.  The amount of continuous time represented
// by a tick is configurable, with 0.1 microsecond being the default.
// The second integer is a priority code which plays a role in ordering
// two events with the same tick count.   As with the first int, the priority
// value with smaller value has the greater priority.  No particular requirement
// or meaning is by default assigned to the priority, but a user is able to
// set it and access it and so use it in any way they like, so long as the
// ordering constraint is understood.

// TicksPerSecond specifies the frequency of the ticker.  Default is 1e7.
var TicksPerSecond int64 = 1_000_000

// FloatTicksPerSecond gives a float64 representation of the size of number of ticks a second has
var FloatTicksPerSecond float64 = float64(TicksPerSecond)

// SecondPerTick gives a float64 representation of the tick size in seconds
var SecondPerTick float64 = 1.0 / FloatTicksPerSecond

// NanoSecPerTick gives a float64 representation of the tick size in nanoseconds
var NanoSecPerTick int64 = int64((float64(1e9)) * SecondPerTick)

// TickValue gives the size of a tick, in seconds
var TickValue float64 = 1.0 / FloatTicksPerSecond

// Time measures the number of ticks since the epoch (in
// TickCnt). The tick count provides a natural ordering
// of time values -- smaller numbers happen earlier than larger
// numbers. In order to provide determinism, the order in
// which simultaneous events occur is specified by Priority --
// among simultaneous events, smaller numbers for Priority
// occur before larger numbers.
type Time struct {
	TickCnt  int64
	Priority int64
}

// SetTicksPerSecond changes the value of [TicksPerSecond]
// (the frequency of the ticker) and the associated values
// [FloatTicksPersecond] and [TickValue].
// the frequency of the ticker. The default value is 1e7.
func SetTicksPerSecond(tps int64) bool {
	TicksPerSecond = tps
	FloatTicksPerSecond = float64(TicksPerSecond)
	SecondPerTick = 1.0 / FloatTicksPerSecond
	NanoSecPerTick = int64((float64(1e9)) * SecondPerTick)
	TickValue = 1.0 / FloatTicksPerSecond
	return true
}

// Ticks returns the primary key of a Time data structure, usually
// used to describe a length of time (e.g. between events)
func (t Time) Ticks() int64 {
	return t.TickCnt
}

// Ticks returns the float64 respresentation of the primary key of a Time data structure, usually
// used to describe a length of time (e.g. between events)
func (t Time) Seconds() float64 {
	return TicksToSeconds(t.TickCnt)
}

// Pri returns the priority of an event, which plays a role in
// ordering when two events have the same primary key value.
// Events with lower priority number are ordered to appear before events with higher priority numbers
func (t *Time) Pri() int64 {
	return t.Priority
}

// SetTicks sets the time associated with an event.
// N.b.: this does not modify the Priority.
func (t *Time) SetTicks(v int64) {
	t.TickCnt = v
}

// SetPri sets the priority of an event. N.b. this does
// not modify the Ticks
func (t *Time) SetPri(p int64) {
	t.Priority = p
}

// TimeStr provides a human-readable version of Time, including
// both Ticks and Priority.
func (t *Time) TimeStr() string {
	return fmt.Sprintf("(%v,%v)", t.TickCnt, t.Priority)
}

// SecondsStr returns a human-readble representation
// of a [Time]. The time is represented as fractional
// seconds rather than ticks.
func (t *Time) SecondsStr() string {
	return fmt.Sprintf("(%e,%v)", TicksToSeconds(t.TickCnt), t.Priority)
}

// CreateTime creates a Time object.
func CreateTime(ticks, priority int64) Time {
	return Time{TickCnt: ticks, Priority: priority}
}

// SecondsToTime converts a fractional number of seconds into
// an equivalent [Time] value. The returned Priority is 0.
func SecondsToTime(v float64) Time {
	ticks := SecondsToTicks(v)
	return Time{TickCnt: ticks, Priority: int64(0)}
}

// SecondsToTicks converts a fractional number of seconds
// into a whole number of ticks.
func SecondsToTicks(v float64) int64 {
	return int64(math.Round(v * FloatTicksPerSecond))
}

// MuSecondsToTicks converts a fractional number of micro-seconds
// into a whole number of ticks.
func MuSecondsToTicks(v float64) int64 {
	return int64(math.Round(v * FloatTicksPerSecond / 1e6))
}

// TimeToSeconds converts a [Time] value into a
// fractional number of seconds. The Priority field
// is ignored.
func TimeToSeconds(t Time) float64 {
	return float64(t.TickCnt) / FloatTicksPerSecond
}

// TicksToSeconds converts a whole number of ticks
// into a fractional number of seconds.
func TicksToSeconds(ticks int64) float64 {
	return float64(ticks) / FloatTicksPerSecond
}

// cmpTime is a utility function underlying all of the comparison operators.
// Uses standard unix-ish return of -1, 0, 1 to report 'lhs has higher priority', 'lhs and rhs are equal', 'lhs has smaller priority'
// Note that both Ticks and Priority particpate in the comparison
func cmpTime(lhs Time, rhs Time) int {
	lhsTicks := lhs.Ticks()
	rhsTicks := rhs.Ticks()

	// compare based on the primary key, ticks
	if lhsTicks < rhsTicks {
		return -1
	} else if lhsTicks > rhsTicks {
		return 1
	}

	// ticks are equal, compare based on priority, smaller appears before larger
	lhsPri := lhs.Pri()
	rhsPri := rhs.Pri()
	if lhsPri < rhsPri {
		return -1
	} else if lhsPri > rhsPri {
		return 1
	}
	return 0
}

// ZeroTime returns a [Time] structure with value zero in both keys
func ZeroTime() Time {
	return Time{TickCnt: 0, Priority: 0}
}

// InfinityTime marks the end of time. Every other time
// in a running simulation is less than InfinityTime
func InfinityTime() Time {
	return Time{TickCnt: math.MaxInt64, Priority: math.MaxInt64}
}

// LT returns true iff the receiver Time is less than the argument Time
func (t Time) LT(t1 Time) bool {
	cmp := cmpTime(t, t1)
	return cmp == -1
}

// GT returns true iff the receiver Time is greater than the argument Time
func (t Time) GT(t1 Time) bool {
	cmp := cmpTime(t, t1)
	return cmp == 1
}

// EQ returns true iff the receiver Time is equal to the argument Time.
// Both ticks and priority have to be the same for true to be returned
func (t Time) EQ(t1 Time) bool {
	cmp := cmpTime(t, t1)
	return cmp == 0
}

// LE returns true iff the receiver Time is less than or equal
// to the argument Time.
func (t Time) LE(t1 Time) bool {
	return t.LT(t1) || t.EQ(t1)
}

// GE returns true iff the receiver Time is greater than or equal
// to the argument Time.
func (t Time) GE(t1 Time) bool {
	return t.GT(t1) || t.EQ(t1)
}

// NEQ returns true iff the receiver time is not equal to the
// argument Time. Like all of the Time comparison functions,
// both the Ticks and the Priority are used.
func (t Time) NEQ(t1 Time) bool {
	return !t.EQ(t1)
}

// Plus adds the receiver Time to the argument Time.
// When adding time, add the ticks and set the priority to be the
// dominant one.

func (t Time) Plus(a Time) Time {
	ticks := t.Ticks() + a.Ticks()
	tPri := t.Pri()
	aPri := a.Pri()

	if tPri > aPri {
		return Time{ticks, tPri}
	}
	return Time{ticks, aPri}
}
