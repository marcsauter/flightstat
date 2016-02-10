// Copyright ©2016 Marc Sauter <marc.sauter@bluewin.ch>
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package flightstat provides data structures and functions to build and wirite flight statistics
package flightstat

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/marcsauter/igc"
	"github.com/tealeg/xlsx"
)

// FlightStat represents the flight statistics
type FlightStat struct {
	Airtime       time.Duration
	Flights       int
	Glider        map[string]Glider
	Year          map[int]FlightStatYear
	DefaultGlider string
}

//
type Glider struct {
	Name    string
	Flights int
	Airtime time.Duration
}

// NewFlightStat creates the FlightStat for the collection of flights
func NewFlightStat(flights *igc.Flights, glider string) (*FlightStat, error) {
	stat := &FlightStat{
		Glider:        make(map[string]Glider),
		Year:          make(map[int]FlightStatYear),
		DefaultGlider: glider,
	}
	for _, f := range *flights {
		if err := stat.Add(f); err != nil {
			return nil, err
		}
	}
	return stat, nil
}

// Add a new flight to the statistics
func (fs *FlightStat) Add(f *igc.Flight) error {
	fs.Airtime += f.Duration
	fs.Flights++
	//
	if len(f.Glider) == 0 {
		f.Glider = fs.DefaultGlider
	}
	// update glider statistics
	g, ok := fs.Glider[f.Glider]
	if !ok {
		g = Glider{}
		g.Name = f.Glider
	}
	g.Flights++
	g.Airtime += f.Duration
	fs.Glider[f.Glider] = g
	//
	y, yok := fs.Year[f.TakeOff.Time.Year()]
	if !yok {
		fs.Year[f.TakeOff.Time.Year()] = FlightStatYear{Month: make(map[time.Month]FlightStatMonth)}
		y = fs.Year[f.TakeOff.Time.Year()]
	}
	if err := y.Add(f); err != nil {
		return err
	}
	fs.Year[f.TakeOff.Time.Year()] = y
	return nil
}

// Csv writes the header, calls each year and then writes the year statistics in csv format
func (fs *FlightStat) Csv(w *csv.Writer) {
	year := []int{}
	for y, _ := range fs.Year {
		year = append(year, y)
	}
	sort.Ints(year)
	// write header
	w.Write([]string{"Period", "Flights", "Airtime"})
	// statistics
	for _, y := range year {
		fsy := fs.Year[y]
		fsy.Csv(w)
	}
	w.Write([]string{"Total", fmt.Sprintf("%d", fs.Flights), fmt.Sprintf("%.2f", fs.Airtime.Minutes())})
	w.Write([]string{"Glider", "Flights", "Airtime"})
	for _, g := range fs.Glider {
		w.Write([]string{g.Name, fmt.Sprintf("%d", g.Flights), fmt.Sprintf("%.2f", g.Airtime.Minutes())})
	}
}

// Xlsx writes the header, calls each year then writes the year statistics in xlsx format
func (fs *FlightStat) Xlsx(s *xlsx.Sheet) {
	year := []int{}
	for y, _ := range fs.Year {
		year = append(year, y)
	}
	sort.Ints(year)
	// write header
	r1 := s.AddRow()
	ti := r1.AddCell()
	ti.Merge(2, 0)
	ti.SetString("Statistics")
	r2 := s.AddRow()
	r2.AddCell().SetString("Period")
	r2.AddCell().SetString("Flights")
	r2.AddCell().SetString("Airtime")
	// statistics
	for _, y := range year {
		fsy := fs.Year[y]
		fsy.Xlsx(s)
	}
	rt := s.AddRow()
	rt.AddCell().SetString("Total")
	rt.AddCell().SetInt(fs.Flights)
	rt.AddCell().SetFloatWithFormat(fs.Airtime.Minutes(), "0.00")
	s.AddRow()
	rg := s.AddRow()
	rg.AddCell().SetString("Glider")
	rg.AddCell().SetString("Flights")
	rg.AddCell().SetString("Airtime")
	for _, g := range fs.Glider {
		r := s.AddRow()
		r.AddCell().SetString(g.Name)
		r.AddCell().SetInt(g.Flights)
		r.AddCell().SetFloatWithFormat(g.Airtime.Minutes(), "0.00")
	}
}

// FlightStatYear represents the year statistics
type FlightStatYear struct {
	Year    time.Time
	Airtime time.Duration
	Flights int
	Month   map[time.Month]FlightStatMonth
}

// Add a flight to the year statistics
func (fsy *FlightStatYear) Add(f *igc.Flight) error {
	fsy.Year = f.TakeOff.Time
	fsy.Airtime += f.Duration
	fsy.Flights++
	m, mok := fsy.Month[f.TakeOff.Time.Month()]
	if !mok {
		fsy.Month[f.TakeOff.Time.Month()] = FlightStatMonth{Day: make(map[int]FlightStatDay)}
		m = fsy.Month[f.TakeOff.Time.Month()]
	}
	if err := m.Add(f); err != nil {
		return err
	}
	fsy.Month[f.TakeOff.Time.Month()] = m
	return nil
}

// Csv calls each month and then writes the year statistics in csv format
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

// Xlsx calls each month and then writes the year statistics in xlsx format
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

// FlightStatMonth represents the month statistics
type FlightStatMonth struct {
	Date    time.Time
	Airtime time.Duration
	Flights int
	Day     map[int]FlightStatDay
}

