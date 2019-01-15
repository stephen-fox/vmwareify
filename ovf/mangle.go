package ovf

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"

	"github.com/stephen-fox/vmwareify/internal/xmlutil"
)

const (
	// NoOp means that the OVF object will not be modified in any way.
	NoOp    EditAction = "no_op"

	// Delete means that the OVF object will be deleted.
	Delete  EditAction = "delete"

	// Replace means that the OVF object will be replaced.
	Replace EditAction = "replace"
)

// EditAction describes what should happen when editing an OVF object.
type EditAction string

func (o EditAction) String() string {
	return string(o)
}

// EditOptions 
type EditOptions interface {
	//
	EditObjectNamed(objectName string, f EditObjectFunc)

	//
	ShouldEditObject(objectName string) ([]EditObjectFunc, bool)

	//
	EditVirtualHardwareItem(EditObjectFunc)

	//
	EditVirtualSystem(EditObjectFunc)
}

type defaultEditOptions struct {
	objectNamesToFuncs map[string][]EditObjectFunc
}

func (o *defaultEditOptions) EditObjectNamed(objectName string, f EditObjectFunc) {
	o.objectNamesToFuncs[objectName] = append(o.objectNamesToFuncs[objectName], f)
}

func (o *defaultEditOptions) ShouldEditObject(objectName string) ([]EditObjectFunc, bool) {
	fns, ok := o.objectNamesToFuncs[objectName]
	return fns, ok
}

func (o *defaultEditOptions) EditVirtualHardwareItem(f EditObjectFunc) {
	o.objectNamesToFuncs[itemFieldName] = append(o.objectNamesToFuncs[itemFieldName], f)
}

func (o *defaultEditOptions) EditVirtualSystem(f EditObjectFunc) {
	o.objectNamesToFuncs[systemFieldName] = append(o.objectNamesToFuncs[systemFieldName], f)
}

// EditObjectFunc
type EditObjectFunc func(originalObject interface{}) EditResult

// EditResult
type EditResult struct {
	Action EditAction
	Object EditedObject
}

// EditedObject
type EditedObject interface {
	// TODO: Hack for https://github.com/golang/go/issues/9519.
	Marshallable() interface{}
}

var (
	crLfEol = []byte{'\r', '\n'}
	lfEol   = []byte{'\n'}
)

// EditRawOvf edits an existing OVF configuration in the form of an io.Reader
// given a set of EditOptions.
func EditRawOvf(r io.Reader, options EditOptions) (*bytes.Buffer, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = xmlutil.ValidateFormatting(raw)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))

	endOfLineChars := lfEol
	lenRaw := len(raw)
	if lenRaw > 1 && raw[lenRaw-2] == '\r' {
		endOfLineChars = crLfEol
	}

	newData := bytes.NewBuffer(nil)

	for scanner.Scan() {
		err := processNextToken(scanner, endOfLineChars, newData, options)
		if err != nil {
			return newData, err
		}
	}

	err = scanner.Err()
	if err != nil {
		return newData, err
	}

	return newData, nil
}

func processNextToken(scanner *bufio.Scanner, eol []byte, newData *bytes.Buffer, options EditOptions) error {
	rawLine := scanner.Bytes()

	element, isStartElement := xmlutil.IsStartElement(rawLine)
	if isStartElement {
		var result []byte
		action := NoOp

		fns, shouldEdit := options.ShouldEditObject(element.Name.Local)
		if shouldEdit {
			findConfig, err := xmlutil.NewFindObjectConfig(element, scanner, eol)
			if err != nil {
				return err
			}

			result, action, err = edit(findConfig, fns)
			if err != nil {
				return err
			}
		}

		switch action {
		case NoOp:
			if len(result) > 0 {
				newData.Write(result)
			} else {
				newData.Write(rawLine)
			}
		case Delete:
			return nil
		case Replace:
			newData.Write(result)
		default:
			return errors.New("unknown EditAction - '" + action.String() + "")
		}

		newData.Write(eol)

		return nil
	}

	newData.Write(rawLine)

	newData.Write(eol)

	return nil
}

func edit(findConfig xmlutil.FindObjectConfig, funcs []EditObjectFunc) ([]byte, EditAction, error) {
	var rawObject xmlutil.RawObject
	var err error

	temp := struct {
		i interface{}
	}{}

	switch findConfig.Start().Name.Local {
	case systemFieldName:
		t := System{}
		rawObject, err = xmlutil.FindAndDeserializeObject(findConfig, &t)
		temp.i = t
	case itemFieldName:
		t := Item{}
		rawObject, err = xmlutil.FindAndDeserializeObject(findConfig, &t)
		temp.i = t
	default:
		return []byte{}, NoOp, errors.New("deserializing object '" +
			findConfig.Start().Name.Local + "' is not supported")
	}
	if err != nil {
		return []byte{}, NoOp, err
	}

	for _, f := range funcs {
		result := f(temp.i)
		switch result.Action {
		case NoOp:
			continue
		case Delete:
			return []byte{}, Delete, nil
		case Replace:
			raw, err := xml.MarshalIndent(result.Object.Marshallable(),
				rawObject.StartAndEndLinePrefix(), rawObject.RelativeBodyPrefix())
			if err != nil {
				return []byte{}, NoOp, err
			}

			return raw, Replace, nil
		}
	}

	return rawObject.Data().Bytes(), NoOp, nil
}

// NewEditOptions
func NewEditOptions() EditOptions {
	return &defaultEditOptions{
		objectNamesToFuncs: make(map[string][]EditObjectFunc),
	}
}
