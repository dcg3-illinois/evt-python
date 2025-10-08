
# vrtime.py: Python translation of Go vrtime.go
# Data types and logic are preserved as closely as possible
#
# This module defines and manages virtual time inside a simulator.
# Time is tracked as an integral number of ticks since the epoch, along with a secondary sort value to provide for deterministic order among simultaneous events.

import numpy as np
from dataclasses import dataclass


# SecondPerTick gives a float64 representation of the tick size in seconds. Default 0.1 ns
SecondPerTick = np.float64(1e-10)

# TicksPerSecond specifies the frequency of the ticker. Default is 1e9.
TicksPerSecond = np.int64(1.0 / np.float64(1e-10))

# FloatTicksPerSecond gives a float64 representation of the number of ticks a second has
FloatTicksPerSecond = np.float64(np.int64(1.0 / np.float64(1e-10)))

# NanoSecPerTick gives a float64 representation of the tick size in nanoseconds
NanoSecPerTick = np.int64(np.float64(1e9) * np.float64(1e-10))

# TickValue gives the size of a tick, in seconds
TickValue = np.float64(1e-10)


def set_ticks_per_second(tps: np.int64) -> bool:
    """
    Changes the value of TicksPerSecond (the frequency of the ticker) and the associated values
    FloatTicksPerSecond and TickValue. The default value is 1e7.
    """
    global TicksPerSecond, FloatTicksPerSecond, SecondPerTick, NanoSecPerTick, TickValue
    TicksPerSecond = np.int64(tps)
    FloatTicksPerSecond = np.float64(tps)
    SecondPerTick = np.float64(1.0) / FloatTicksPerSecond
    NanoSecPerTick = np.int64(np.float64(1e9) * SecondPerTick)
    TickValue = np.float64(1.0) / FloatTicksPerSecond
    return True


# Time measures the number of ticks since the epoch (in TickCnt). The tick count provides a natural ordering
# of time values -- smaller numbers happen earlier than larger numbers. In order to provide determinism, the order in
# which simultaneous events occur is specified by Priority -- among simultaneous events, smaller numbers for Priority
# occur before larger numbers.
@dataclass
class Time:
    TickCnt: np.int64
    Priority: np.int64


    def Ticks(self) -> np.int64:
        """Returns the primary key of a Time data structure, usually used to describe a length of time (e.g. between events)"""
        return self.TickCnt


    def Seconds(self) -> np.float64:
        """Returns the float64 representation of the primary key of a Time data structure, usually used to describe a length of time (e.g. between events)"""
        return ticks_to_seconds(self.TickCnt)


    def Pri(self) -> int:
        """Returns the priority of an event, which plays a role in ordering when two events have the same primary key value. Events with lower priority number are ordered to appear before events with higher priority numbers."""
        return self.Priority


    def SetTicks(self, v: np.int64):
        """Sets the time associated with an event. N.b.: this does not modify the Priority."""
        self.TickCnt = np.int64(v)


    def SetPri(self, p: np.int64):
        """Sets the priority of an event. N.b. this does not modify the Ticks."""
        self.Priority = np.int64(p)


    def TimeStr(self) -> str:
        """Provides a human-readable version of Time, including both Ticks and Priority."""
        return f"({self.TickCnt},{self.Priority})"


    def SecondsStr(self) -> str:
        """Returns a human-readable representation of a Time. The time is represented as fractional seconds rather than ticks."""
        return f"({ticks_to_seconds(self.TickCnt):e},{self.Priority})"


    def LT(self, t1: 'Time') -> bool:
        """Returns True iff the receiver Time is less than the argument Time."""
        return cmp_time(self, t1) == -1


    def GT(self, t1: 'Time') -> bool:
        """Returns True iff the receiver Time is greater than the argument Time."""
        return cmp_time(self, t1) == 1


    def EQ(self, t1: 'Time') -> bool:
        """Returns True iff the receiver Time is equal to the argument Time. Both ticks and priority have to be the same for True to be returned."""
        return cmp_time(self, t1) == 0


    def LE(self, t1: 'Time') -> bool:
        """Returns True iff the receiver Time is less than or equal to the argument Time."""
        return self.LT(t1) or self.EQ(t1)


    def GE(self, t1: 'Time') -> bool:
        """Returns True iff the receiver Time is greater than or equal to the argument Time."""
        return self.GT(t1) or self.EQ(t1)


    def NEQ(self, t1: 'Time') -> bool:
        """Returns True iff the receiver time is not equal to the argument Time. Like all of the Time comparison functions, both the Ticks and the Priority are used."""
        return not self.EQ(t1)


    def Plus(self, a: 'Time') -> 'Time':
        """Adds the receiver Time to the argument Time. When adding time, add the ticks and set the priority to be the dominant one."""
        ticks = self.Ticks() + a.Ticks()
        tPri = self.Pri()
        aPri = a.Pri()
        if tPri > aPri:
            return Time(ticks, tPri)
        return Time(ticks, aPri)


