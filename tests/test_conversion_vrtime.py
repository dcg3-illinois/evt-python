import unittest
import subprocess
import numpy as np
import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../vrtime')))
import vrtime

GO_CLI = os.path.abspath(os.path.join(os.path.dirname(__file__), 'go_vrtime_compare'))

class TestVrTimeGoPythonEquivalence(unittest.TestCase):
    def run_go(self, *args):
        result = subprocess.run([GO_CLI] + [str(a) for a in args], capture_output=True, text=True)
        if result.returncode != 0:
            print(f"Go CLI output: {result.stdout}")
            raise RuntimeError(f"Go CLI error: {result.stderr}")
        return result.stdout.strip()

    def test_TicksToSeconds(self):
        ticks = np.int64(123456789)
        py = vrtime.ticks_to_seconds(ticks)
        go = float(self.run_go("TicksToSeconds", ticks))
        self.assertAlmostEqual(py, go, places=9)

    def test_SecondsToTicks(self):
        sec = np.float64(1.23456789)
        py = vrtime.seconds_to_ticks(sec)
        go = int(self.run_go("SecondsToTicks", sec))
        self.assertEqual(py, go)

    def test_MuSecondsToTicks(self):
        mus = np.float64(123.456)
        py = vrtime.mu_seconds_to_ticks(mus)
        go = int(self.run_go("MuSecondsToTicks", mus))
        self.assertEqual(py, go)

    def test_SetTicksPerSecond(self):
        tps = np.int64(1e7)
        py = vrtime.set_ticks_per_second(tps)
        goTps = self.run_go("SetTicksPerSecond", tps)
        self.assertEqual(vrtime.TicksPerSecond, int(goTps))
        # Reset to default after test
        vrtime.set_ticks_per_second(np.int64(1.0 / np.float64(1e-10)))

    def test_TimeToSeconds(self):
        tick = np.int64(1000)
        pri = np.int64(5)
        t = vrtime.create_time(tick, pri)
        py = vrtime.time_to_seconds(t)
        go = float(self.run_go("TimeToSeconds", tick, pri))
        self.assertAlmostEqual(py, go, places=9)

    def test_CreateTime(self):
        tick = np.int64(100)
        pri = np.int64(7)
        py = vrtime.create_time(tick, pri)
        go = self.run_go("CreateTime", tick, pri)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_InfinityTime(self):
        py = vrtime.infinity_time()
        not_used = 0
        go = self.run_go("InfinityTime", not_used)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_ZeroTime(self):
        py = vrtime.zero_time()
        not_used = 0
        go = self.run_go("ZeroTime", not_used)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_SecondsToTime(self):
        sec = np.float64(2.345)
        py = vrtime.seconds_to_time(sec)
        go = self.run_go("SecondsToTime", sec)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_SecondsToTimePri(self):
        sec = np.float64(2.345)
        pri = np.int64(9)
        py = vrtime.seconds_to_time_pri(sec, pri)
        go = self.run_go("SecondsToTimePri", sec, pri)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_CmpTime(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(2))
        py = vrtime.cmp_time(t1, t2)
        go = int(self.run_go("CmpTime", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority))
        self.assertEqual(py, go)

    def test_EQ(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(1))
        py = t1.EQ(t2)
        go = self.run_go("EQ", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_GE(self):
        t1 = vrtime.create_time(np.int64(6), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(1))
        py = t1.GE(t2)
        go = self.run_go("GE", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_GT(self):
        t1 = vrtime.create_time(np.int64(6), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(1))
        py = t1.GT(t2)
        go = self.run_go("GT", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_LE(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(6), np.int64(1))
        py = t1.LE(t2)
        go = self.run_go("LE", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_LT(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(6), np.int64(1))
        py = t1.LT(t2)
        go = self.run_go("LT", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_NEQ(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(6), np.int64(1))
        py = t1.NEQ(t2)
        go = self.run_go("NEQ", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        self.assertEqual(str(py).lower(), go.lower())

    def test_Plus(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(3))
        t2 = vrtime.create_time(np.int64(7), np.int64(2))
        py = t1.Plus(t2)
        go = self.run_go("Plus", t1.TickCnt, t1.Priority, t2.TickCnt, t2.Priority)
        go_tick, go_pri = map(int, go.split(","))
        self.assertEqual(py.TickCnt, go_tick)
        self.assertEqual(py.Priority, go_pri)

    def test_Pri(self):
        t = vrtime.create_time(np.int64(5), np.int64(3))
        py = t.Pri()
        go = int(self.run_go("Pri", t.TickCnt, t.Priority))
        self.assertEqual(py, go)

    def test_Seconds(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        py = t.Seconds()
        go = float(self.run_go("Seconds", t.TickCnt, t.Priority))
        self.assertAlmostEqual(py, go, places=9)

    def test_SecondsStr(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        py = t.SecondsStr()
        go = self.run_go("SecondsStr", t.TickCnt, t.Priority)
        self.assertEqual(py, go)

    def test_SetPri(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        newpri = np.int64(7)
        t.SetPri(newpri)
        py = f"{t.TickCnt},{t.Priority}"
        go = self.run_go("SetPri", t.TickCnt, 3, newpri)
        self.assertEqual(py, go)

    def test_SetTicks(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        newtick = np.int64(654321)
        t.SetTicks(newtick)
        py = f"{t.TickCnt},{t.Priority}"
        go = self.run_go("SetTicks", 123456, 3, newtick)
        self.assertEqual(py, go)

    def test_Ticks(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        py = t.Ticks()
        go = int(self.run_go("Ticks", t.TickCnt, t.Priority))
        self.assertEqual(py, go)

    def test_TimeStr(self):
        t = vrtime.create_time(np.int64(123456), np.int64(3))
        py = t.TimeStr()
        go = self.run_go("TimeStr", t.TickCnt, t.Priority)
        self.assertEqual(py, go)

if __name__ == "__main__":
    unittest.main()
