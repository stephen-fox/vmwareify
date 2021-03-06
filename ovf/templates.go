package ovf

import (
	"strings"
)

// SetVirtualSystemTypeFunc returns an EditObjectFunc that sets the
// VirtualSystemType to the specified value.
func SetVirtualSystemTypeFunc(newVirtualSystemType string) EditObjectFunc {
	return func(i interface{}) EditObjectResult {
		o, ok := i.(System)
		if !ok {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		o.VirtualSystemType = newVirtualSystemType

		return EditObjectResult{
			Action: Replace,
			Object: &o,
		}
	}
}

// DeleteHardwareItemsMatchingFunc returns an EditObjectFunc that deletes
// an OVF Item whose element name matches the provided prefix. If the specified
// limit is less than 0, then the resulting function will have no limit.
func DeleteHardwareItemsMatchingFunc(elementNamePrefix string, limit int) EditObjectFunc {
	deleteFunc := deleteHardwareItemsMatchingFunc(elementNamePrefix)

	return func(i interface{}) EditObjectResult {
		o, ok := i.(Item)
		if !ok {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		if limit == 0 {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		result := deleteFunc(i)
		if result.Action == Delete {
			limit = limit - 1
		}

		return result
	}
}

func deleteHardwareItemsMatchingFunc(elementNamePrefix string) EditObjectFunc {
	return func(i interface{}) EditObjectResult {
		o, ok := i.(Item)
		if !ok {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		if strings.HasPrefix(o.ElementName, elementNamePrefix) {
			return EditObjectResult{
				Action: Delete,
				Object: &o,
			}
		}

		return EditObjectResult{
			Action: NoOp,
			Object: &o,
		}
	}
}

// ReplaceHardwareItemFunc returns an EditObjectFunc that replaces an OVF
// Item with a specific element name.
func ReplaceHardwareItemFunc(elementName string, replacement Item) EditObjectFunc {
	return func(i interface{}) EditObjectResult {
		o, ok := i.(Item)
		if !ok {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		if o.ElementName == elementName {
			return EditObjectResult{
				Action: Replace,
				Object: &replacement,
			}
		}

		return EditObjectResult{
			Action: NoOp,
			Object: &o,
		}
	}
}

// ModifyHardwareItemsOfResourceTypeFunc returns an EditObjectFunc that
// modifies OVF Item of a certain resource type.
func ModifyHardwareItemsOfResourceTypeFunc(resourceType string, modifyFunc func(i Item) Item) EditObjectFunc {
	return func(i interface{}) EditObjectResult {
		o, ok := i.(Item)
		if !ok {
			return EditObjectResult{
				Action: NoOp,
				Object: &o,
			}
		}

		if o.ResourceType == resourceType {
			newItem := modifyFunc(o)

			return EditObjectResult{
				Action: Replace,
				Object: &newItem,
			}
		}

		return EditObjectResult{
			Action: NoOp,
			Object: &o,
		}
	}
}
