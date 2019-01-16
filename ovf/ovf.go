package ovf

import (
	"encoding/xml"
	"io"
	"io/ioutil"
)

const (
	CdDriveResourceType            = "15"
	OtherStorageDeviceResourceType = "20"
)

const (
	VirtualHardwareSystemName ObjectName = "System"
	VirtualHardwareItemName   ObjectName = "Item"
)

// ObjectName represents an OVF object name.
type ObjectName string

func (o ObjectName) String() string {
	return string(o)
}

// Ovf is the parent that represents a single OVF configuration.
//
// TODO: Be advised: Not all fields are currently implemented.
//
// TODO: Be advised: Golang does not support XML namespaces when marshalling
//  (i.e., serializing) to XML. Please see the following GitHub issue:
//  https://github.com/golang/go/issues/9519.
type Ovf struct {
	Envelope Envelope
}

type Envelope struct {
	XMLName       xml.Name `xml:"Envelope"`
	Version       string   `xml:"version,attr"`
	Lang          string   `xml:"lang,attr"`
	Xmlns         string   `xml:"xmlns,attr"`
	Ovf           string   `xml:"ovf,attr"`
	Rasd          string   `xml:"rasd,attr"`
	Vssd          string   `xml:"vssd,attr"`
	Xsi           string   `xml:"xsi,attr"`
	Vbox          string   `xml:"vbox,attr"`
	VirtualSystem VirtualSystem
}

type VirtualSystem struct {
	XMLName                xml.Name `xml:"VirtualSystem"`
	Id                     string   `xml:"id,attr"`
	VirtualHardwareSection VirtualHardwareSection
}

type VirtualHardwareSection struct {
	XMLName xml.Name `xml:"VirtualHardwareSection"`
	Info    string   `xml:"Info"`
	System  System
	Items   []Item `xml:"Item"`
}

type System struct {
	XMLName                 xml.Name `xml:"System"`
	ElementName             string   `xml:"ElementName"`
	InstanceId              string   `xml:"InstanceID"`
	VirtualSystemIdentifier string   `xml:"VirtualSystemIdentifier"`
	VirtualSystemType       string   `xml:"VirtualSystemType"`
}

// TODO: Hack for https://github.com/golang/go/issues/9519.
func (o *System) Marshallable() interface{} {
	return marshableSystem{
		ElementName:             o.ElementName,
		InstanceId:              o.InstanceId,
		VirtualSystemIdentifier: o.VirtualSystemIdentifier,
		VirtualSystemType:       o.VirtualSystemType,
	}
}

// TODO: Hack for https://github.com/golang/go/issues/9519.
type marshableSystem struct {
	XMLName                 xml.Name `xml:"System"`
	ElementName             string   `xml:"vssd:ElementName"`
	InstanceId              string   `xml:"vssd:InstanceID"`
	VirtualSystemIdentifier string   `xml:"vssd:VirtualSystemIdentifier"`
	VirtualSystemType       string   `xml:"vssd:VirtualSystemType"`
}

type Item struct {
	XMLName             xml.Name `xml:"Item"`
	Address             string   `xml:"Address"`
	AddressOnParent     string   `xml:"AddressOnParent"`
	AllocationUnits     string   `xml:"AllocationUnits"`
	AutomaticAllocation bool     `xml:"AutomaticAllocation"`
	Caption             string   `xml:"Caption"`
	Description         string   `xml:"Description"`
	ElementName         string   `xml:"ElementName"`
	InstanceID          string   `xml:"InstanceID"`
	Parent              string   `xml:"Parent"`
	ResourceSubType     string   `xml:"ResourceSubType"`
	ResourceType        string   `xml:"ResourceType"`
	VirtualQuantity     string   `xml:"VirtualQuantity"`
}

// TODO: Hack for https://github.com/golang/go/issues/9519.
func (o *Item) Marshallable() interface{} {
	return marshableItem{
		Address:             o.Address,
		AddressOnParent:     o.AddressOnParent,
		AllocationUnits:     o.AllocationUnits,
		AutomaticAllocation: o.AutomaticAllocation,
		Caption:             o.Caption,
		Description:         o.Description,
		ElementName:         o.ElementName,
		InstanceID:          o.InstanceID,
		Parent:              o.Parent,
		ResourceSubType:     o.ResourceSubType,
		ResourceType:        o.ResourceType,
		VirtualQuantity:     o.VirtualQuantity,
	}
}

// TODO: Hack for https://github.com/golang/go/issues/9519.
type marshableItem struct {
	XMLName             xml.Name `xml:"Item"`
	Address             string   `xml:"rasd:Address,omitempty"`
	AddressOnParent     string   `xml:"rasd:AddressOnParent,omitempty"`
	AllocationUnits     string   `xml:"rasd:AllocationUnits,omitempty"`
	AutomaticAllocation bool     `xml:"rasd:AutomaticAllocation,omitempty"`
	Caption             string   `xml:"rasd:Caption"`
	Description         string   `xml:"rasd:Description"`
	ElementName         string   `xml:"rasd:ElementName"`
	InstanceID          string   `xml:"rasd:InstanceID"`
	Parent              string   `xml:"rasd:Parent,omitempty"`
	ResourceSubType     string   `xml:"rasd:ResourceSubType,omitempty"`
	ResourceType        string   `xml:"rasd:ResourceType"`
	VirtualQuantity     string   `xml:"rasd:VirtualQuantity,omitempty"`
}

// ToOvf produces an Ovf for the data provided by the io.Reader.
func ToOvf(r io.Reader) (Ovf, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return Ovf{}, err
	}

	var env Envelope

	err = xml.Unmarshal(raw, &env)
	if err != nil {
		return Ovf{}, err
	}

	return Ovf{
		Envelope: env,
	}, nil
}
