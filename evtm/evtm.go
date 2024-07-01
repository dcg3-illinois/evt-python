// Package evtm manages the scheduling and execution of events.
// It depends on the package [evtq] for implementing the event list,
// and the package [vrtime] for implementing virtual time.
// mutexes are used to support concurrent access to an EventManager
// by multiple goroutines
package evtm

import (
	"fmt"
	"github.com/iti/evt/evtq"
	"github.com/iti/evt/vrtime"
	"log"
	"sync"
	"time"
)

// This logic is built around the EventManager data structure.
// It uses an evtq.EventQueue data structure to hold events, which itself
// uses the vrtime module's Time data structure to hold the values ordered by the EventQueue.
//
//	In addition to the usual functionality one expects from an event list,
//	- the EventManager has an option of running in 'wallclock' time, which ties
//	  the advancement of virtual time to be the same rate (in real-time) as physical time
//	  as might be seen on the clock on a wall.  This feature has particular utility when
//	  the simulation model integrates real devices that participate in the model execution,
//	  but are not controlled by the EventManager.
//	- access to critial data elements are guarded by mutex's, to enable use in a multi-threaded context
//	- all functions whose executions are triggered by the execution of an event have a particular
//	  function type declared.  With respect to what Go allows with function types, this one is simple and easy to use.
//	  The interface specifies the 'context' for the event, through provisioning of a pointer to some struct representing
//	  an entity in the model, and provisioning of some 'data' the event handler views as being presented to the event
//	  handler for application at that position in the model.
//	- there is user-selectable operating mode where if the event list goes empty the thread running the EventManager
//	  suspends;  the action of (some other thread) scheduling an event reactivates it.  This turns out to be needed
//	  when simulation is mixed with emulation, the emulation thread is running and needs to complete and possibly
//	  cause the scheduling of an event.
//
// To use the EventManager the simulation experiment code first builds the data structures of a model,
// and schedules one or more events whose execution starts the model execution.  In general, the execution of an event function
// will involve the scheduling of subsequent events (and their event handlers) to propagate forward the consequences of that
// event execution.
//
//	The process of executing an event is to (a) identify the event in the EventQueue with least Time value, (b) set the
//
// virtual time of the EventManager to be that value, and (c) call the event handler listed in the event,
// including as call parameters the context and data stored in the event.
//
// After an initial set of events are scheduled, one starts the simulation by calling the EventManager function method
// Run(LimitTime float64). Control is returned from this when the least time event in the EventQueue has a time that
// is strictly larger than LimitTime, or when there are no further events to process in the EventQueue.
// In the former case the clock of the EventManager is advanced to LimitTime even though no event necessarily took
// place at that time, in the latter case the EventManager time is left at the time of the least event executed.
// This design decision makes it possible for one to run the simulation, window of simulation time by window of simulation time,
// after a window completes loading the event list with events to execute in the next window without needing to dither around with
// the EventManager clock value.  Certain parallel simulation time management
// protocols work this way, and the design is meant to support that.
//
// evtMgrTrace is a flag used while debugging to selectively print/log information
var evtMgrTrace = false

// EventHandlerFunction is invoked when the corresponding event fires
type EventHandlerFunction func(*EventManager, any, any) any

// Event packages up the context and data sensitive
// information to be included when scheduling an event, and
// is used in dispatching the event handler
type Event struct {
	// Context is opaque information the event handler may need about where
	// and what it is executing.  The code that schedules the event
	// and the code that handles the event have to be using the same
	// type.
	Context any

	// Data is information the event handler uses to execute the event,
	// e.g., a message or frame.   Again, the code that schedules
	// the event and the code that executes it have to agree on the format,
	// because Go type checking won't do that.
	Data any

	// Time defines the instant (in the future) when the EventHandler
	// will be called.
	Time vrtime.Time

	// EventHandlerFunction is the function that is called to handle
	// the event.
	// These all have the signature  func(context any, data any) bool
	EventHandler EventHandlerFunction

	// EventID is a permanent identifier for this event. It can
	// be used to subsequently access or delete the event.
	// It is permanantly assigned when the event is scheduled.
	EventID int
}

