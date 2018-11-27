package ovf

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

const (
	NoOp    EditAction = "no_op"
	Delete  EditAction = "delete"
	Replace EditAction = "replace"
)

type EditAction string

type EditOptions struct {
	OnSystem        []OnSystemFunc
	OnHardwareItems []OnHardwareItemFunc
}

type OnSystemFunc func(System) SystemResult

type SystemResult struct {
	EditAction EditAction
	NewSystem  System
}

type OnHardwareItemFunc func(Item) HardwareItemResult

type HardwareItemResult struct {
	EditAction EditAction
	NewItem    Item
}

type mangler struct {
	ioReader                io.Reader
	subtractCurrentOffsetBy int64
	result                  []byte
	// TODO: This is not ideal.
	original                []byte
}

func (o *mangler) editToken(decoder *xml.Decoder, options EditOptions) (bool, error) {
	token, couldRead, err := readNextToken(decoder)
	if err != nil {
		return false, err
	}

	if !couldRead {
		return false, nil
	}

	startElement, shouldEdit := shouldEditToken(token, options)
	if !shouldEdit {
		return true, nil
	}

	switch startElement.Name.Local {
	case systemFieldName:
		err := o.editSystem(decoder, &startElement, options.OnSystem)
		if err != nil {
			return false, err
		}
	case itemFieldName:
		err := o.editItem(decoder, &startElement, options.OnHardwareItems)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func readNextToken(decoder *xml.Decoder) (xml.Token, bool, error) {
	token, err := decoder.Token()
	if err != nil && err != io.EOF  {
		return nil, false, err
	}
	if token == nil {
		return nil, false, nil
	}

	return token, true, nil
}

func shouldEditToken(token xml.Token, options EditOptions) (xml.StartElement, bool) {
	startElement, ok := token.(xml.StartElement)
	if !ok {
		return xml.StartElement{}, false
	}

	if len(options.OnHardwareItems) > 0 && startElement.Name.Local == itemFieldName {
		return startElement, true
	}

	if len(options.OnSystem) > 0 && startElement.Name.Local == systemFieldName {
		return startElement, true
	}

	return xml.StartElement{}, false
}

func (o *mangler) editSystem(decoder *xml.Decoder, startElement *xml.StartElement, funcs []OnSystemFunc) error {
	startLine := o.originalLineInfo(decoder.InputOffset())

	var system System

	err := decoder.DecodeElement(&system, startElement)
	if err != nil {
		return err
	}

	for _, f := range funcs {
		result := f(system)
		switch result.EditAction {
		case NoOp:
			continue
		case Delete:
			// TODO: Deal with hardcoded deletion of new lines (offset + 1).
			o.deleteFrom(startLine.lineStartIndex, decoder.InputOffset() + 1)
		case Replace:
			endLine := o.originalLineInfo(decoder.InputOffset())
			raw, err := xml.MarshalIndent(result.NewSystem.marshableFriendly(),
				strings.Repeat(" ", endLine.numberOfSpaces), "  ")
			if err != nil {
				return err
			}

			o.replaceFrom(startLine.lineStartIndex, decoder.InputOffset(), raw)
		}
	}

	return nil
}

func (o *mangler) editItem(decoder *xml.Decoder, startElement *xml.StartElement, funcs []OnHardwareItemFunc) error {
	startLine := o.originalLineInfo(decoder.InputOffset())

	var item Item

	err := decoder.DecodeElement(&item, startElement)
	if err != nil {
		return err
	}

	for _, f := range funcs {
		result := f(item)
		switch result.EditAction {
		case NoOp:
			continue
		case Delete:
			// TODO: Deal with hardcoded deletion of new lines (offset + 1).
			o.deleteFrom(startLine.lineStartIndex, decoder.InputOffset() + 1)
		case Replace:
			endLine := o.originalLineInfo(decoder.InputOffset())
			raw, err := xml.MarshalIndent(result.NewItem.marshableFriendly(),
				strings.Repeat(" ", endLine.numberOfSpaces), "  ")
			if err != nil {
				return err
			}

			o.replaceFrom(startLine.lineStartIndex, decoder.InputOffset(), raw)
		}
	}

	return nil
}

func (o *mangler) Read(p []byte) (n int, err error) {
	n, err = o.ioReader.Read(p)
	if err != nil {
		return n, err
	}

	o.result = append(o.result, p[:n]...)
	o.original = append(o.original, p[:n]...)

	return n, err
}

func (o *mangler) originalLineInfo(decoderOffset int64) lineInfo {
	openTagIndex := bytes.LastIndex(o.original[:decoderOffset], []byte{'<'})
	if openTagIndex < 0 {
		return lineInfo{
			abort: true,
		}
	}

	previousElementEndIndex := bytes.LastIndex(o.original[:openTagIndex], []byte{'>'})
	if previousElementEndIndex < 0 {
		return lineInfo{
			abort: true,
		}
	}

	prevLineHasNewLine := bytes.Contains(o.original[previousElementEndIndex:openTagIndex], []byte{'\n'})

	numSpaces := bytes.Count(o.original[previousElementEndIndex:openTagIndex], []byte{' '})

	hasNewLine := false
	if decoderOffset < int64(len(o.original)) && o.original[decoderOffset] == '\n' {
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

func (o *mangler) deleteFrom(from int64, to int64) int64 {
	numCharsDeleted := to - from
	if numCharsDeleted <= 0 {
		return 0
	}

	from = from - o.subtractCurrentOffsetBy
	to = to - o.subtractCurrentOffsetBy

	o.subtractCurrentOffsetBy = o.subtractCurrentOffsetBy + numCharsDeleted

	o.result = append(o.result[:from], o.result[to:]...)

	return numCharsDeleted
}

func (o *mangler) replaceFrom(from int64, to int64, newRaw []byte) {
	// Need to capture the previous offset adjustment before deleting
	// the data.
	lastOffsetAdjustment := o.subtractCurrentOffsetBy

	deleted := o.deleteFrom(from, to)

	// Update the starting offset.
	from = from - lastOffsetAdjustment

	replacementLength := int64(len(newRaw))
	if replacementLength- deleted > 0 {
		// The replacement is larger than what was there.
		o.subtractCurrentOffsetBy = o.subtractCurrentOffsetBy + replacementLength
	} else {
		// The replacement is smaller than what was there.
		o.subtractCurrentOffsetBy = o.subtractCurrentOffsetBy - replacementLength
	}

	if from > int64(len(o.result)) {
		o.result = append(o.result, newRaw...)
		return
	}

	o.result = append(o.result[:from], append(newRaw, o.result[from:]...)...)
}

func (o *mangler) buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.result)
}

type lineInfo struct {
	abort              bool
	lineStartIndex     int64
	tagStartIndex      int64
	numberOfSpaces     int
	endsWithNewLine    bool
	prevLineHasNewLine bool
}

func newMangler(r io.Reader) *mangler {
	return &mangler{
		ioReader: r,
	}
}

func SetVirtualSystemTypeFunc(newVirtualSystemType string) OnSystemFunc {
	return func(s System) SystemResult {
		s.VirtualSystemType = newVirtualSystemType

		return SystemResult{
			EditAction: Replace,
			NewSystem:  s,
		}
	}
}

func DeleteHardwareItemsMatchingFunc(elementNamePrefix string, limit int) OnHardwareItemFunc {
	deleteFunc := deleteHardwareItemsMatchingFunc(elementNamePrefix)

	return func(i Item) HardwareItemResult {
		if limit == 0 {
			return HardwareItemResult{
				EditAction: NoOp,
			}
		}

		result := deleteFunc(i)
		if result.EditAction == Delete {
			limit = limit - 1
		}

		return result
	}
}

func deleteHardwareItemsMatchingFunc(elementNamePrefix string) OnHardwareItemFunc {
	return func(i Item) HardwareItemResult {
		if strings.HasPrefix(i.ElementName, elementNamePrefix) {
			return HardwareItemResult{
				EditAction: Delete,
			}
		}

		return HardwareItemResult{
			EditAction: NoOp,
		}
	}
}

func ReplaceHardwareItemFunc(elementName string, item Item) OnHardwareItemFunc {
	return func(i Item) HardwareItemResult {
		if i.ElementName == elementName {
			return HardwareItemResult{
				EditAction: Replace,
				NewItem:    item,
			}
		}

		return HardwareItemResult{
			EditAction: NoOp,
		}
	}
}

func EditRawOvf(r io.Reader, options EditOptions) (*bytes.Buffer, error) {
	mangler := newMangler(r)
	decoder := xml.NewDecoder(mangler)

	for {
		shouldContinue, err := mangler.editToken(decoder, options)
		if err != nil {
			return mangler.buffer(), err
		}

		if !shouldContinue {
			break
		}
	}

	return mangler.buffer(), nil
}
