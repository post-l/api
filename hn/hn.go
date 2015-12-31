package hn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	prefixSetting     = "Setting    : "
	prefixGroup       = "Group      : "
	prefixSampleCount = "SampleCount: "

	timeLayout         = "02-01-2006 15:04:05"
	economicTimeLayout = "2006-01-02;15:04"

	economicSep = ";"
)

var (
	errInvalidTimeValue     = errors.New("Can't parse time value line")
	errInvalidTime          = errors.New("Invalid time")
	errInvalidSetting       = errors.New("Invalid setting")
	errInvalidGroup         = errors.New("Invalid group")
	errInvalidSampleCount   = errors.New("Invalid sample count")
	errNoContent            = errors.New("No content")
	errInvalidSectionHeader = errors.New("Invalid section header")
	errInvalidSectionData   = errors.New("Invalid section data")
)

// Sections is a slice of section with some useful methods.
type Sections []*Section

// Section represents a isii section.
type Section struct {
	Setting     string
	Group       string
	SampleCount int
	Data        []TimeValue
}

// TimeValue associates a time to a string value.
type TimeValue struct {
	Time  time.Time
	Value string
}

// ParseEconomicFile parses an economic file to get the sections.
func ParseEconomicFile(reader io.Reader) (Sections, error) {
	scanner := bufio.NewScanner(reader)
	for i := 0; i < 3; i++ {
		if !scanner.Scan() {
			return nil, errNoContent
		}
	}
	var sections Sections
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			break
		}
		line := scanner.Text()
		row := strings.Split(line, economicSep)
		if len(row) != 3 {
			return nil, errInvalidSectionHeader
		}
		sections = append(sections, &Section{
			Setting: row[2],
			Group:   row[0],
		})
	}
	if len(sections) == 0 {
		return nil, errNoContent
	}
	err := sections.parseEconomicData(scanner)
	return sections, err
}

// ParseIsiiFile parses an isii file to get the sections.
func ParseIsiiFile(reader io.Reader) (Sections, error) {
	var sections Sections
	scanner := bufio.NewScanner(reader)
	for lineIndex := 1; scanner.Scan(); lineIndex++ {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		section, sectionLen, err := newIsiiSection(scanner, &lineIndex)
		lineIndex += sectionLen
		if err != nil {
			return nil, fmt.Errorf("Can't parse line %d: %s", lineIndex, err)
		}
		sections = append(sections, section)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Invalid File: %s", err)
	}
	if len(sections) == 0 {
		return nil, errNoContent
	}
	return sections, nil
}

func newIsiiSection(scanner *bufio.Scanner, lineIndex *int) (*Section, int, error) {
	section := new(Section)
	sectionIndex, err := section.parseIsiiSectionHeader(scanner)
	if err != nil {
		return nil, sectionIndex, err
	}
	i := 0
	for scanner.Scan() {
		sectionIndex++
		line := scanner.Text()
		if line == "" {
			break
		}
		spaceIndex := strings.LastIndex(line, " ")
		if spaceIndex == -1 {
			return nil, sectionIndex, errInvalidTimeValue
		}
		t, err := time.Parse(timeLayout, line[:spaceIndex])
		if err != nil {
			return nil, sectionIndex, errInvalidTime
		}
		section.Data[i].Time = t
		section.Data[i].Value = line[spaceIndex+1:]
		i++
	}
	return section, sectionIndex, nil
}

func (section *Section) parseIsiiSectionHeader(scanner *bufio.Scanner) (int, error) {
	setting := scanner.Text()
	if !strings.HasPrefix(setting, prefixSetting) {
		return 0, errInvalidSetting
	}
	if !scanner.Scan() {
		return 1, errInvalidGroup
	}
	group := scanner.Text()
	if !strings.HasPrefix(group, prefixGroup) {
		return 1, errInvalidGroup
	}
	if !scanner.Scan() {
		return 2, errInvalidSampleCount
	}
	sampleCountStr := scanner.Text()
	if !strings.HasPrefix(sampleCountStr, prefixSampleCount) {
		return 2, errInvalidSampleCount
	}
	sampleCount, err := strconv.Atoi(sampleCountStr[len(prefixSampleCount):])
	if err != nil {
		return 2, errInvalidSampleCount
	}
	section.Setting = setting[len(prefixSetting):]
	section.Group = group[len(prefixGroup):]
	section.SampleCount = sampleCount
	section.Data = make([]TimeValue, sampleCount)
	return 2, nil
}

