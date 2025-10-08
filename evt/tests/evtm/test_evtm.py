import unittest
import threading
import evt.evtm as evtm
from evt.vrtime import Time, zero_time

class TestEventManager(unittest.TestCase):
    def setUp(self):
        self.mgr = evtm.EventManager()

    def test_set_external(self):
        self.mgr.set_external(True)
        self.assertTrue(self.mgr.External)
        self.mgr.set_external(False)
        self.assertFalse(self.mgr.External)

    def test_set_wallclock(self):
        self.mgr.set_wallclock(True)
        self.assertTrue(self.mgr.Wallclock)
        self.mgr.set_wallclock(False)
        self.assertFalse(self.mgr.Wallclock)

    def test_current_time_and_set_time(self):
        t = Time(123, 4)
        self.mgr.set_time(t)
        self.assertEqual(self.mgr.current_time().Ticks(), 123)
        self.assertEqual(self.mgr.current_time().Pri(), 4)
        t2 = Time(1234, 5)
        self.mgr.set_time(t2)
        self.assertEqual(self.mgr.current_time().Ticks(), 1234)
        self.assertEqual(self.mgr.current_time().Pri(), 5)

    def test_current_seconds_and_ticks(self):
        t = Time(1000, 2)
        self.mgr.set_time(t)
        self.assertEqual(self.mgr.current_ticks(), 1000)
        self.assertAlmostEqual(self.mgr.current_seconds(), t.Seconds())

    def test_schedule_and_run(self):
        results = []
        def handler(mgr, context, data):
            results.append((context, data))
        t_offset = Time(10, 1)
        eid, scheduled_time = self.mgr.schedule("ctx", "dat", handler, t_offset)
        self.assertIsInstance(eid, int)
        self.assertEqual(scheduled_time.Ticks(), 10)
        self.mgr.run(20)  # Should execute the event
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], ("ctx", "dat"))

    def test_cancel_event(self):
        results = []
        def handler(mgr, context, data):
            results.append((context, data))
        eid, _ = self.mgr.schedule("ctx", "dat", handler, Time(5, 1))
        cancelled = self.mgr.cancel_event(eid)
        self.assertTrue(cancelled)
        # After cancellation, event should not execute
        self.mgr.run(10)
        self.assertEqual(len(results), 0)

    def test_remove_event(self):
        def handler(mgr, context, data):
            pass
        eid, _ = self.mgr.schedule("ctx", "dat", handler, Time(5, 1))
        removed = self.mgr.remove_event(eid)
        self.assertTrue(removed)
        # After removal, event should not execute
        self.mgr.run(10)
        self.assertEqual(self.mgr.EventList.Len(), 0)

    def test_stop(self):
        self.mgr.RunFlag = True
        self.mgr.stop()
        self.assertFalse(self.mgr.RunFlag)

    def test_event_structure(self):
        e = Event("context", "data", Time(1, 2), lambda m, c, d: None, event_id=42, cancel=True)
        self.assertEqual(e.Context, "context")
        self.assertEqual(e.Data, "data")
        self.assertEqual(e.Time.Ticks(), 1)
        self.assertEqual(e.Time.Pri(), 2)
        self.assertEqual(e.EventID, 42)
        self.assertTrue(e.Cancel)

if __name__ == "__main__":
    unittest.main()
