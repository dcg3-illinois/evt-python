"""
Python translation of Go evtm.go
Manages scheduling and execution of events, using evtq for the event queue and vrtime for virtual time.
Thread-safe, supports wallclock mode, external suspension, and event handler dispatch.

This logic is built around the EventManager data structure.
It uses an EventQueue to hold events, which itself uses the vrtime module's Time data structure to hold the values ordered by the EventQueue.

In addition to the usual functionality one expects from an event list:
- the EventManager has an option of running in 'wallclock' time, which ties the advancement of virtual time to be the same rate (in real-time) as physical time as might be seen on the clock on a wall. This feature has particular utility when the simulation model integrates real devices that participate in the model execution, but are not controlled by the EventManager.
- access to critical data elements are guarded by mutexes, to enable use in a multi-threaded context
- all functions whose executions are triggered by the execution of an event have a particular function type declared. The interface specifies the 'context' for the event, through provisioning of a pointer to some struct representing an entity in the model, and provisioning of some 'data' the event handler views as being presented to the event handler for application at that position in the model.
- there is user-selectable operating mode where if the event list goes empty the thread running the EventManager suspends; the action of (some other thread) scheduling an event reactivates it. This turns out to be needed when simulation is mixed with emulation, the emulation thread is running and needs to complete and possibly cause the scheduling of an event.

To use the EventManager the simulation experiment code first builds the data structures of a model, and schedules one or more events whose execution starts the model execution. In general, the execution of an event function will involve the scheduling of subsequent events (and their event handlers) to propagate forward the consequences of that event execution.

The process of executing an event is to (a) identify the event in the EventQueue with least Time value, (b) set the virtual time of the EventManager to be that value, and (c) call the event handler listed in the event, including as call parameters the context and data stored in the event.

After an initial set of events are scheduled, one starts the simulation by calling the EventManager function method Run(LimitTime). Control is returned from this when the least time event in the EventQueue has a time that is strictly larger than LimitTime, or when there are no further events to process in the EventQueue. In the former case the clock of the EventManager is advanced to LimitTime even though no event necessarily took place at that time, in the latter case the EventManager time is left at the time of the least event executed. This design decision makes it possible for one to run the simulation, window of simulation time by window of simulation time, after a window completes loading the event list with events to execute in the next window without needing to dither around with the EventManager clock value. Certain parallel simulation time management protocols work this way, and the design is meant to support that.
"""
import threading
import time as pytime
# import vrtime
from evt.vrtime import Time, zero_time, seconds_to_ticks, ticks_to_seconds
import evt.evtq as evtq

# evtMgrTrace is a flag used while debugging to selectively print/log information
evtMgrTrace = False

class Event:
    """
    Event packages up the context and data sensitive information to be included when scheduling an event, and is used in dispatching the event handler.
    Context: opaque information the event handler may need about where and what it is executing.
    Data: information the event handler uses to execute the event, e.g., a message or frame.
    Time: defines the instant (in the future) when the EventHandler will be called.
    EventHandler: the function that is called to handle the event.
    EventID: permanent identifier for this event. It can be used to subsequently access or delete the event.
    Cancel: flag to mark event as cancelled.
    """
    def __init__(self, context, data, time, handler, event_id=None, cancel=False):
        self.Context = context
        self.Data = data
        self.Time = time
        self.EventHandler = handler  # Callable: (EventManager, context, data) -> any
        self.EventID = event_id
        self.Cancel = cancel

