package flightstat

import (
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/marcsauter/igc"
	"github.com/tealeg/xlsx"
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
func (fs *FlightStat) Csv(w *csv.Writer) {
	year := []int{}
	for y, _ := range fs.Year {
		year = append(year, y)
	}
	sort.Ints(year)
	// header
	w.Write([]string{"Period", "Flights", "Duration"})
	// statistics
	for _, y := range year {
		fsy := fs.Year[y]
		fsy.Csv(w)
	}
	w.Write([]string{"Total", fmt.Sprintf("%d", fs.Flights), fmt.Sprintf("%.2f", fs.Airtime.Minutes())})
}

//
func (fs *FlightStat) Xlsx(s *xlsx.Sheet) {
	year := []int{}
	for y, _ := range fs.Year {
		year = append(year, y)
	}
	sort.Ints(year)
	// header
	r1 := s.AddRow()
	ti := r1.AddCell()
	ti.Merge(2, 0)
	ti.SetString("Statistics")
	r2 := s.AddRow()
	r2.AddCell().SetString("Period")
	r2.AddCell().SetString("Flights")
	r2.AddCell().SetString("Duration")
	// statistics
	for _, y := range year {
		fsy := fs.Year[y]
		fsy.Xlsx(s)
	}
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
func (fsy *FlightStatYear) Csv(w *csv.Writer) {
	month := []int{}
	for m, _ := range fsy.Month {
		month = append(month, int(m))
	}
	sort.Ints(month)
	for _, m := range month {
		fsm := fsy.Month[time.Month(m)]
		fsm.Csv(w)
	}
	w.Write([]string{fmt.Sprintf("Total %d", fsy.Year.Year()), fmt.Sprintf("%d", fsy.Flights), fmt.Sprintf("%.2f", fsy.Airtime.Minutes())})
}

//
func (fsy *FlightStatYear) Xlsx(s *xlsx.Sheet) {
	month := []int{}
	for m, _ := range fsy.Month {
		month = append(month, int(m))
	}
	sort.Ints(month)
	for _, m := range month {
		fsm := fsy.Month[time.Month(m)]
		fsm.Xlsx(s)
	}
	r := s.AddRow()
	r.AddCell().SetString(fmt.Sprintf("Total %d", fsy.Year.Year()))
	r.AddCell().SetInt(fsy.Flights)
	r.AddCell().SetFloatWithFormat(fsy.Airtime.Minutes(), "0.00")
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
func (fsm *FlightStatMonth) Csv(w *csv.Writer) {
	days := []int{}
	for d, _ := range fsm.Day {
		days = append(days, d)
	}
	sort.Ints(days)
	for _, d := range days {
		fsd := fsm.Day[d]
		fsd.Csv(w)
	}
	w.Write([]string{fmt.Sprintf("Total %s", fsm.Date.Format("January 2006")), fmt.Sprintf("%d", fsm.Flights), fmt.Sprintf("%.2f", fsm.Airtime.Minutes())})
}

//
func (fsm *FlightStatMonth) Xlsx(s *xlsx.Sheet) {
	days := []int{}
	for d, _ := range fsm.Day {
		days = append(days, d)
	}
	sort.Ints(days)
	for _, d := range days {
		fsd := fsm.Day[d]
		fsd.Xlsx(s)
	}
	r := s.AddRow()
	r.AddCell().SetString(fmt.Sprintf("Total %s", fsm.Date.Format("January 2006")))
	r.AddCell().SetInt(fsm.Flights)
	r.AddCell().SetFloatWithFormat(fsm.Airtime.Minutes(), "0.00")
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
func (fsd *FlightStatDay) Csv(w *csv.Writer) {
	w.Write([]string{fsd.Date.Format("02.01.2006"), fmt.Sprintf("%d", fsd.Flights), fmt.Sprintf("%.2f", fsd.Airtime.Minutes())})
}

//
func (fsd *FlightStatDay) Xlsx(s *xlsx.Sheet) {
	r := s.AddRow()
	c := r.AddCell()
	c.SetDate(fsd.Date)
	c.NumFmt = "dd.mm.yyyy"
	r.AddCell().SetInt(fsd.Flights)
	r.AddCell().SetFloatWithFormat(fsd.Airtime.Minutes(), "0.00")
}
