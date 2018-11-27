package vmwareify

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/stephen-fox/vmwareify/ovf"
)

const (
	virtualBoxPrimarySataController = "sataController0"
)

func BasicConvert(ovfFilePath string, newFilePath string) error {
	if ovfFilePath == newFilePath {
		return errors.New("Output .ovf file path cannot be the same as the input file path")
	}

	existing, err := os.Open(ovfFilePath)
	if err != nil {
		return err
	}
	defer existing.Close()

	ovfData, err := ovf.ToOvf(existing)
	if err != nil {
		return err
	}

	_, err = existing.Seek(0, 0)
	if err != nil {
		return err
	}

	editOptions := ovf.EditOptions{
		OnSystem: []ovf.OnSystemFunc{
			SetVirtualSystemTypeFunc("vmx-10"),
		},
		OnHardwareItems: []ovf.OnHardwareItemFunc{
			RemoveIdeControllersFunc(-1),
		},
	}

	for _, item := range ovfData.Envelope.VirtualSystem.VirtualHardwareSection.Items {
		if item.ElementName == virtualBoxPrimarySataController {
			editOptions.OnHardwareItems = append(editOptions.OnHardwareItems, ConvertPrimarySataControllerFunc(item))
			break
		}
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

func SetVirtualSystemTypeFunc(systemType string) ovf.OnSystemFunc {
	return ovf.SetVirtualSystemTypeFunc(systemType)
}

func RemoveIdeControllersFunc(limit int) ovf.OnHardwareItemFunc {
	return ovf.DeleteHardwareItemsMatchingFunc("ideController", limit)
}

func ConvertPrimarySataControllerFunc(existingController ovf.Item) ovf.OnHardwareItemFunc {
	editedController := existingController

	editedController.Caption = "SATA Controller"
	editedController.Description = "SATAController"
	editedController.ElementName = "SATAController0"
	editedController.ResourceSubType = "vmware.sata.ahci"

	return ovf.ReplaceHardwareItemFunc(virtualBoxPrimarySataController, editedController)
}
