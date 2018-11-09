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
	OnHardwareItems []OnHardwareItemFunc
	DeleteLimit     int
}

type OnHardwareItemFunc func(Item) HardwareItemResult

type HardwareItemResult struct {
	EditAction EditAction
	NewItem    Item
}

type mangler struct {
	r               io.Reader
	numCharsDeleted int64
	result          []byte
}

func (o *mangler) editToken(decoder *xml.Decoder, options EditOptions) (bool, error) {
	token, err := decoder.Token()
	if err != nil && err != io.EOF  {
		return false, err
	}
	if token == nil {
		return false, nil
	}

	itemFieldName := "Item"

	switch tokenData := token.(type) {
	case xml.StartElement:
		switch tokenData.Name.Local {
		case itemFieldName:
			startLine := o.lineInfo(decoder.InputOffset())

			var item Item

			err := decoder.DecodeElement(&item, &tokenData)
			if err != nil {
				return false, err
			}

			for _, f := range options.OnHardwareItems {
				result := f(item)
				switch result.EditAction {
				case NoOp:
					continue
				case Delete:
					// TODO: Deal with hardcoded deletion of new lines (offset + 1).
					o.deleteFrom(startLine.lineStartIndex, decoder.InputOffset() + 1)
					break
				case Replace:
					endLine := o.lineInfo(decoder.InputOffset())
					raw, err := xml.MarshalIndent(result.NewItem.marshableFriendly(),
						strings.Repeat(" ", endLine.numberOfSpaces), "  ")
					if err != nil {
						return false, err
					}

					o.replace(startLine.lineStartIndex, decoder.InputOffset(), raw)
					break
				}
			}
		}
	}

	return true, nil
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

	o.result = append(o.result[:from], o.result[to:]...)
}

func (o *mangler) replace(from int64, to int64, raw []byte) {
	o.deleteFrom(from, to)

	length := int64(len(o.result))

	if from > length {
		o.result = append(o.result, raw...)
		return
	}

	o.result = append(o.result[:from], append(raw, o.result[from:]...)...)
}

func (o *mangler) buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.result)
}

func newMangler(r io.Reader) *mangler {
	return &mangler{
		r: r,
	}
}

func DeleteHardwareItemsMatchingFunc(elementNamePrefixes []string) OnHardwareItemFunc {
	return func(i Item) HardwareItemResult {
		for _, name := range elementNamePrefixes {
			if strings.HasPrefix(i.ElementName, name) {
				return HardwareItemResult{
					EditAction: Delete,
				}
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
