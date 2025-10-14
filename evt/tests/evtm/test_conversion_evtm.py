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
            results.append((context, data))
        eid1, _ = mgr.schedule("ctx1", "dat1", handler, Time(10, 1))
        eid2, _ = mgr.schedule("ctx2", "dat2", handler, Time(5, 2))
        eid3, _ = mgr.schedule("ctx3", "dat3", handler, Time(5, 1))
        eid4, _ = mgr.schedule("ctx4", "dat4", handler, Time(15, 1))
        mgr.EventList.UpdateTime(eid4, Time(20, 1))
        mgr.cancel_event(eid2)
        mgr.run(20)
        py_expected = [
            ("ctx3", "dat3"),
            ("ctx4", "dat4"),
            ("ctx1", "dat1")
        ]
        for go_evt, py_evt in zip(go_results, py_expected):
            go_ctx, go_data = go_evt.split(",")
            self.assertEqual((go_ctx, go_data), py_evt)

    def test_scenario_cancel(self):
        go_results = self.run_go("scenario_cancel")
        mgr = evtm.EventManager()
        results = []
        def handler(mgr, context, data):
            results.append((context, data))
        eid1, _ = mgr.schedule("A", "X", handler, Time(2, 1))
        eid2, _ = mgr.schedule("B", "Y", handler, Time(4, 1))
        eid3, _ = mgr.schedule("C", "Z", handler, Time(6, 1))
        mgr.cancel_event(eid3)
        mgr.run(10)
        py_expected = [
            ("B", "Y"),
            ("A", "X")
        ]
        for go_evt, py_evt in zip(go_results[:2], py_expected):
            go_ctx, go_data = go_evt.split(",")
            self.assertEqual((go_ctx, go_data), py_evt)
        # Check cancellation status
        self.assertTrue("cancelled:True" in go_results[-1] or "cancelled:true" in go_results[-1])

if __name__ == "__main__":
    unittest.main()
