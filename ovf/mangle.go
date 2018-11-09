package ovf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	numXmlOpenChars  = 2
	numXmlCloseChars = 3
)

const (
	NoOp    Action = "no_op"
	Delete  Action = "delete"
	Replace Action = "replace"
)

type Action string

type ManipulateOptions struct {
	OnHardwareItems []OnHardwareItemsFunc
	DeleteLimit     int
}

type OnHardwareItemsFunc func(Item) (Action, Item)

type mangler struct {
	numCharsDeleted int64
	result          []byte
	r               io.Reader
}

func (o *mangler) Read(p []byte) (n int, err error) {
	n, err = o.r.Read(p)
	if err != nil {
		return n, err
	}

	o.result = append(o.result, p[:n]...)

	return n, err
}

type lineInfo struct {
	abort              bool
	lineStartIndex     int64
	tagStartIndex      int64
	numberOfSpaces     int
	endsWithNewLine    bool
	prevLineHasNewLine bool
}

func (o *mangler) lineInfo(decoderOffset int64) lineInfo {
	//decoderOffset = decoderOffset - o.numCharsDeleted

	openTagIndex := bytes.LastIndex(o.result[:decoderOffset], []byte{'<'})
	if openTagIndex < 0 {
		return lineInfo{
			abort: true,
		}
	}

	previousElementEndIndex := bytes.LastIndex(o.result[:openTagIndex], []byte{'>'})
	if previousElementEndIndex < 0 {
		return lineInfo{
			abort: true,
		}
	}

	prevLineHasNewLine := bytes.Contains(o.result[previousElementEndIndex:openTagIndex], []byte{'\n'})

	numSpaces := bytes.Count(o.result[previousElementEndIndex:openTagIndex], []byte{' '})

	hasNewLine := false
	if decoderOffset < int64(len(o.result)) && o.result[decoderOffset] == '\n' {
		hasNewLine = true
	}

	return lineInfo{
		tagStartIndex:      int64(openTagIndex),
		lineStartIndex:     int64(openTagIndex) - int64(numSpaces),
		prevLineHasNewLine: prevLineHasNewLine,
		endsWithNewLine:    hasNewLine,
		numberOfSpaces:     numSpaces,
	}
}

func (o *mangler) deleteFrom(from int64, to int64) {
	toDelete := to - from
	if toDelete <= 0 {
		return
	}

	to = to - o.numCharsDeleted
	from = from - o.numCharsDeleted
	o.numCharsDeleted = o.numCharsDeleted + toDelete

	fmt.Println("Delete from", from, "to", to)

	o.result = append(o.result[:from], o.result[to:]...)
}

func (o *mangler) replace(from int64, to int64, raw []byte) {
	//fmt.Println("Before delete\n'" + string(o.result) + "'")

	o.deleteFrom(from, to)

	//fmt.Println("After delete\n'" + string(o.result) + "'")
	length := int64(len(o.result))

	if from > length {
		o.result = append(o.result, raw...)
		return
	}

	fmt.Println("Len", length, "- from", from)
	fmt.Println("Deleted", to - from)
	fmt.Println("Inserting", len(raw))
	//fmt.Println("Before\n'" + string(o.result) + "'")
	fmt.Println("Up to from:\n'" + string(o.result[:from]) + "'")

	o.result = append(o.result[:from], append(raw, o.result[from:]...)...)

	//if from > length || to > length {
	//	o.result = append(o.result, raw...)
	//	return
	//}
	//
	//o.result = append(o.result[:from], append(raw, o.result[to:]...)...)
}

func (o *mangler) buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.result)
}

func newMangler(r io.Reader) *mangler {
	return &mangler{
		r: r,
	}
}

func DeleteHardwareItemsMatchingFunc(elementNamePrefixes []string) OnHardwareItemsFunc {
	return func(i Item) (Action, Item) {
		for _, name := range elementNamePrefixes {
			if strings.HasPrefix(i.ElementName, name) {
				return Delete, Item{}
			}
		}

		return NoOp, Item{}
	}
}

func ReplaceHardwareItemFunc(elementName string, item Item) OnHardwareItemsFunc {
	return func(i Item) (Action, Item) {
		if i.ElementName == elementName {
			return Replace, item
		}

		return NoOp, Item{}
	}
}

func Manipulate(r io.Reader, options ManipulateOptions) (*bytes.Buffer, error) {
	mangler := newMangler(r)
	decoder := xml.NewDecoder(mangler)

	for {
		token, err := decoder.Token()
		if err != nil && err != io.EOF  {
			return mangler.buffer(), err
		}
		if token == nil {
			break
		}

		itemFieldName := "Item"

		switch tokenData := token.(type) {
		case xml.StartElement:
			switch tokenData.Name.Local {
			case itemFieldName:
				startLine := mangler.lineInfo(decoder.InputOffset())

				var item Item

				err := decoder.DecodeElement(&item, &tokenData)
				if err != nil {
					return mangler.buffer(), err
				}

				for _, f := range options.OnHardwareItems {
					action, result := f(item)
					switch action {
					case NoOp:
						continue
					case Delete:
						// TODO: Deal with hardcoded deletion of new lines (offset + 1).
						mangler.deleteFrom(startLine.lineStartIndex, decoder.InputOffset() + 1)
						break
					case Replace:
						endLine := mangler.lineInfo(decoder.InputOffset())
						raw, err := xml.MarshalIndent(result, strings.Repeat(" ", endLine.numberOfSpaces), "  ")
						if err != nil {
							return mangler.buffer(), err
						}

						mangler.replace(startLine.lineStartIndex, decoder.InputOffset(), raw)
						break
					}
				}
			}
		}
	}

	return mangler.buffer(), nil
}
