package vmwareify

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"unicode"

	"github.com/stephen-fox/vmwareify/ovf"
)

// BasicConvert converts a non-VMWare .ovf file to a VMWare friendly .ovf
// file. It does the following:
//
//  - Removes any IDE controllers
//  - Converts any existing SATA controllers to the VMWare kind
//  - Set the VMWare compatibility level to vmx-10
//  - Disables automatic allocation of CD/DVD drives
func BasicConvert(ovfFilePath string, newFilePath string) error {
	if ovfFilePath == newFilePath {
		return errors.New("output .ovf file path cannot be the same as the input file path")
	}

	existing, err := os.Open(ovfFilePath)
	if err != nil {
		return err
	}
	defer existing.Close()

	buff, err := basicConvert(existing)
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

func basicConvert(existing io.Reader) (*bytes.Buffer, error) {
	editScheme := ovf.NewEditScheme().
		Propose(SetVirtualSystemTypeFunc("vmx-10"), ovf.VirtualHardwareSystemName).
		Propose(RemoveIdeControllersFunc(-1), ovf.VirtualHardwareItemName).
		Propose(ConvertSataControllersFunc(), ovf.VirtualHardwareItemName).
		Propose(DisableCdromAutomaticAllocationFunc(), ovf.VirtualHardwareItemName)

	buff, err := ovf.EditRawOvf(existing, editScheme)
	if err != nil {
		return bytes.NewBuffer(nil), err
	}

	return buff, nil
}

// SetVirtualSystemTypeFunc returns an ovf.EditObjectFunc that will set the
// .ovf's VirtualSystemType to the specified value.
func SetVirtualSystemTypeFunc(systemType string) ovf.EditObjectFunc {
	return ovf.SetVirtualSystemTypeFunc(systemType)
}

// RemoveIdeControllersFunc returns an ovf.EditObjectFunc that will remove
// the specified number of IDE controllers.
func RemoveIdeControllersFunc(limit int) ovf.EditObjectFunc {
	return ovf.DeleteHardwareItemsMatchingFunc("ideController", limit)
}

// ConvertSataControllersFunc returns an ovf.EditObjectFunc that
// will convert an existing SATA controller to a VMWare friendly
// SATA controller.
func ConvertSataControllersFunc() ovf.EditObjectFunc {
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

	return ovf.ModifyHardwareItemsOfResourceTypeFunc(ovf.OtherStorageDeviceResourceType, modifyFunc)
}

// DisableCdromAutomaticAllocationFunc returns an ovf.EditObjectFunc that
// will disable AutomaticAllocation for OVF ResourceType 15 devices.
func DisableCdromAutomaticAllocationFunc() ovf.EditObjectFunc {
	modifyFunc := func(cdrom ovf.Item) ovf.Item {
		cdrom.AutomaticAllocation = false
		return cdrom
	}

	return ovf.ModifyHardwareItemsOfResourceTypeFunc(ovf.CdDriveResourceType, modifyFunc)
}
