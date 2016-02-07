package flightstat

import (
	"testing"
	"time"

	"github.com/marcsauter/igc"
)

// Minimal statistics test
func Test(t *testing.T) {
	flights := igc.NewFlights()
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 11)})
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 10, 5, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 12)})
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 11, 2, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 13)})
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 11, 6, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 14)})
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 12, 3, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 15)})
	flights.Add(&igc.Flight{TakeOff: time.Date(2016, 12, 7, 0, 0, 0, 0, time.UTC), Duration: time.Duration(time.Minute * 16)})
	s, err := NewFlightStat(flights)
	if err != nil {
		t.Error(err)
	}
	if flights.Len() != 6 || s.Airtime != time.Duration(time.Minute*81) {
		t.Error("statistic does not match")
	}
}
