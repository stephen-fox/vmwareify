package ovf

import (
	"bytes"
	"encoding/xml"
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
	OnItemsWithElementNamePrefixes []OnItemsWithElementNamePrefixes
	ReplaceItemsWithElementNamePrefixes map[string]Item
	DeleteItemsWithElementNamePrefixes  []string
}

type OnItemsWithElementNamePrefixes func(Item) (Action, Item)

type mangler struct {
	deleted int64
	result  []byte
	r       io.Reader
}

func (o *mangler) Read(p []byte) (n int, err error) {
	n, err = o.r.Read(p)
	if err != nil {
		return n, err
	}

	o.result = append(o.result, p[:n]...)

	return n, err
}

func (o *mangler) deleteFrom(from int64, to int64) {
	toDelete := to - from
	if toDelete <= 0 {
		return
	}

	to = to - o.deleted
	from = from - o.deleted
	o.deleted = o.deleted + toDelete

	o.result = append(o.result[:from], o.result[to:]...)
}

func (o *mangler) replace(raw []byte, from int64, to int64) {
	// a = append(a[:i], append([]T{x}, a[i:]...)...)

}

func (o *mangler) buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.result)
}

func newMangler(r io.Reader) *mangler {
	return &mangler{
		r: r,
	}
}

func DeleteItemsWithElementPrefix(r io.Reader, names []string) (*bytes.Buffer, error) {
	f := func(i Item) (Action, Item) {
		for _, name := range names {
			if strings.HasPrefix(i.ElementName, name) {
				return Delete, Item{}
			}
		}

		return NoOp, Item{}
	}

	options := ManipulateOptions{
		OnItemsWithElementNamePrefixes: []OnItemsWithElementNamePrefixes{f},
	}

	return Manipulate(r, options)
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
				// TODO: Dynamically find and replace number of spaces.
				from := decoder.InputOffset() - int64(len(itemFieldName)) - numXmlOpenChars - 4

				var item Item

				err := decoder.DecodeElement(&item, &tokenData)
				if err != nil {
					return mangler.buffer(), err
				}

				for _, f := range options.OnItemsWithElementNamePrefixes {
					action, result := f(item)
					switch action {
					case NoOp:
						continue
					case Delete:
						to := decoder.InputOffset() + numXmlCloseChars
						mangler.deleteFrom(from, to)
						break
					case Replace:

					}
				}
			}
		}
	}

	return mangler.buffer(), nil
}