class EventManager:
    """
    An EventManager structure holds information needed to schedule and execute events.
    It has a pointer to an EventQueue, and a copy of the virtual time of the last event to have been removed from the EventQueue (which is interpreted as the virtual time of that event).
    It has a Boolean flag also which, if changed to false when processing an event, will inhibit the dispatch of further events until the event manager is told to run again.
    """
    def __init__(self):
        self.EventList = evtq.EventQueue()  # order events
        self.Time = zero_time()         # time of last event pulled off the EventList
        self.EventID = 0               # identifier needed if we aim to remove events from EventList
        self.NumEvts = 0               # number of events executed by the event manager
        self.RunFlag = False           # indicate whether the EventManager is actively in use right now
        self.Wallclock = False         # scale virtual time advance to wallclock time, approximately
        self.StartTime = None          # wallclock time at time of first event
        self.External = False          # if true we don't close up when the event list is empty
        self._lock = threading.Lock()  # needed for thread safety
        self.suspended = False         # true when the thread running the EventManager is waiting for a signal sent when an event is scheduled
        self.suspChan = threading.Event() # used for suspension signaling
        self.autoPri = 1               # use when time on event being scheduled has a priority of 0
        self.entryNum = 1

    def set_external(self, external):
        """Set the flag which, when true, puts the EventManager into a mode where if the event list empties before reaching the end simulation time, the thread running the EventManager suspends until the scheduling of an event releases it."""
        self.External = external

    def set_wallclock(self, wallclock):
        """Assigns a value to the flag which when true puts the EventManager into a model where it runs in tandem with wallclock time."""
        self.Wallclock = wallclock

    def current_time(self):
        """Returns a copy of the simulation's current time."""
        with self._lock:
            return self.Time.copy()

    def set_time(self, new_time):
        """Sets the Event Manager's clock to a specified vrtime."""
        with self._lock:
            self.Time = new_time.copy()

    def current_seconds(self):
        """Gives the time using the seconds units."""
        with self._lock:
            return ticks_to_seconds(self.Time.Ticks())

    def current_ticks(self):
        """Returns the number of ticks since the EventManager started executing events, at tick 0."""
        with self._lock:
            return self.Time.Ticks()

    def _real_time_delay(self, current, tgt):
        """Computes how long the EventManager should sleep now if running wallclock time, and causes it to sleep that long."""
        if not self.Wallclock:
            return
        gap_in_ticks = tgt.Ticks() - current.Ticks()
        gap_in_seconds = ticks_to_seconds(gap_in_ticks)
        if gap_in_seconds > 0:
            pytime.sleep(gap_in_seconds)

    def run(self, limit_time):
        """Starts the event dispatch loop for an EventManager that has been inactive. Will stay in this processing loop until (a) there are events in queue, but none with timestamps greater than limit_time, (b) there are no events in queue, or (c) the last event executed set the Event Manager's RunFlag to false. In case (a) the clock of the Event Manager is set to limit_time, in cases (b) and (c) the clock is left at the time of the last event executed."""
        # input argument is in seconds, so transform to ticks
        limit_ticks = seconds_to_ticks(limit_time)

        # as long as RunFlag is true the EventManager will stay in a loop
        # the next event is pulled from the EventQueue and dispatched
        self.RunFlag = True
        
        # remember the wallclock time when events started executing
        self.StartTime = pytime.time()
        
        entry = True
        # keep working if the RunFlag is true and there are events to dispatch
        while self.RunFlag and (entry or (self.EventList.Len() > 0 and self.current_ticks() < limit_ticks)):
            entry = False
            # nxt_evt pulls off the package associated with the event with least time-stamp and unpacks it
            #   a) context is information the event handler may need about where and what it is executing.
            #   b) data is information the event handler uses to execute the event, e.g., a message or frame.
            #   c) handler is the function to call to handle the event. These all have the signature  func(context, data) -> bool
            #   d) Events are given unique integer id numbers when scheduled, and evt_id returns that of the event being dispatched
            if self.EventList.Len() > 0:
                # "wake up Clyde, we got something to do" (with apologies to JJ Cale)
                
                # virtual time when me and Clyde do it
                nxt_evt_time = self.EventList.MinTime()
                if evtMgrTrace:
                    print(f"1. evt len {self.EventList.Len()}, nxtTime {ticks_to_seconds(nxt_evt_time.Ticks())}")
                if limit_ticks < nxt_evt_time.Ticks():
                    self.Time = Time(limit_ticks, 0)
                    break

                # if so configured, hold back this thread to align with the wallclock
                if self.Wallclock:
                    self._real_time_delay(self.current_time(), nxt_evt_time)
                
                # get the next event, and call its handling function
                with self._lock:
                    event = self._nxt_evt()
                
                self.Time = event.Time.copy()       # update the EventManager's clock to be that of the next event
                self.EventID = event.EventID        # remember the eventId while we can, before the event disappears
                
                # dispatch the event using the information carried along by the event
                if not event.Cancel:
                    event.EventHandler(self, event.Context, event.Data)
                    self.NumEvts += 1
                    
            # if configured for external suspension, check if we need to suspend
            if self.External:
                if evtMgrTrace:
                    print(f"Checking suspension {self.EventList.Len()}, {self.suspended}, lock {self._lock}")
                len_flag = False

                # check the length and set suspended with the lock held
                with self._lock:
                    if self.EventList.Len() == 0:
                        len_flag = True
                        self.suspended = True
                        if evtMgrTrace:
                            print("Suspending evtmgr")
                # release lock, but still wait if list was empty
                if len_flag:
                    self.suspChan.wait()  # block on release message
                    self.suspended = False
            
            # in order to see if we're done yet we need to get the time of next event so that it can be compared with a termination time
            with self._lock:
                if self.EventList.Len() > 0:
                    nxt_evt_time = self.EventList.MinTime()
        
        
        # if we fell out of the loop because RunFlag was set to false by an event,
        # leave the clock of the event manager at the time of the last event executed.
        # Likewise, if the loop ends because there are no further events, leave the event manager time at the time of the last event executed.
        # This means we do nothing here.
        if self.RunFlag and self.current_time().Ticks() < limit_ticks:
            # Either the queue is exhausted, or the next item in the queue starts beyond the termination time. In either event, we soak up the remaining time.
            self.Time = Time(limit_ticks, 0)
       
        # falling out of the dispatch loop we know the EventManager isn't running anymore
        self.EventID = 0
        self.RunFlag = False

    def stop(self):
        """Stops the event dispatch loop of the EventManager."""
        self.RunFlag = False

    def schedule(self, context, data, handler, offset):
        """Creates a new event and puts it on the EventManager's event queue. Returns the eventId of the new event and the virtual time when the execution will occur."""
        eid = self.entryNum
        self.entryNum += 1

        if evtMgrTrace:
            print(f"enter Schedule entry {eid} with mutex {self._lock}, event time {ticks_to_seconds(self.Time.Plus(offset).Ticks())}")
        
        # change offset priority if it has a priority of 0
        if offset.Pri() == 0:
            offset.SetPri(self.autoPri)
            self.autoPri += 1
        
        with self._lock:
            # time of the last event to be pulled from the EventQueue
            current_time = self.Time.copy()
            
            # this event is offset time in the future
            new_time = current_time.Plus(offset)
            
            # set the priority to be that of the event being scheduled
            new_time.SetPri(offset.Pri())
            
            # bundle together the information needed for event dispatch
            new_event = Event(context, data, new_time, handler)
            
            # put the event bundle into the EventQueue with priority equal to the scheduled time, and get in return the unique event id
            event_id = self.EventList.Insert(new_event, new_time)
            
            # new_event just got placed into the EventQueue but we can still get at it and put in the identify of the event that carries it
            new_event.EventID = event_id
            
            if evtMgrTrace:
                print(f"Schedule entry {eid} schedules event {event_id} at {ticks_to_seconds(new_time.Ticks())}")
            
            # we block the thread managing the EventManager if the event list becomes empty, or unblock the thread when it is blocked and this scheduling transitions the event list from being empty to non-empty
            if self.External and self.suspended and self.EventList.Len() == 1:
                if evtMgrTrace:
                    print("Schedule unsuspends EventManager")
                self.suspChan.set()
        
        if evtMgrTrace:
            print(f"Schedule entry {eid} returns")
        
        # return the eventId gotten from EventQueue, and the time of the scheduled event
        return event_id, new_time.copy()

    def cancel_event(self, event_id):
        """Cancels the indicated event from the event list."""
        item = self.EventList.GetItem(event_id)
        if item is not None:
            item.Value.Cancel = True
        return item is not None

    def remove_event(self, event_id):
        """Removes the indicated event from the event list, and returns a flag indicating whether the event was found and removed."""
        return self.EventList.Remove(event_id)

    def _nxt_evt(self):
        """Pulls off the minimum time event from an EventQueue and returns it."""
        return self.EventList.Pop()
