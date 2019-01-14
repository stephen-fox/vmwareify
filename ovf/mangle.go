package ovf

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"strings"

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

// EditOptions contains the functions to execute when their respective objects
// are encountered when editing an OVF configuration.
type EditOptions struct {
	OnSystem        []OnSystemFunc
	OnHardwareItems []OnHardwareItemFunc
}

// OnSystemFunc is a function that will receive an OVF System. It must return
// a SystemResult, which will dictate what should happen to the System, along
// with the resulting System.
type OnSystemFunc func(System) SystemResult

type SystemResult struct {
	EditAction EditAction
	NewSystem  System
}

// OnHardwareItemFunc is a function that will receive an OVF Item. It must
// return a HardwareItemResult, which will dictate what should happen to the
// Item, along with the resulting Item.
type OnHardwareItemFunc func(Item) HardwareItemResult

type HardwareItemResult struct {
	EditAction EditAction
	NewItem    Item
}

// SetVirtualSystemTypeFunc returns an OnSystemFunc that sets the
// VirtualSystemType to the specified value.
func SetVirtualSystemTypeFunc(newVirtualSystemType string) OnSystemFunc {
	return func(s System) SystemResult {
		s.VirtualSystemType = newVirtualSystemType

		return SystemResult{
			EditAction: Replace,
			NewSystem:  s,
		}
	}
}

// DeleteHardwareItemsMatchingFunc returns an OnHardwareItemFunc that deletes
// an OVF Item whose element name matches the provided prefix. If the specified
// limit is less than 0, then the resulting function will have no limit.
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

// ReplaceHardwareItemFunc returns an OnHardwareItemFunc that replaces an OVF
// Item with a specific element name.
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

// ModifyHardwareItemsOfResourceTypeFunc returns an OnHardwareItemFunc that
// modifies OVF Item of a certain resource type.
func ModifyHardwareItemsOfResourceTypeFunc(resourceType string, modifyFunc func(i Item) Item) OnHardwareItemFunc {
	return func(i Item) HardwareItemResult {
		if i.ResourceType == resourceType {
			newItem := modifyFunc(i)

			return HardwareItemResult{
				EditAction: Replace,
				NewItem:    newItem,
			}
		}

		return HardwareItemResult{
			EditAction: NoOp,
		}
	}
}

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

	newData := bytes.NewBuffer(nil)

	for scanner.Scan() {
		err := processNextToken(scanner, newData, options)
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

func processNextToken(scanner *bufio.Scanner, newData *bytes.Buffer, options EditOptions) error {
	rawLine := scanner.Bytes()

	element, isStartElement := xmlutil.IsStartElement(rawLine)
	if isStartElement {
		var result []byte
		var err error
		action := NoOp

		switch element.Name.Local {
		case systemFieldName:
			if len(options.OnSystem) == 0 {
				break
			}

			var findConfig xmlutil.FindObjectConfig
			findConfig, err = xmlutil.NewFindObjectConfig(element, scanner, []byte{'\n'})
			if err != nil {
				return err
			}

			result, action, err = editSystem(findConfig, options)
		case itemFieldName:
			if len(options.OnHardwareItems) == 0 {
				break
			}

			var findConfig xmlutil.FindObjectConfig
			findConfig, err = xmlutil.NewFindObjectConfig(element, scanner, []byte{'\n'})
			if err != nil {
				return err
			}

			result, action, err = editItem(findConfig, options)
		}
		if err != nil {
			return err
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

		// TODO: Do not assume line ending.
		newData.Write([]byte{'\n'})

		return nil
	}

	newData.Write(rawLine)

	// TODO: Do not assume line ending.
	newData.Write([]byte{'\n'})

	return nil
}

// TODO: Replace typed 'edit*' functions with something more abstract.
func editSystem(findConfig xmlutil.FindObjectConfig, options EditOptions) ([]byte, EditAction, error) {
	var system System
	rawObject, err := xmlutil.FindAndDeserializeObject(findConfig, &system)
	if err != nil {
		return []byte{}, NoOp, err
	}

	for _, f := range options.OnSystem {
		result := f(system)
		switch result.EditAction {
		case NoOp:
			continue
		case Delete:
			return []byte{}, Delete, nil
		case Replace:
			raw, err := xml.MarshalIndent(result.NewSystem.marshableFriendly(),
				rawObject.StartAndEndLinePrefix(), rawObject.RelativeBodyPrefix())
			if err != nil {
				return []byte{}, NoOp, err
			}

			return raw, Replace, nil
		}
	}

	return rawObject.Data().Bytes(), NoOp, nil
}

// TODO: Replace typed 'edit*' functions with something more abstract.
func editItem(findConfig xmlutil.FindObjectConfig, options EditOptions) ([]byte, EditAction, error) {
	var item Item
	rawObject, err := xmlutil.FindAndDeserializeObject(findConfig, &item)
	if err != nil {
		return []byte{}, NoOp, err
	}

	for _, f := range options.OnHardwareItems {
		result := f(item)
		switch result.EditAction {
		case NoOp:
			continue
		case Delete:
			return []byte{}, Delete, nil
		case Replace:
			raw, err := xml.MarshalIndent(result.NewItem.marshableFriendly(),
				rawObject.StartAndEndLinePrefix(), rawObject.RelativeBodyPrefix())
			if err != nil {
				return []byte{}, NoOp, err
			}

			return raw, Replace, nil
		}
	}

	return rawObject.Data().Bytes(), NoOp, nil
}
