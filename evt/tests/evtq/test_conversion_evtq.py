import unittest
import subprocess
import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../evtq')))
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../vrtime')))
import vrtime
import evtq

GO_CLI = os.path.abspath(os.path.join(os.path.dirname(__file__), 'go_evtq_compare'))

class TestEventQueueGoVsPython(unittest.TestCase):
    def run_go(self, *args):
        result = subprocess.run([GO_CLI] + [str(a) for a in args], capture_output=True, text=True)
        if result.returncode != 0:
            raise RuntimeError(f"Go CLI error: {result.stderr}")
        return result.stdout.strip()

    def test_len(self):
        go_out = self.run_go("Len")

        q = evtq.EventQueue.New()
        self.assertIn(f"empty:{q.Len()}", go_out)

        q = evtq.EventQueue.New()
        q.Insert("foo", vrtime.create_time(42, 7))
        self.assertIn(f"single:{q.Len()}", go_out)

        q = evtq.EventQueue.New()
        q.Insert("a", vrtime.create_time(15, 2))
        q.Insert("b", vrtime.create_time(3, 99))
        q.Insert("c", vrtime.create_time(27, 0))
        self.assertIn(f"multiple:{q.Len()}", go_out)

    def test_mintime(self):
        go_out = self.run_go("MinTime")

        q = evtq.EventQueue.New()
        q.Insert("foo", vrtime.create_time(42, 7))
        self.assertIn(f"single:{q.MinTime().TickCnt}", go_out)

        q = evtq.EventQueue.New()
        q.Insert("a", vrtime.create_time(15, 2))
        q.Insert("b", vrtime.create_time(3, 99))
        q.Insert("c", vrtime.create_time(27, 0))
        self.assertIn(f"multiple:{q.MinTime().TickCnt}", go_out)

    def test_pop(self):
        go_out = self.run_go("Pop")

        q = evtq.EventQueue.New()
        q.Insert("foo", vrtime.create_time(42, 7))
        val = q.Pop()
        self.assertIn(f"single:{val}", go_out)

        q = evtq.EventQueue.New()
        q.Insert("a", vrtime.create_time(15, 2))
        q.Insert("b", vrtime.create_time(3, 99))
        q.Insert("c", vrtime.create_time(27, 0))
        val = q.Pop()
        self.assertIn(f"multiple:{val}", go_out)

    def test_updatetime(self):
        go_out = self.run_go("UpdateTime")

        q = evtq.EventQueue.New()
        idA = q.Insert("a", vrtime.create_time(15, 2))
        idB = q.Insert("b", vrtime.create_time(3, 99))
        idC = q.Insert("c", vrtime.create_time(27, 0))
        q.UpdateTime(idB, vrtime.create_time(99, 0))
        item = q.GetItem(idB)
        self.assertIn(f"99 0", go_out)

        q.UpdateTime(9999, vrtime.create_time(42, 0))
        self.assertIn(f"nonexistent:<nil>", go_out)

    def test_remove(self):
        go_out = self.run_go("Remove")

        q = evtq.EventQueue.New()
        idA = q.Insert("a", vrtime.create_time(15, 2))
        idB = q.Insert("b", vrtime.create_time(3, 99))
        idC = q.Insert("c", vrtime.create_time(27, 0))
        ok = q.Remove(idB)
        self.assertIn(f"removed:{str(ok).lower()}", go_out)
        self.assertIn(f"length after remove:{q.Len()}", go_out)

        ok = q.Remove(9999)
        self.assertIn(f"nonexistent:{str(ok).lower()}", go_out)
        self.assertIn(f"length after nonexistent remove:{q.Len()}", go_out)

    def test_getitem(self):
        go_out = self.run_go("GetItem")

        q = evtq.EventQueue.New()
        idA = q.Insert("a", vrtime.create_time(15, 2))
        idB = q.Insert("b", vrtime.create_time(3, 99))
        idC = q.Insert("c", vrtime.create_time(27, 0))
        item = q.GetItem(idB)
        self.assertIn(f"3 99", go_out)

        item = q.GetItem(9999)
        self.assertIn(f"notfound:<nil>", go_out)

if __name__ == "__main__":
    unittest.main()