// An EventManager structure holds information needed
// to schedule and execute events.  It has a pointer to an
// EventQueue (defined in package [evtq]), and a copy of the virtual
// time of the last event to have been removed from the EventQueue (which
// is interpreted as the virtual time of that event).  It has a Boolean
// flag also which, if changed to false when processing an event, will
// inhibit the dispatch of further events until the event manager
// is told to run again.
type EventManager struct {
	EventList *evtq.EventQueue // order events
	Time      vrtime.Time      // time of last event pulled off the EventList (but not necessarily yet executed completely)
	EventID   int              // identifier needed if we aim to remove events from EventList
	RunFlag   bool             // indicate whether the EventManager is actively in use right now
	Wallclock bool             // scale virtual time advance to wallclock time, approximately
	StartTime time.Time        // wallclock time at time of first event
	External  bool             // if true we don't close up when the event list is empy
	mu        sync.Mutex       // needed for thread safety
	suspended bool             // true when the thread running the EventManager is waiting for a signal sent when an event is scheduled
	suspChan  chan bool        //
	autoPri   int64            // use when time on event being scheduled has a priority of int64(0)
}

// New creates an empty event queue,
// sets the virtual time to zero, initializes
// the 'running' flag to false.
func New() *EventManager {
	if evtMgrTrace {
		fmt.Println("Creating new EvtMgr")
		log.Println("Creating new EvtMgr")
	}
	newEq := evtq.New()
	newEm := &EventManager{
		EventList: newEq,
		Time:      vrtime.ZeroTime(),
		RunFlag:   false,
		External:  false,
		suspended: false,
		suspChan:  make(chan bool, 1),
		autoPri:   int64(1),
		Wallclock: false}
	return newEm
}

// SetExternal sets the flag assigns a value to the flag which when true
// puts the EventManager into a mode where if the event list
// empties before reaching the end simulation time, the thread running the EventManager suspends
// until the scheduling (by a different thread) of an event on the EventManager releases it.
func (evtmgr *EventManager) SetExternal(external bool) {
	evtmgr.External = external
}

// SetWallclock assigns a value to the flag which when true puts the EventManager
// into a model where it runs in tandem with wallclock time
func (evtmgr *EventManager) SetWallclock(wallclock bool) {
	evtmgr.Wallclock = wallclock
}

// CurrentTime returns a copy of the simulation's current time.
func (evtmgr *EventManager) CurrentTime() vrtime.Time {
	evtmgr.mu.Lock()
	ct := evtmgr.Time
	evtmgr.mu.Unlock()
	return ct
}

// SetTime sets the Event Manager's clock to a specified vrtime
func (evtmgr *EventManager) SetTime(new_time vrtime.Time) {
	evtmgr.mu.Lock()
	evtmgr.Time = new_time
	evtmgr.mu.Unlock()
}

// CurrentSeconds gives the time using the seconds units
func (evtmgr *EventManager) CurrentSeconds() float64 {
	evtmgr.mu.Lock()
	cs := vrtime.TimeToSeconds(evtmgr.Time)
	evtmgr.mu.Unlock()
	return cs
}

// CurrentTicks returns the number of ticks since the EventManager started executing events, at tick 0
func (evtmgr *EventManager) CurrentTicks() int64 {
	evtmgr.mu.Lock()
	ct := evtmgr.Time.Ticks()
	evtmgr.mu.Unlock()
	return ct
}

// realTimeDelay computes how long the EventManager should sleep now if running wallclock time,
// and causes it to sleep that long
func (evtmgr *EventManager) realTimeDelay(current, tgt vrtime.Time) {
	if !evtmgr.Wallclock {
		return
	}

	// represent next event time in terms of ticks only
	evtmgr.mu.Lock()
	currentTimeInTicks := current.Ticks()
	tgtTimeInTicks := tgt.Ticks()

	// compute how long the thread running EventManager should sleep,
	// in the units of nanoseconds
	gapInTicks := tgtTimeInTicks - currentTimeInTicks
	gapInNanoseconds := gapInTicks * vrtime.NanoSecPerTick
	gapInDuration := time.Duration(gapInNanoseconds)
	evtmgr.mu.Unlock()

	// fmt.Printf("For a vt gap of %f seconds, suspend %f seconds\n",
	//	vrtime.TicksToSeconds(gapInTicks), gapInDuration.Seconds())

	time.Sleep(gapInDuration)
}

