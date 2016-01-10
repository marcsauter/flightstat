package flightstat

//TODO: sort keys

import (
	"fmt"
	"time"

	"github.com/marcsauter/igc"
)

type FlightStat struct {
	Airtime time.Duration
	Flights int
	Year    map[int]FlightStatYear
}

func NewFlightStat() *FlightStat {
	return &FlightStat{Year: make(map[int]FlightStatYear)}
}

func (fs *FlightStat) Add(f *igc.Flight) error {
	fs.Airtime += f.Duration
	fs.Flights++
	y, yok := fs.Year[f.TakeOff.Year()]
	if !yok {
		fs.Year[f.TakeOff.Year()] = FlightStatYear{Month: make(map[time.Month]FlightStatMonth)}
		y = fs.Year[f.TakeOff.Year()]
	}
	if err := y.Add(f); err != nil {
		return err
	}
	fs.Year[f.TakeOff.Year()] = y
	return nil
}

func (fs *FlightStat) Ouput() [][]string {
	s := [][]string{}
	for _, v := range fs.Year {
		s = append(s, v.Ouput()...)
	}
	return append(s, fs.Record())
}

func (fs *FlightStat) Record() []string {
	return []string{"Total", fmt.Sprintf("%d", fs.Flights), fmt.Sprintf("%.2f", fs.Airtime.Minutes())}
}

type FlightStatYear struct {
	Year    int
	Airtime time.Duration
	Flights int
	Month   map[time.Month]FlightStatMonth
}

func (fsy *FlightStatYear) Add(f *igc.Flight) error {
	fsy.Year = f.TakeOff.Year()
	fsy.Airtime += f.Duration
	fsy.Flights++
	m, mok := fsy.Month[f.TakeOff.Month()]
	if !mok {
		fsy.Month[f.TakeOff.Month()] = FlightStatMonth{Day: make(map[int]FlightStatDay)}
		m = fsy.Month[f.TakeOff.Month()]
	}
	if err := m.Add(f); err != nil {
		return err
	}
	fsy.Month[f.TakeOff.Month()] = m
	return nil
}

func (fsy *FlightStatYear) Ouput() [][]string {
	y := [][]string{}
	for _, v := range fsy.Month {
		y = append(y, v.Ouput()...)
	}
	return append(y, fsy.Record())
}

func (fsy *FlightStatYear) Record() []string {
	return []string{fmt.Sprintf("Total %d:", fsy.Year), fmt.Sprintf("%d", fsy.Flights), fmt.Sprintf("%.2f", fsy.Airtime.Minutes())}
}

type FlightStatMonth struct {
	Month   time.Month
	Airtime time.Duration
	Flights int
	Day     map[int]FlightStatDay
}

func (fsm *FlightStatMonth) Add(f *igc.Flight) error {
	fsm.Month = f.TakeOff.Month()
	fsm.Airtime += f.Duration
	fsm.Flights++
	d, dok := fsm.Day[f.TakeOff.Day()]
	if !dok {
		fsm.Day[f.TakeOff.Day()] = FlightStatDay{}
		d = fsm.Day[f.TakeOff.Day()]
	}
	if err := d.Add(f); err != nil {
		return err
	}
	fsm.Day[f.TakeOff.Day()] = d
	return nil
}

func (fsm *FlightStatMonth) Ouput() [][]string {
	m := [][]string{}
	for _, v := range fsm.Day {
		m = append(m, v.Record())
	}
	return append(m, fsm.Record())
}

func (fsm *FlightStatMonth) Record() []string {
	return []string{fsm.Month.String(), fmt.Sprintf("%d", fsm.Flights), fmt.Sprintf("%.2f", fsm.Airtime.Minutes())}
}

type FlightStatDay struct {
	Day     int
	Airtime time.Duration
	Flights int
}

func (fsd *FlightStatDay) Add(f *igc.Flight) error {
	fsd.Day = f.TakeOff.Day()
	fsd.Airtime += f.Duration
	fsd.Flights++
	return nil
}

func (fsd *FlightStatDay) Record() []string {
	return []string{fmt.Sprintf("%d", fsd.Day), fmt.Sprintf("%d", fsd.Flights), fmt.Sprintf("%.2f", fsd.Airtime.Minutes())}
}
