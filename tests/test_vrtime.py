import unittest
import numpy as np
import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../vrtime')))
import vrtime

class TestVrTime(unittest.TestCase):
    def setUp(self):
        # Reset globals to default before each test
        vrtime.set_ticks_per_second(np.int64(1.0 / np.float64(1e-10)))

    def test_set_ticks_per_second(self):
        vrtime.set_ticks_per_second(np.int64(1e7))
        self.assertEqual(vrtime.TicksPerSecond, np.int64(1e7))
        self.assertAlmostEqual(vrtime.FloatTicksPerSecond, np.float64(1e7))
        self.assertAlmostEqual(vrtime.SecondPerTick, np.float64(1.0/1e7))
        self.assertAlmostEqual(vrtime.TickValue, np.float64(1.0/1e7))
        self.assertEqual(vrtime.NanoSecPerTick, np.int64(np.float64(1e9) * (1.0/1e7)))

    def test_time_struct_and_methods(self):
        t = vrtime.create_time(np.int64(10), np.int64(2))
        self.assertEqual(t.Ticks(), np.int64(10))
        self.assertEqual(t.Pri(), np.int64(2))
        t.SetTicks(np.int64(20))
        t.SetPri(np.int64(5))
        self.assertEqual(t.Ticks(), np.int64(20))
        self.assertEqual(t.Pri(), np.int64(5))
        self.assertEqual(t.TimeStr(), f"(20,5)")
        self.assertTrue(isinstance(t.SecondsStr(), str))

    def test_seconds_to_time_and_ticks(self):
        t = vrtime.seconds_to_time(np.float64(1.0))
        ticks = vrtime.seconds_to_ticks(np.float64(1.0))
        self.assertEqual(t.Ticks(), ticks)
        t2 = vrtime.seconds_to_time_pri(np.float64(1.0), np.int64(7))
        self.assertEqual(t2.Pri(), np.int64(7))

    def test_mu_seconds_to_ticks(self):
        vrtime.set_ticks_per_second(np.int64(1e6))
        ticks = vrtime.mu_seconds_to_ticks(np.float64(1.0))
        self.assertEqual(ticks, np.int64(1))

    def test_time_to_seconds_and_ticks_to_seconds(self):
        t = vrtime.create_time(np.int64(100), np.int64(0))
        s = vrtime.time_to_seconds(t)
        s2 = vrtime.ticks_to_seconds(np.int64(100))
        self.assertAlmostEqual(s, s2)

    def test_cmp_time(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(2))
        t3 = vrtime.create_time(np.int64(6), np.int64(1))
        self.assertEqual(vrtime.cmp_time(t1, t2), -1)
        self.assertEqual(vrtime.cmp_time(t2, t1), 1)
        self.assertEqual(vrtime.cmp_time(t1, t1), 0)
        self.assertEqual(vrtime.cmp_time(t3, t1), 1)
        self.assertEqual(vrtime.cmp_time(t1, t3), -1)

    def test_time_comparisons(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(1))
        t2 = vrtime.create_time(np.int64(5), np.int64(2))
        t3 = vrtime.create_time(np.int64(6), np.int64(1))
        self.assertTrue(t1.LT(t2))
        self.assertTrue(t2.GT(t1))
        self.assertTrue(t1.EQ(t1))
        self.assertTrue(t1.LE(t2))
        self.assertTrue(t1.LE(t1))
        self.assertTrue(t2.GE(t1))
        self.assertTrue(t2.GE(t2))
        self.assertTrue(t1.NEQ(t2))

    def test_plus(self):
        t1 = vrtime.create_time(np.int64(5), np.int64(3))
        t2 = vrtime.create_time(np.int64(7), np.int64(2))
        t3 = t1.Plus(t2)
        self.assertEqual(t3.Ticks(), np.int64(12))
        self.assertEqual(t3.Pri(), np.int64(3))
        t4 = t2.Plus(t1)
        self.assertEqual(t4.Pri(), np.int64(3))

    def test_zero_and_infinity_time(self):
        z = vrtime.zero_time()
        inf = vrtime.infinity_time()
        self.assertEqual(z.Ticks(), np.int64(0))
        self.assertEqual(z.Pri(), np.int64(0))
        self.assertEqual(inf.Ticks(), np.iinfo(np.int64).max)
        self.assertEqual(inf.Pri(), np.iinfo(np.int64).max)

if __name__ == "__main__":
    unittest.main()
