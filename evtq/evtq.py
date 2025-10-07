
# evtq.py: Python translation of Go evtq.go
# Event queue implementation using heapq and threading for thread safety
#
# Package evtq creates and manages event queues
# It depends upon Python's heapq to manage a heap structure

import heapq
import threading
from typing import Any, Dict, Optional
import vrtime


# InvalidEventID will never be returned from EventQueue.Insert()
InvalidEventID = 0

class Item:
    """
    The item struct defines the item organized by Time value
    itemID: unique identifier with every event inserted into the queue
    Value: completely general payload for the item
    Time: the field used to order the elements
    index: the position of the item in the (heap-organized) list of events
    Cancel: has been marked for removal
    """
    def __init__(self, itemID: int, value: Any, time: 'vrtime.Time'):
        self.itemID = itemID
        self.Value = value
        self.Time = time
        self.index = 0
        self.Cancel = False

    def __lt__(self, other: 'Item'):
        return self.Time.LT(other.Time)

class ItemHeap:
    """
    ItemHeap is the type of the data structure written to satisfy the Heap interface
    Implements Len, Less, Swap, Push, and Pop required for a type to satisfy the Heap interface
    """
    def __init__(self):
        self.data = []

    def __len__(self):
        return len(self.data)

    def push(self, item: Item):
        heapq.heappush(self.data, item)

    def pop(self) -> Item:
        return heapq.heappop(self.data)

    def fix(self, index: int):
        heapq.heapify(self.data)

    def get(self, index: int) -> Item:
        return self.data[index]

    def swap(self, i: int, j: int):
        self.data[i], self.data[j] = self.data[j], self.data[i]
        self.data[i].index = i
        self.data[j].index = j

class EventQueue:
    """
    EventQueue represents the queue
    evtID: monotonically increasing counter used for default secondary time in event Time
    itemHeap: data structure holding items
    lookup: event identifier to event, used for marking events to be ignored
    MaxTime: Largest vrtime.Time value pushed onto to the heap as yet
    mu: used to support thread safety
    """
    def __init__(self):
        self.evtID = InvalidEventID
        self.itemHeap = ItemHeap()
        self.lookup: Dict[int, Item] = {}
        self.MaxTime = vrtime.zero_time()
        self.mu = threading.Lock()

    @staticmethod
    def New():
        """New is a constructor. Initializes an empty list of events"""
        return EventQueue()

    def Len(self) -> int:
        """Len returns the number of elements in the queue."""
        with self.mu:
            return len(self.itemHeap)

    def MinTime(self) -> 'vrtime.Time':
        """MinTime returns the Time associated with the next event."""
        with self.mu:
            if len(self.itemHeap) == 0:
                return vrtime.zero_time()
            return self.itemHeap.get(0).Time

    def Insert(self, v: Any, time: 'vrtime.Time') -> int:
        """Insert inserts a new element into the queue. No action is performed on duplicate elements."""
        with self.mu:
            self.evtID += 1

            # update maximum time of inserted event
            if self.MaxTime.LT(time):
                self.MaxTime = time

            # if the priority in the time stamp is -1
            # change it to be something monotonically increasing
            # to try to give some determinism for events with the
            # same tick count when some other priority isn't given
            if time.Pri() == -1:
                time.SetPri(int(self.evtID))

            # create an item for insertion
            newItem = Item(self.evtID, v, time)

            self.itemHeap.push(newItem)
            self.lookup[self.evtID] = newItem
            return self.evtID

    def Pop(self) -> Optional[Any]:
        """Pop removes the element with the least time from the queue and returns it. In case of an empty queue, returns None."""
        with self.mu:
            if len(self.itemHeap) == 0:
                return None
            popped = self.itemHeap.pop()
            self.lookup.pop(popped.itemID, None)
            return popped.Value

    def UpdateTime(self, evtID: int, newTime: 'vrtime.Time'):
        """UpdateTime changes the priority of a given item. If the specified item is not present in the queue, no action is performed."""
        with self.mu:
            item = self.lookup.get(evtID)

            if item is None:
                return
            
            item.Time = newTime
            self.itemHeap.fix(item.index)

    def GetItem(self, evtID: int) -> Optional[Item]:
        """GetItem returns the item for the given event ID, or None if not present."""
        with self.mu:
            return self.lookup.get(evtID)

    def Remove(self, evtID: int) -> bool:
        """Remove an element. Returns True on success."""
        with self.mu:
            element = self.lookup.get(evtID)
            if element is None:
                return False
            
            # we're going to push the element to be rid of to the top
            element.Time = vrtime.zero_time()
            self.itemHeap.fix(element.index)

            # pop it off
            popped = self.itemHeap.pop()
            self.lookup.pop(popped.itemID, None)
            
            return True
