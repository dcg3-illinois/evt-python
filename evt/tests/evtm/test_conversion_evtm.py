import unittest
import subprocess
import os
import evt.evtm as evtm
from evt.vrtime import Time

GO_CLI = os.path.abspath(os.path.join(os.path.dirname(__file__), 'go_evtm_compare'))

class TestEventManagerGoVsPython(unittest.TestCase):
    def run_go(self, scenario):
        result = subprocess.run([GO_CLI, scenario], capture_output=True, text=True)
        if result.returncode != 0:
            raise RuntimeError(f"Go CLI error: {result.stderr}")
        return [line.strip() for line in result.stdout.strip().splitlines() if line.strip()]

    def test_scenario_basic(self):
        go_results = self.run_go("scenario_basic")
        mgr = evtm.EventManager()
        results = []
        def handler(mgr, context, data):
            context.append(data)
        eid1, _ = mgr.schedule(results, 1, handler, Time(10, 1))
        eid2, _ = mgr.schedule(results, 2, handler, Time(5, 2))
        eid3, _ = mgr.schedule(results, 3, handler, Time(5, 1))
        eid4, _ = mgr.schedule(results, 4, handler, Time(15, 1))
        mgr.EventList.UpdateTime(eid4, Time(7, 1))
        mgr.cancel_event(eid2)
        mgr.run(20)
        py_expected = [
            3,
            4,
            1
        ]
        # go output has the form ['[val1 val2 val3]']
        go_arr = go_results[0].strip("[]").split()
        # ensure the two outputs match in order
        for go_evt, py_evt in zip(go_arr, py_expected):
            self.assertEqual(int(go_evt), py_evt) 

    def test_scenario_cancel(self):
        go_results = self.run_go("scenario_cancel")
        mgr = evtm.EventManager()
        results = []
        def handler(mgr, context, data):
            context.append(data)
        eid1, _ = mgr.schedule(results, 1, handler, Time(2, 1))
        eid2, _ = mgr.schedule(results, 2, handler, Time(4, 1))
        eid3, _ = mgr.schedule(results, 3, handler, Time(6, 1))
        mgr.cancel_event(eid3)
        mgr.run(10)
        py_expected = [
            1,
            2
        ]

        # go output has the form ['[val1 val2 val3]']
        go_arr = go_results[0].strip("[]").split()
        # ensure the two outputs match in order
        for go_evt, py_evt in zip(go_arr, py_expected):
            self.assertEqual(int(go_evt), py_evt) 

if __name__ == "__main__":
    unittest.main()
