package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	Base      = "https://api.weather.gov"
	RefTime   = "2006-01-02T15:04:00-07:00"
	ShortTime = "01-02-2006 15:04"
)

// Flags
var (
	ForecastOffice = flag.String("f", "DMX", "forecast office code")
	GridX          = flag.Int("x", 63, "grid x coordinate")
	GridY          = flag.Int("y", 26, "grid y coordinate")
	HiLoInt        = flag.Int("r", 12, "range in hours for high and low")
)

var (
	api string
	tmp = filepath.Join(os.TempDir(), "weatherbar.json")
)

func init() {
	flag.Parse()
	api = fmt.Sprintf("%s/gridpoints/%s/%d,%d/forecast/hourly",
		Base, *ForecastOffice, *GridX, *GridY)
}

type Outer struct {
	Properties Property
}

type Property struct {
	Periods []Period
}

type Period struct {
	WindDirection    string
	WindSpeed        string
	Name             string
	StartTime        string
	EndTime          string
	TemperatureTrend string
	ShortForecast    string
	DetailedForecast string
	Temperature      int
	Number           int
	IsDaytime        bool
}

func WriteCache(byts []byte) {
	f, err := os.Create(tmp)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	f.Write(byts)
}

func LoadCache() []byte {
	f, err := os.Open(tmp)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	byts, _ := io.ReadAll(f)
	return byts
}

func main() {
	// if something panics, just exit gracefully
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprint(os.Stderr, r)
			fmt.Print("null")
			os.Exit(0)
		}
	}()
	client := http.DefaultClient
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		panic(err)
	}
	req.Header["User-Agent"] = []string{
		"weatherbar",
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	var byts []byte
	if resp.StatusCode != 200 {
		byts = LoadCache()
	} else {
		byts, err = io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		WriteCache(byts)
	}
	data := new(Outer)
	json.Unmarshal(byts, data)
	// Assume Periods[0] is current time -> 12 hour high/low by
	// finding max/min in that range
	end := *HiLoInt
	if l := len(data.Properties.Periods); l < end {
		end = l
	}
	now := data.Properties.Periods[0]
	high := now.Temperature
	low := now.Temperature
	for _, p := range data.Properties.Periods[:end] {
		if p.Temperature > high {
			high = p.Temperature
		}
		if p.Temperature < low {
			low = p.Temperature
		}
	}
	fmt.Printf("Hi:%d Lo:%d %s %s Cur: %d°F",
		high, low,
		now.WindDirection, now.WindSpeed,
		now.Temperature)
}
