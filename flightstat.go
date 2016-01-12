package flightstat

import (
	"fmt"
	"sort"
	"time"

	"github.com/marcsauter/igc"
)

//
type FlightStat struct {
	Airtime time.Duration
	Flights int
	Year    map[int]FlightStatYear
}

//
func NewFlightStat(flights *igc.Flights) (*FlightStat, error) {
	stat := &FlightStat{Year: make(map[int]FlightStatYear)}
	for _, f := range *flights {
		if err := stat.Add(f); err != nil {
			return nil, err
		}
	}
	return stat, nil
}

//
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

//
func (fs *FlightStat) Output() *[][]string {
	s := [][]string{}
	year := []int{}
	for y, _ := range fs.Year {
		year = append(year, y)
	}
	sort.Ints(year)
	for _, y := range year {
		v := fs.Year[y]
		s = append(s, v.Output()...)
	}
	s = append(s, fs.Record())
	return &s
}

//
func (fs *FlightStat) Record() []string {
	return []string{"Total", fmt.Sprintf("%d", fs.Flights), fmt.Sprintf("%.2f", fs.Airtime.Minutes())}
}

//
type FlightStatYear struct {
	Year    time.Time
	Airtime time.Duration
	Flights int
	Month   map[time.Month]FlightStatMonth
}

//
func (fsy *FlightStatYear) Add(f *igc.Flight) error {
	fsy.Year = f.TakeOff
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

//
func (fsy *FlightStatYear) Output() [][]string {
	s := [][]string{}
	month := []int{}
	for m, _ := range fsy.Month {
		month = append(month, int(m))
	}
	sort.Ints(month)
	for _, m := range month {
		v := fsy.Month[time.Month(m)]
		s = append(s, v.Output()...)
	}
	return append(s, fsy.Record())
}

//
func (fsy *FlightStatYear) Record() []string {
	return []string{fmt.Sprintf("Total %d", fsy.Year.Year()), fmt.Sprintf("%d", fsy.Flights), fmt.Sprintf("%.2f", fsy.Airtime.Minutes())}
}

//
type FlightStatMonth struct {
	Date    time.Time
	Airtime time.Duration
	Flights int
	Day     map[int]FlightStatDay
}

//
func (fsm *FlightStatMonth) Add(f *igc.Flight) error {
	fsm.Date = f.TakeOff
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

//
func (fsm *FlightStatMonth) Output() [][]string {
	s := [][]string{}
	days := []int{}
	for d, _ := range fsm.Day {
		days = append(days, d)
	}
	sort.Ints(days)
	for _, d := range days {
		v := fsm.Day[d]
		s = append(s, v.Record())
	}
	return append(s, fsm.Record())
}

//
func (fsm *FlightStatMonth) Record() []string {
	return []string{fsm.Date.Format("January 2006"), fmt.Sprintf("%d", fsm.Flights), fmt.Sprintf("%.2f", fsm.Airtime.Minutes())}
}

//
type FlightStatDay struct {
	Date    time.Time
	Airtime time.Duration
	Flights int
}

//
func (fsd *FlightStatDay) Add(f *igc.Flight) error {
	fsd.Date = f.TakeOff
	fsd.Airtime += f.Duration
	fsd.Flights++
	return nil
}

//
func (fsd *FlightStatDay) Record() []string {
	return []string{fsd.Date.Format("02.01.2006"), fmt.Sprintf("%d", fsd.Flights), fmt.Sprintf("%.2f", fsd.Airtime.Minutes())}
}