// Average creates a new data section with the average for a given interval.
func (section *Section) Average(interval int) error {
	if len(section.Data) == 0 {
		return nil
	}
	if _, err := strconv.ParseFloat(section.Data[0].Value, 32); err == nil {
		return section.AverageFloat(interval)
	}
	return section.AverageString(interval)
}

// AverageString takes the first element of each interval.
func (section *Section) AverageString(interval int) error {
	newSize := (len(section.Data) / interval) + len(section.Data)%2
	newData := make([]TimeValue, newSize)
	curElem := 0
	i := 0
	for _, data := range section.Data {
		if curElem == 0 {
			newData[i].Time = data.Time
			newData[i].Value = data.Value
		}
		curElem++
		if curElem == interval {
			curElem = 0
			i++
		}
	}
	section.Data = newData
	section.SampleCount = newSize
	return nil
}

// AverageFloat averages float data section for each interval.
func (section *Section) AverageFloat(interval int) error {
	newSize := (len(section.Data) / interval) + len(section.Data)%2
	newData := make([]TimeValue, newSize)
	curElem := 0
	i := 0
	var tValue float32
	for _, data := range section.Data {
		if curElem == 0 {
			newData[i].Time = data.Time
		}
		v, err := strconv.ParseFloat(data.Value, 32)
		if err != nil {
			return err
		}
		tValue += float32(v)
		curElem++
		if curElem == interval {
			v := tValue / float32(curElem)
			newData[i].Value = strconv.FormatFloat(float64(v), 'f', 1, 32)
			tValue = 0.0
			curElem = 0
			i++
		}
	}
	if curElem != 0 {
		v := tValue / float32(curElem)
		newData[i].Value = strconv.FormatFloat(float64(v), 'f', 1, 32)
	}
	section.Data = newData
	section.SampleCount = newSize
	return nil
}

// WriteIsii writes the section on the writer (Isii format).
func (section *Section) WriteIsii(w io.Writer) {
	fmt.Fprintln(w, prefixSetting, section.Setting)
	fmt.Fprintln(w, prefixGroup, section.Group)
	fmt.Fprintln(w, prefixSampleCount, section.SampleCount)
	for _, data := range section.Data {
		fmt.Fprintln(w, data.Time, data.Value)
	}
	fmt.Fprintln(w)
}

// WriteIsii writes the sections on the writer (Isii format).
func (sections Sections) WriteIsii(w io.Writer) {
	for _, section := range sections {
		section.WriteIsii(w)
	}
}

// WriteEconomic writes the sections on the writer (Econmic format).
func (sections Sections) WriteEconomic(w io.Writer, filename string) {
	fmt.Fprintf(w, "%s\r\nyyyy-mm-dd\r\nhh:mm\r\n", filename)
	for _, section := range sections {
		fmt.Fprintf(w, "%s;000000000;%s\r\n", section.Group, section.Setting)
	}
	fmt.Fprint(w, "\r\n")
	for i, data := range sections[0].Data {
		fmt.Fprint(w, data.Time.Format(economicTimeLayout))
		for _, section := range sections {
			fmt.Fprintf(w, ";%s", section.Data[i].Value)
		}
		fmt.Fprint(w, "\r\n")
	}
}

// Average calculates the average of all the sections. See Average section for more details.
func (sections Sections) Average(interval int) error {
	for _, section := range sections {
		if err := section.Average(interval); err != nil {
			return err
		}
	}
	return nil
}

func (sections Sections) parseEconomicData(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) <= len(economicTimeLayout)+1 {
			return errInvalidSectionData
		}
		timeStr := line[:len(economicTimeLayout)]
		data := line[len(economicTimeLayout)+1:]
		row := strings.Split(data, economicSep)
		if len(row) != len(sections) {
			return errInvalidSectionData
		}
		t, err := time.Parse(economicTimeLayout, timeStr)
		if err != nil {
			return errInvalidTime
		}
		for i, section := range sections {
			section.Data = append(section.Data, TimeValue{
				Time:  t,
				Value: row[i],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Invalid File: %s", err)
	}
	// Legacy Sample Count...
	for _, section := range sections {
		section.SampleCount = len(section.Data)
	}
	return nil
}
