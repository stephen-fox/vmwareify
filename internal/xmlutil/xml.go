package xmlutil

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
)

// FindObjectConfig provides configuration for finding XML objects in a
// given document.
type FindObjectConfig interface {
	// Start returns the xml.StartElement for the XML object that
	// is being searched for.
	Start() *xml.StartElement

	// Scanner returns the bufio.Scanner that contains the XML
	// document's data.
	Scanner() *bufio.Scanner

	// Eol returns the document's end of line characters (e.g., '\n').
	Eol() []byte
}

type defaultFindObjectConfig struct {
	start   *xml.StartElement
	scanner *bufio.Scanner
	eol     []byte
}

func (o defaultFindObjectConfig) Start() *xml.StartElement {
	return o.start
}

func (o defaultFindObjectConfig) Scanner() *bufio.Scanner {
	return o.scanner
}

func (o defaultFindObjectConfig) Eol() []byte {
	return o.eol
}

// RawObject represents one serialized XML object. It provides helpful
// functions for building a new XML object off of it.
type RawObject interface {
	// Data returns a bytes.Buffer containing the raw
	// XML object's data.
	Data() *bytes.Buffer

	// StartAndEndLinePrefix returns the string that prefixes the
	// first and last lines of the object.
	StartAndEndLinePrefix() string

	// BodyPrefix returns the string that prefixes the body of the
	// XML object.
	BodyPrefix() string

	// RelativeBodyPrefix returns the string that prefixes the body
	// of the XML object relative to the prefix of the first and
	// last lines of the object.
	//
	// For example, if the first and last lines are prefixed by
	// six spaces, and the body is prefixed by eight spaces, the
	// function will only return two spaces.
	RelativeBodyPrefix() string
}

type defaultRawObject struct {
	data               *bytes.Buffer
	initialIndentCount int
	bodyIndentCount    int
	indentChar         rune
}

func (o defaultRawObject) Data() *bytes.Buffer {
	return o.data
}

func (o defaultRawObject) StartAndEndLinePrefix() string {
	return strings.Repeat(string(o.indentChar), o.initialIndentCount)
}

func (o defaultRawObject) BodyPrefix() string {
	return strings.Repeat(string(o.indentChar), o.bodyIndentCount)
}

func (o defaultRawObject) RelativeBodyPrefix() string {
	difference := o.bodyIndentCount - o.initialIndentCount

	if difference < 0 {
		return ""
	}

	return strings.Repeat(string(o.indentChar), difference)
}

// ValidateFormatting returns a non-nil error if the provided slice of bytes
// is not a valid XML document.
func ValidateFormatting(raw []byte) error {
	var temp struct{}

	err := xml.Unmarshal(raw, &temp)
	if err != nil {
		return err
	}

	return nil
}

// IsStartElement returns true and a pointer to the xml.StartElement if the
// provided line is a valid XML start element.
func IsStartElement(line []byte) (*xml.StartElement, bool) {
	d := xml.NewDecoder(bytes.NewReader(bytes.TrimSpace(line)))

	t, err := d.RawToken()
	if err != nil {
		return &xml.StartElement{}, false
	}

	if t == nil {
		return &xml.StartElement{}, false
	}

	v, ok := t.(xml.StartElement)
	if ok {
		return &v, true
	}

	return &xml.StartElement{}, false
}

// NewFindObjectConfig returns a new instance of FindObjectConfig, which is used for
// searching XML documents for specific objects.
func NewFindObjectConfig(start *xml.StartElement, scanner *bufio.Scanner, eol []byte) (FindObjectConfig, error) {
	if start == nil {
		return &defaultFindObjectConfig{}, errors.New("a nil xml.StartElement was provided")
	}

	if scanner == nil {
		return &defaultFindObjectConfig{}, errors.New("a nil bufio.Scanner was provided")
	}

	return &defaultFindObjectConfig{
		start:   start,
		scanner: scanner,
		eol:     eol,
	}, nil
}

// FindAndDeserializeObject searches the provided document for a XML object
// matching the provided xml.StartElement. It then deserializes (unmarshals)
// the raw data into the provided pointer.
func FindAndDeserializeObject(config FindObjectConfig, pointer interface{}) (RawObject, error) {
	rawObject, err := FindObject(config)
	if err != nil {
		return rawObject, err
	}

	err = xml.Unmarshal(rawObject.Data().Bytes(), pointer)
	if err != nil {
		return rawObject, err
	}

	return rawObject, nil
}

// FindObject searches the provided document for a XML object matching
// the provided xml.StartElement. It returns a RawObject representing
// the object.
func FindObject(config FindObjectConfig) (RawObject, error) {
	firstLine := config.Scanner().Text()
	indentChar, count := lineIndentInfo(firstLine)
	rawObject := &defaultRawObject{
		data:               bytes.NewBuffer(nil),
		initialIndentCount: count,
		indentChar:         indentChar,
	}

	rawObject.data.WriteString(firstLine)

	checkedBodyIntent := false

	for config.Scanner().Scan() {
		text := config.Scanner().Text()

		if !checkedBodyIntent {
			checkedBodyIntent = true
			_, count := lineIndentInfo(text)
			rawObject.bodyIndentCount = count
		}

		rawObject.data.Write(config.Eol())

		end, isEnd := IsEndElement(text)
		rawObject.data.WriteString(text)
		if isEnd && end.Name.Local == config.Start().Name.Local {
			break
		}
	}

	err := config.Scanner().Err()
	if err != nil {
		return rawObject, err
	}

	err = ValidateFormatting(rawObject.data.Bytes())
	if err != nil {
		return rawObject, err
	}

	return rawObject, nil
}

func lineIndentInfo(line string) (indentChar rune, count int) {
	if len(line) == 0 {
		return ' ', 0
	}

	indentChar = rune(line[0])

	indents := 0

	for i := range line {
		if rune(line[i]) == indentChar {
			indents = indents + 1
		} else {
			break
		}
	}

	return indentChar, indents
}

// IsEndElement returns true and a pointer to the xml.EndElement if the
// provided line is a valid XML end element.
func IsEndElement(line string) (*xml.EndElement, bool) {
	d := xml.NewDecoder(strings.NewReader(strings.TrimSpace(line)))

	t, err := d.RawToken()
	if err != nil {
		return &xml.EndElement{}, false
	}

	if t == nil {
		return &xml.EndElement{}, false
	}

	v, ok := t.(xml.EndElement)
	if ok {
		return &v, true
	}

	return &xml.EndElement{}, false
}