// function Run(LimitTime) starts the event dispatch loop for an EventManager
// that has been inactive.  It will stay in this processing loop
// until (a) there are events in queue, but none with timestamps greater than LimitTime,
// (b) there are no events in queue, or (c) the last event executed set the Event Manager's
// RunFlag to false.  In case (a) the clock of the Event Manager is set to LimitTime,
// in cases (b) and (c) the clock is left at the time of the last event executed.
func (evtmgr *EventManager) Run(LimitTime float64) {

	// input argument is in seconds, so transform to ticks
	LimitTimeInTicks := vrtime.SecondsToTicks(LimitTime)

	// as long as RunFlag is true the EventManager will stay in a loop
	// the next event is pulled from the EventQueue and dispatched
	evtmgr.RunFlag = true

	// remember the wallclock time when events started executing
	evtmgr.StartTime = time.Now()

	var entry bool = true
	// keep working if the RunFlag is true and there are events to dispatch
	for evtmgr.RunFlag == (entry || (evtmgr.EventList.Len() > 0 && evtmgr.CurrentTicks() < LimitTimeInTicks)) {

		entry = false
		// nxtEvt pulls off the package associated with the event with least
		// time-stamp and unpacks it into
		//   a) context is information the event handler may need about where and what
		//      it is executing.  The code that schedules the event and the code that
		//      handles the event have to be using the formatting, as the representation
		//      internal to go is interface{}.
		//   b) data is information the event handler uses to execute the event, e.g.,
		//      a message or frame.   Again, the code that schedules the event and the
		//      code that executes it have to agree on the format, because Go type checking
		//      won't do that
		//   c) handler is the function to call to handle the event.  These all have the
		//      signature  func(context any, data any) bool
		//      The boolean return flags whether the event was dispatched without error
		//   d) Events are given unique integer id numbers when scheduled, and evt_id
		//      returns that of the event being dispatched
		var nxtEvtTime vrtime.Time

		if evtmgr.EventList.Len() > 0 {
			// "wake up Clyde, we got something to do" (with apologies to JJ Cale)

			// virtual time when me and Cldye do it
			nxtEvtTime = evtmgr.EventList.MinTime()

			if evtMgrTrace {
				fmt.Printf("1. evt len %d, nxtTime %f\n", evtmgr.EventList.Len(), nxtEvtTime.Seconds())
				log.Printf("1. evt len %d, nxtTime %f\n", evtmgr.EventList.Len(), nxtEvtTime.Seconds())
			}

			// if the minimum next event falls beyond the termination time set the
			// event manager's time to the termination time and exit
			if LimitTimeInTicks < nxtEvtTime.Ticks() {
				evtmgr.Time = vrtime.CreateTime(LimitTimeInTicks, 0)
				break
			}

			// if so configured, hold back this thread to align with the wallclock
			if evtmgr.Wallclock {
				evtmgr.realTimeDelay(evtmgr.CurrentTime(), nxtEvtTime)
			}

			// get the next event, and call its handling function
			evtmgr.mu.Lock()
			event := nxtEvt(evtmgr.EventList) // safely extract the next event
			evtmgr.mu.Unlock()

			evtmgr.Time = event.Time       // update the EventManager's clock to be that of the next event
			evtmgr.EventID = event.EventID // remember the eventId while we can, before the event disappears

			// dispatch the event using the information carried along by the event
			event.EventHandler(evtmgr, event.Context, event.Data)

		}

		if evtmgr.External {
			if evtMgrTrace {
				fmt.Printf("Checking suspension %d, %t, lock %v\n", evtmgr.EventList.Len(), evtmgr.suspended, evtmgr.mu)
				log.Printf("Checking suspension %d, %t, lock %v\n", evtmgr.EventList.Len(), evtmgr.suspended, evtmgr.mu)
			}
			evtmgr.mu.Lock()
			if evtmgr.EventList.Len() == 0 {
				evtmgr.suspended = true
				if evtMgrTrace {
					fmt.Println("Suspending evtmgr")
					log.Println("Suspending evtmgr")
				}
				evtmgr.mu.Unlock()
				// block on release message
				_ = <-evtmgr.suspChan
				evtmgr.suspended = false
			} else {
				evtmgr.mu.Unlock()
			}
		}

		// in order to see if we're done yet we need to get the time of next event
		// so that it can be compared with a termination time
		evtmgr.mu.Lock()
		if evtmgr.EventList.Len() > 0 {
			nxtEvtTime = evtmgr.EventList.MinTime()
		}
		evtmgr.mu.Unlock()

	}
	// if we fell out of the loop because evtmgr.RunFlag was set to false by an event,
	// leave the clock of the event manager at the time of the last event executed.
	//   Likewise, if the loop ends because there are no further events, leave
	// the event manager time at the time of the last event executed.  This means
	// we do nothing here.
	if evtmgr.RunFlag && evtmgr.CurrentTime().Ticks() < LimitTimeInTicks {
		// Either the queue is exhausted, or the next item in the queue
		// starts beyond the termination time. In either event,
		// we soak up the remaining time.
		evtmgr.Time = vrtime.CreateTime(LimitTimeInTicks, 0)
	}

	// falling out of the displatch loop we know the EventManager isn't running anymore
	evtmgr.EventID = evtq.InvalidEventID
	evtmgr.RunFlag = false
}

