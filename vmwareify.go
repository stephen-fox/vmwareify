package vmwareify

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"unicode"

	"github.com/stephen-fox/vmwareify/ovf"
)

// BasicConvert converts a non-VMWare .ovf file to a VMWare friendly .ovf
// file. It will remove any IDE controllers and convert any existing
// SATA controllers to the VMWare kind. It will also set the VMWare
// compatibility level to vmx-10.
func BasicConvert(ovfFilePath string, newFilePath string) error {
	if ovfFilePath == newFilePath {
		return errors.New("Output .ovf file path cannot be the same as the input file path")
	}

	existing, err := os.Open(ovfFilePath)
	if err != nil {
		return err
	}
	defer existing.Close()

	editOptions := ovf.EditOptions{
		OnSystem: []ovf.OnSystemFunc{
			SetVirtualSystemTypeFunc("vmx-10"),
		},
		OnHardwareItems: []ovf.OnHardwareItemFunc{
			RemoveIdeControllersFunc(-1),
			ConvertSataControllersFunc(),
		},
	}

	buff, err := ovf.EditRawOvf(existing, editOptions)
	if err != nil {
		return err
	}

	info, err := existing.Stat()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(newFilePath, buff.Bytes(), info.Mode())
	if err != nil {
		return err
	}

	return nil
}

// SetVirtualSystemTypeFunc returns an ovf.OnSystemFunc that will set the
// .ovf's VirtualSystemType to the specified value.
//
// See ovf.OnSystemFunc for details.
func SetVirtualSystemTypeFunc(systemType string) ovf.OnSystemFunc {
	return ovf.SetVirtualSystemTypeFunc(systemType)
}

// RemoveIdeControllersFunc returns an ovf.OnHardwareItemFunc that will remove
// the specified number of IDE controllers.
//
// See ovf.OnHardwareItemFunc for details.
func RemoveIdeControllersFunc(limit int) ovf.OnHardwareItemFunc {
	return ovf.DeleteHardwareItemsMatchingFunc("ideController", limit)
}

// ConvertSataControllersFunc returns an ovf.OnHardwareItemFunc that
// will convert an existing SATA controller to a VMWare friendly
// SATA controller.
//
// See ovf.OnHardwareItemFunc for details.
func ConvertSataControllersFunc() ovf.OnHardwareItemFunc {
	modifyFunc := func(sataController ovf.Item) ovf.Item {
		sataController.Caption = "SATA Controller"
		sataController.Description = "SATAController"

		updatedElementNameBuffer := bytes.NewBuffer(nil)
		updatedElementNameBuffer.WriteString("SATAController")
		for i := range sataController.ElementName {
			char := rune(sataController.ElementName[i])
			if unicode.IsDigit(char) {
				updatedElementNameBuffer.WriteString(string(char))
			}
		}
		sataController.ElementName = updatedElementNameBuffer.String()

		sataController.ResourceSubType = "vmware.sata.ahci"

		return sataController
	}

	return ovf.ModifyHardwareItemsOfResourceTypeFunc(ovf.SataControllerResourceType, modifyFunc)
}