# Utility functions

# CreateTime creates a Time object.
def create_time(ticks: np.int64, priority: np.int64) -> Time:
    """Creates a Time object."""
    return Time(np.int64(ticks), np.int64(priority))


# SecondsToTime converts a fractional number of seconds into an equivalent Time value. The returned Priority is 0.
def seconds_to_time(v: np.float64) -> Time:
    """Converts a fractional number of seconds into an equivalent Time value. The returned Priority is 0."""
    ticks = seconds_to_ticks(v)
    return Time(ticks, np.int64(0))


# SecondsToTimePri converts a fractional number of seconds into an equivalent Time value. The returned Priority is pri.
def seconds_to_time_pri(v: np.float64, pri: np.int64) -> Time:
    """Converts a fractional number of seconds into an equivalent Time value. The returned Priority is pri."""
    ticks = seconds_to_ticks(v)
    return Time(ticks, np.int64(pri))


# SecondsToTicks converts a fractional number of seconds into a whole number of ticks.
def seconds_to_ticks(v: np.float64) -> np.int64:
    """Converts a fractional number of seconds into a whole number of ticks."""
    return np.int64(round(v * FloatTicksPerSecond))


# MuSecondsToTicks converts a fractional number of micro-seconds into a whole number of ticks.
def mu_seconds_to_ticks(v: np.float64) -> np.int64:
    """Converts a fractional number of micro-seconds into a whole number of ticks."""
    return np.int64(round(v * FloatTicksPerSecond / np.float64(1e6)))


# TimeToSeconds converts a Time value into a fractional number of seconds. The Priority field is ignored.
def time_to_seconds(t: Time) -> np.float64:
    """Converts a Time value into a fractional number of seconds. The Priority field is ignored."""
    return np.float64(t.TickCnt) / FloatTicksPerSecond


# TicksToSeconds converts a whole number of ticks into a fractional number of seconds.
def ticks_to_seconds(ticks: np.int64) -> np.float64:
    """Converts a whole number of ticks into a fractional number of seconds."""
    return np.float64(ticks) / FloatTicksPerSecond


# cmpTime is a utility function underlying all of the comparison operators.
# Uses standard unix-ish return of -1, 0, 1 to report 'lhs has higher priority', 'lhs and rhs are equal', 'lhs has smaller priority'.
# Note that both Ticks and Priority participate in the comparison.
def cmp_time(lhs: Time, rhs: Time) -> int:
    """Compares two Time values. Returns -1, 0, or 1 as in standard unix comparison."""
    lhsTicks = lhs.Ticks()
    rhsTicks = rhs.Ticks()
    if lhsTicks < rhsTicks:
        return -1
    elif lhsTicks > rhsTicks:
        return 1
    lhsPri = lhs.Pri()
    rhsPri = rhs.Pri()
    if lhsPri < rhsPri:
        return -1
    elif lhsPri > rhsPri:
        return 1
    return 0


# ZeroTime returns a Time structure with value zero in both keys.
def zero_time() -> Time:
    """Returns a Time structure with value zero in both keys."""
    return Time(np.int64(0), np.int64(0))


# InfinityTime marks the end of time. Every other time in a running simulation is less than InfinityTime.
def infinity_time() -> Time:
    """Marks the end of time. Every other time in a running simulation is less than InfinityTime."""
    maxint = np.iinfo(np.int64).max
    return Time(np.int64(maxint), np.int64(maxint))