// Add a flight to the month statistics
func (fsm *FlightStatMonth) Add(f *igc.Flight) error {
	fsm.Date = f.TakeOff.Time
	fsm.Airtime += f.Duration
	fsm.Flights++
	d, dok := fsm.Day[f.TakeOff.Time.Day()]
	if !dok {
		fsm.Day[f.TakeOff.Time.Day()] = FlightStatDay{}
		d = fsm.Day[f.TakeOff.Time.Day()]
	}
	if err := d.Add(f); err != nil {
		return err
	}
	fsm.Day[f.TakeOff.Time.Day()] = d
	return nil
}

// Csv calls each day and then writes the month statistics in csv format
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

// Xlsx calls each day and then writes the month statistics in xlsx format
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

// FlightStatMonth represents the day statistics
type FlightStatDay struct {
	Date    time.Time
	Airtime time.Duration
	Flights int
}

// Add a flight to the day statistics
func (fsd *FlightStatDay) Add(f *igc.Flight) error {
	fsd.Date = f.TakeOff.Time
	fsd.Airtime += f.Duration
	fsd.Flights++
	return nil
}

// Csv writes the day statistics in csv format
func (fsd *FlightStatDay) Csv(w *csv.Writer) {
	w.Write([]string{fsd.Date.Format("02.01.2006"), fmt.Sprintf("%d", fsd.Flights), fmt.Sprintf("%.2f", fsd.Airtime.Minutes())})
}

// Xlsx writes the day statistics in xlsx format
func (fsd *FlightStatDay) Xlsx(s *xlsx.Sheet) {
	r := s.AddRow()
	c := r.AddCell()
	c.SetDate(fsd.Date)
	c.NumFmt = "dd.mm.yyyy"
	r.AddCell().SetInt(fsd.Flights)
	r.AddCell().SetFloatWithFormat(fsd.Airtime.Minutes(), "0.00")
}

//
func Csv(flights *igc.Flights, stat *FlightStat, filename string) {
	// write file
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	// flights
	w.Write([]string{"Date", "Takeoff", "Takeoff Site", "Takeoff Coord", "Landing", "Landing Site", "Landing Coord", "Airtime", "Glider", "Filename"})
	for _, f := range *flights {
		w.Write([]string{f.Date.Format("02.01.2006"), f.TakeOff.Time.Format("15:04"), f.TakeOffSite, f.TakeOff.Coord(), f.Landing.Time.Format("15:04"), f.LandingSite, f.Landing.Coord(), fmt.Sprintf("%.2f", f.Duration.Minutes()), f.Glider, f.Filename})
	}
	// statistics
	stat.Csv(w)
	w.Flush()
}

//
func Xlsx(flights *igc.Flights, stat *FlightStat, filename string) {
	// write file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Flight Statistics")
	if err != nil {
		log.Fatal(err)
	}
	// header
	// 1st line/row
	r0 := sheet.AddRow()
	ti := r0.AddCell()
	ti.Merge(9, 0) // merge with the following 6 cells
	ti.SetString("Flights")
	// 2nd line/row
	r1 := sheet.AddRow()
	r1.AddCell().SetString("Date")
	//
	to := r1.AddCell()
	to.Merge(2, 0) // merge with the following cell
	to.SetString("Takeoff")
	r1.AddCell() // cell to merge
	r1.AddCell() // cell to merge
	//
	la := r1.AddCell()
	la.Merge(2, 0) // merge with the following cell
	la.SetString("Landing")
	r1.AddCell() // cell to merge
	r1.AddCell() // cell to merge
	//
	r1.AddCell().SetString("Airtime")
	r1.AddCell().SetString("Glider")
	r1.AddCell().SetString("Filename")
	// 3rd line/row
	r2 := sheet.AddRow()
	r2.AddCell() // start with an empty cell
	r2.AddCell().SetString("Time")
	r2.AddCell().SetString("Site")
	r2.AddCell().SetString("Coord")
	r2.AddCell().SetString("Time")
	r2.AddCell().SetString("Site")
	r2.AddCell().SetString("Coord")
	// flights
	for _, f := range *flights {
		r := sheet.AddRow()
		//
		c1 := r.AddCell()
		c1.SetDate(f.Date)
		c1.NumFmt = "dd.mm.yyyy"
		//
		c2 := r.AddCell()
		c2.SetDateTime(f.TakeOff.Time)
		c2.NumFmt = "hh:mm"
		//
		r.AddCell().SetString(f.TakeOffSite)
		r.AddCell().SetString(f.TakeOff.Coord())
		//
		c3 := r.AddCell()
		c3.SetDateTime(f.Landing.Time)
		c3.NumFmt = "hh:mm"
		//
		r.AddCell().SetString(f.LandingSite)
		r.AddCell().SetString(f.Landing.Coord())
		//
		r.AddCell().SetFloatWithFormat(f.Duration.Minutes(), "0.00")
		//
		r.AddCell().SetString(f.Glider)
		//
		if len(f.Comment) > 0 {
			r.AddCell().SetString(f.Comment)
		} else {
			r.AddCell().SetString(f.Filename)
		}
	}
	sheet.AddRow()
	stat.Xlsx(sheet)
	err = file.Save(filename)
	if err != nil {
		log.Fatal(err)
	}
}
