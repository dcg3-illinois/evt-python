import unittest
# this makes vrtime visible to us
import evt.vrtime as vrtime
from evt.evtq import EventQueue

class TestEventQueuePython(unittest.TestCase):
    def setUp(self):
        self.q = EventQueue.New()

    def test_insert_and_len(self):
        t1 = vrtime.seconds_to_time(1.0)
        t2 = vrtime.seconds_to_time(2.0)
        self.q.Insert("event1", t1)
        self.q.Insert("event2", t2)
        self.assertEqual(self.q.Len(), 2)

    def test_min_time(self):
        t1 = vrtime.seconds_to_time(1.0)
        t2 = vrtime.seconds_to_time(2.0)
        self.q.Insert("event1", t2)
        self.q.Insert("event2", t1)
        min_time = self.q.MinTime()
        self.assertEqual(min_time.Ticks(), t1.Ticks())

    def test_pop(self):
        t1 = vrtime.seconds_to_time(1.0)
        t2 = vrtime.seconds_to_time(2.0)
        self.q.Insert("event1", t1)
        self.q.Insert("event2", t2)
        val = self.q.Pop()
        self.assertEqual(val, "event1")
        self.assertEqual(self.q.Len(), 1)

    def test_update_time(self):
        t1 = vrtime.seconds_to_time(1.0)
        t2 = vrtime.seconds_to_time(2.0)
        eid = self.q.Insert("event1", t1)
        self.q.UpdateTime(eid, t2)
        min_time = self.q.MinTime()
        self.assertEqual(min_time.Ticks(), t2.Ticks())

    def test_remove(self):
        t1 = vrtime.seconds_to_time(1.0)
        t2 = vrtime.seconds_to_time(2.0)
        eid = self.q.Insert("event1", t1)
        self.q.Insert("event2", t2)
        removed = self.q.Remove(eid)
        self.assertTrue(removed)
        self.assertEqual(self.q.Len(), 1)

    def test_get_item(self):
        t1 = vrtime.seconds_to_time(1.0)
        eid = self.q.Insert("event1", t1)
        item = self.q.GetItem(eid)
        self.assertIsNotNone(item)
        self.assertEqual(item.Value, "event1")

if __name__ == "__main__":
    unittest.main()
