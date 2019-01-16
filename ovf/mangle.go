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

// EditScheme specifies how an OVF configuration should be modified.
// There is no guarantee that the specified edits will be executed as the
// specified OVF object(s) may not be present in the file.
type EditScheme interface {
	// ShouldEditObject returns true and a non-empty slice of
	// EditObjectFunc if the specified OVF object has been
	// targeted for editing.
	ShouldEditObject(objectName ObjectName) ([]EditObjectFunc, bool)

	// Propose will execute the provided EditObjectFunc if it
	// encounters the specified ObjectName.
	Propose(EditObjectFunc, ObjectName) EditScheme
}

type defaultEditScheme struct {
	objectNamesToFuncs map[ObjectName][]EditObjectFunc
}

func (o *defaultEditScheme) ShouldEditObject(objectName ObjectName) ([]EditObjectFunc, bool) {
	fns, ok := o.objectNamesToFuncs[objectName]
	return fns, ok
}

func (o *defaultEditScheme) Propose(f EditObjectFunc, objectName ObjectName, ) EditScheme {
	o.objectNamesToFuncs[objectName] = append(o.objectNamesToFuncs[objectName], f)
	return o
}

// EditObjectFunc receives an OVF object and returns the resulting object
// as an EditObjectResult.
type EditObjectFunc func(originalObject interface{}) EditObjectResult

// EditObjectResult represents the result of editing an OVF object.
type EditObjectResult struct {
	Action EditAction
	Object EditedObject
}

// EditedObject represents an edited OVF object.
type EditedObject interface {
	// TODO: Hack for https://github.com/golang/go/issues/9519.
	//  This method returns a XML struct that contains fields tagged
	//  with the proper XML namespace.
	Marshallable() interface{}
}

var (
	crLfEol = []byte{'\r', '\n'}
	lfEol   = []byte{'\n'}
)

// EditRawOvf edits an existing OVF configuration in the form of an io.Reader
// given a set of EditScheme.
func EditRawOvf(r io.Reader, scheme EditScheme) (*bytes.Buffer, error) {
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
		err := processNextToken(scanner, endOfLineChars, newData, scheme)
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

func processNextToken(scanner *bufio.Scanner, eol []byte, newData *bytes.Buffer, scheme EditScheme) error {
	rawLine := scanner.Bytes()

	element, isStartElement := xmlutil.IsStartElement(rawLine)
	if isStartElement {
		var result []byte
		action := NoOp

		fns, shouldEdit := scheme.ShouldEditObject(ObjectName(element.Name.Local))
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
	case VirtualHardwareSystemName.String():
		t := System{}
		rawObject, err = xmlutil.FindAndDeserializeObject(findConfig, &t)
		temp.i = t
	case VirtualHardwareItemName.String():
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

// NewEditScheme returns a new instance of EditScheme.
func NewEditScheme() EditScheme {
	return &defaultEditScheme{
		objectNamesToFuncs: make(map[ObjectName][]EditObjectFunc),
	}
}
