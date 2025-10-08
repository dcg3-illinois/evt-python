# evt
Discrete event simulator converted to python

# evt

[![Report Card](https://goreportcard.com/badge/github.com/iti/evt)](https://goreportcard.com/report/github.com/iti/evt) 
[![Go Reference](https://pkg.go.dev/badge/github.com/iti/evt.svg)](https://pkg.go.dev/github.com/iti/evt) 
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Coverage](https://img.shields.io/badge/Coverage-0.0%25-red)

## evt/evtm

Package [evtm] manages the scheduling and execution of events.
It depends on the package [evtq] for implementing the event list,
and the package [vrtime] for implementing virtual time.
Mutexes are used to support concurrent access to an EventManager
by multiple goroutines.


## evt/evtq

Package [evtq] creates and manages event queues

## vrtime

Package vrtime defines and manages virtual time inside a simulator.
Time is tracked as an integral number of ticks since
the epoch, along with a secondary sort value to provide for
deterministic order among simultaneous events.

## Running tests

To run tests, allow `setup.sh` and `run_tests.sh` to be executable:

`chmod +x setup.sh`

`chmod +x run_tests.sh`

Then run `setup.sh` to create and activate the virtual environment and install dependencies

`source setup.sh`

Finally, run `run_tests.sh` to run the tests

`./run_tests.sh`

Copyright 2024 Board of Trustees of the University of Illinois.
See [the license](LICENSE) for details.
