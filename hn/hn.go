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

	timeLayout     = "02-01-2006 15:04:05"
	isiiTimeLayout = "2006-01-02;15:04"
)

var (
	errInvalidTimeValue   = errors.New("Can't parse time value line")
	errInvalidTime        = errors.New("Invalid time")
	errInvalidSetting     = errors.New("Invalid setting")
	errInvalidGroup       = errors.New("Invalid group")
	errInvalidSampleCount = errors.New("Invalid sample count")
	errNoContent          = errors.New("No content")
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

// Merge merges a float of string section.
func (section *Section) Merge(interval int) error {
	if len(section.Data) == 0 {
		return nil
	}
	if _, err := strconv.ParseFloat(section.Data[0].Value, 32); err == nil {
		return section.MergeFloat(interval)
	}
	return section.MergeString(interval)
}

// MergeString merges string data section by taking the first element of each interval.
func (section *Section) MergeString(interval int) error {
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

// MergeFloat merges float data section. Average value of each interval.
func (section *Section) MergeFloat(interval int) error {
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
			newData[i].Value = strconv.FormatFloat(float64(v), 'f', -1, 32)
			tValue = 0.0
			curElem = 0
			i++
		}
	}
	if curElem != 0 {
		v := tValue / float32(curElem)
		newData[i].Value = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	section.Data = newData
	section.SampleCount = newSize
	return nil
}

// Write writes the section on the writer (Isii format).
func (section *Section) Write(w io.Writer) {
	fmt.Fprintln(w, prefixSetting, section.Setting)
	fmt.Fprintln(w, prefixGroup, section.Group)
	fmt.Fprintln(w, prefixSampleCount, section.SampleCount)
	for _, data := range section.Data {
		fmt.Fprintln(w, data.Time, data.Value)
	}
	fmt.Fprintln(w)
}

// Write writes the sections on the writer (Isii format).
func (sections Sections) Write(w io.Writer) {
	for _, section := range sections {
		section.Write(w)
	}
}

// WriteStd writes the sections on the writer (Econmic format).
func (sections Sections) WriteStd(w io.Writer, filename string) {
	fmt.Fprintln(w, filename)
	fmt.Fprintln(w, "yyyy-mm-dd")
	fmt.Fprintln(w, "hh:mm")
	for _, section := range sections {
		fmt.Fprintf(w, "%s;826409204;%s\n", section.Group, section.Setting)
	}
	fmt.Fprintln(w, "")
	for i, data := range sections[0].Data {
		fmt.Fprint(w, data.Time.Format(isiiTimeLayout))
		for _, section := range sections {
			fmt.Fprintf(w, ";%s", section.Data[i].Value)
		}
		fmt.Fprintln(w, "")
	}
}

// Merge merges all the sections. See Merge section for more details.
func (sections Sections) Merge(interval int) error {
	for _, section := range sections {
		if err := section.Merge(interval); err != nil {
			return err
		}
	}
	return nil
}
