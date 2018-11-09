package ovf

import (
	"encoding/xml"
	"io"
	"io/ioutil"
)

// TODO: Hack for https://github.com/golang/go/issues/9519.
type xmlMarshableWorkAround interface {
	marshableFriendly() interface{}
}

type Ovf struct {
	Envelope Envelope
}

type Envelope struct {
	XMLName       xml.Name      `xml:"Envelope"`
	Version       string        `xml:"version,attr"`
	Lang          string        `xml:"lang,attr"`
	Xmlns         string        `xml:"xmlns,attr"`
	Ovf           string        `xml:"ovf,attr"`
	Rasd          string        `xml:"rasd,attr"`
	Vssd          string        `xml:"vssd,attr"`
	Xsi           string        `xml:"xsi,attr"`
	Vbox          string        `xml:"vbox,attr"`
	VirtualSystem VirtualSystem
}

type VirtualSystem struct {
	XMLName                xml.Name               `xml:"VirtualSystem"`
	Id                     string                 `xml:"id,attr"`
	VirtualHardwareSection VirtualHardwareSection
}

type VirtualHardwareSection struct {
	XMLName xml.Name `xml:"VirtualHardwareSection"`
	Info    string   `xml:"Info"`
	System  System
	Items   []Item   `xml:"Item"`
}

type System struct {
	XMLName                 xml.Name `xml:"System"`
	ElementName             string   `xml:"vssd:ElementName"`
	InstanceId              string   `xml:"vssd:InstanceID"`
	VirtualSystemIdentifier string   `xml:"vssd:VirtualSystemIdentifier"`
	VirtualSystemType       string   `xml:"vssd:VirtualSystemType"`
}

type Item struct {
	XMLName         xml.Name `xml:"Item"`
	Address         string   `xml:"Address"`
	AllocationUnits string   `xml:"AllocationUnits"`
	Caption         string   `xml:"Caption"`
	Description     string   `xml:"Description"`
	ElementName     string   `xml:"ElementName"`
	InstanceID      string   `xml:"InstanceID"`
	ResourceSubType string   `xml:"ResourceSubType"`
	ResourceType    string   `xml:"ResourceType"`
	VirtualQuantity string   `xml:"VirtualQuantity"`
}

// TODO: Hack for https://github.com/golang/go/issues/9519.
func (o *Item) marshableFriendly() interface{} {
	return marshableItem{
		Address:         o.Address,
		AllocationUnits: o.AllocationUnits,
		Caption:         o.Caption,
		Description:     o.Description,
		ElementName:     o.ElementName,
		InstanceID:      o.InstanceID,
		ResourceSubType: o.ResourceSubType,
		ResourceType:    o.ResourceType,
		VirtualQuantity: o.VirtualQuantity,
	}
}

type marshableItem struct {
	XMLName         xml.Name `xml:"Item"`
	Address         string   `xml:"rasd:Address"`
	AllocationUnits string   `xml:"rasd:AllocationUnits,omitempty"`
	Caption         string   `xml:"rasd:Caption"`
	Description     string   `xml:"rasd:Description"`
	ElementName     string   `xml:"rasd:ElementName"`
	InstanceID      string   `xml:"rasd:InstanceID"`
	ResourceSubType string   `xml:"rasd:ResourceSubType"`
	ResourceType    string   `xml:"rasd:ResourceType"`
	VirtualQuantity string   `xml:"rasd:VirtualQuantity,omitempty"`
}

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