// Stop stops the event dispatch loop of the EventManager.
func (evtmgr *EventManager) Stop() {
	evtmgr.RunFlag = false
}

var entryNum int = 1

// Schedule creates a new event and puts it on the EventManager's event queue.
// The call to Schedule passes all the parameters needed to create that event
//   - event handler function
//   - context where event is executed
//   - data used by the event execution
//   - how far into the virtual time future the event takes place.
//
// Schedule returns the eventId of the new event (a handle that can be used to cancel it)
// and the virtual time when the execution will occur.
func (evtmgr *EventManager) Schedule(context any, data any,
	handler func(*EventManager, any, any) any, offset vrtime.Time) (int, vrtime.Time) {

	// entryNum and eid are used in print statements during debugging
	eid := entryNum
	entryNum += 1
	if evtMgrTrace {
		fmt.Printf("enter Schedule entry %d with mutex %v, event time %f\n", eid, evtmgr.mu, evtmgr.Time.Plus(offset).Seconds())
		log.Printf("enter Schedule entry %d with mutex %v, event time %f\n", eid, evtmgr.mu, evtmgr.Time.Plus(offset).Seconds())
	}

	// change offset priority if it has a priority of 0
	if offset.Pri() == int64(0) {
		offset.SetPri(evtmgr.autoPri)
		evtmgr.autoPri += 1
	}

	evtmgr.mu.Lock()
	// time of the last event to be pulled from the EventQueue
	currentTime := evtmgr.Time

	// this event is offset time in the future
	newTime := currentTime.Plus(offset)

	// set the priority to be that of the event being scheduled
	newTime.SetPri(offset.Pri())

	// bundle together the information needed for event dispatch
	newEvent := Event{Context: context, EventHandler: handler, Data: data, Time: newTime}

	// put the event bundle into the EventQueue with priority equal to the
	// scheduled time, and get in return the unique event id

	eventID := evtmgr.EventList.Insert(&newEvent, newTime)

	// newEvent just got placed into the EventQueue but we can still get
	// at it and put in the identify of the event that carries it
	newEvent.EventID = eventID
	if evtMgrTrace {
		fmt.Printf("Schedule entry %d schedules event %d at %f\n", eid, eventID, newTime.Seconds())
		log.Printf("Schedule entry %d schedules event %d at %f\n", eid, eventID, newTime.Seconds())
	}
	evtmgr.mu.Unlock()
	if evtmgr.External {
		// we block the thread managing the EventManger if the event list becomes empty, or unblock
		// the thread when it is blocked and this scheduling transitions the event list from being
		// empty to non-empty
		evtmgr.mu.Lock()
		if evtmgr.suspended && evtmgr.EventList.Len() == 1 {
			// trigger for the unblock
			if evtMgrTrace {
				fmt.Println("Schedule unsuspends EventManager")
				log.Println("Schedule unsuspends EventManager")
			}

			// the thread is blocked on channel suspChan, so we unblock with sending a message down the channel
			evtmgr.suspChan <- true
		}
		evtmgr.mu.Unlock()
	}
	if evtMgrTrace {
		fmt.Printf("Schedule entry %d returns\n", eid)
		log.Printf("Schedule entry %d returns\n", eid)
	}

	// return the eventId gotten from EventQueue, and the time of the scheduled event
	return eventID, newTime
}

// nxtEvt pulls off the minimum time event from an EventQueue and
// debundles the information it contains, returning the
// unbundled fields
func nxtEvt(queue *evtq.EventQueue) *Event {
	event := queue.Pop().(*Event)
	return event
}

// RemoveEvent removes the indicated event from the event list,
// and returns a flag indicating whether the event was found and removed
func (evtmgr *EventManager) RemoveEvent(eventID int) bool {
	return evtmgr.EventList.Remove(eventID)
}
