package xmlutil

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

type testVhs struct {
	XMLName xml.Name  `xml:"VirtualHardwareSection"`
	Info   string     `xml:"Info"`
	System testSystem
}

type testSystem struct {
	XMLName xml.Name `xml:"System"`
	ElementName             string
	InstanceId              string
	VirtualSystemIdentifier string
	VirtualSystemType       string
}

var (
	testEol = []byte{'\n'}
)

func TestFindObject(t *testing.T) {
	junk := `<VirtualHardwareSection>
    <Info>Virtual hardware requirements for a virtual machine</Info>
    <System>
        <ElementName>Virtual Hardware Family</ElementName>
        <InstanceID>0</InstanceID>
        <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
        <VirtualSystemType>junk</VirtualSystemType>
    </System>
</VirtualHardwareSection>
`

	scanner := bufio.NewScanner(strings.NewReader(junk))

	for scanner.Scan() {
		line := scanner.Bytes()

		start, isStart := IsStartElement(line)
		if isStart && start.Name.Local == "System" {
			config, err := NewFindObjectConfig(start, scanner, testEol)
			if err != nil {
				t.Fatal(err.Error())
			}

			rawObject, err := FindObject(config)
			if err != nil {
				t.Fatal(err.Error())
			}

			expected := `    <System>
        <ElementName>Virtual Hardware Family</ElementName>
        <InstanceID>0</InstanceID>
        <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
        <VirtualSystemType>junk</VirtualSystemType>
    </System>`

			if rawObject.StartAndEndLinePrefix() != "    " {
				t.Fatal("Got unexpected start/end prefix of '" + rawObject.RelativeBodyPrefix() + "'")
			}

			if rawObject.BodyPrefix() != "        " {
				t.Fatal("Got unexpected body prefix of '" + rawObject.BodyPrefix() + "'")
			}

			if rawObject.RelativeBodyPrefix() != "    " {
				t.Fatal("Got unexpected relative body prefix of '" + rawObject.RelativeBodyPrefix() + "'")
			}

			if rawObject.Data().String() == expected {
				return
			} else {
      			t.Fatal("Got unexpected result: \n'" + rawObject.Data().String() + "'")
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Fatal("Could not find target object")
}

func TestFindObjectEmbeddedObject(t *testing.T) {
	junk := `<VirtualHardwareSection>
    <Info>Virtual hardware requirements for a virtual machine</Info>
    <System>
        <ElementName>Virtual Hardware Family</ElementName>
        <InstanceID>0</InstanceID>
        <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
        <VirtualSystemType>junk</VirtualSystemType>
        <System>
            <ElementName>Virtual Hardware Family</ElementName>
            <InstanceID>0</InstanceID>
            <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
            <VirtualSystemType>junk</VirtualSystemType>
        </System>
    </System>
</VirtualHardwareSection>
`

	scanner := bufio.NewScanner(strings.NewReader(junk))

	for scanner.Scan() {
		line := scanner.Bytes()

		start, isStart := IsStartElement(line)
		if isStart && start.Name.Local == "System" {
			config, err := NewFindObjectConfig(start, scanner, testEol)
			if err != nil {
				t.Fatal(err.Error())
			}

			rawObject, err := FindObject(config)
			if err != nil {
				t.Fatal(err.Error())
			}

			expected := `    <System>
        <ElementName>Virtual Hardware Family</ElementName>
        <InstanceID>0</InstanceID>
        <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
        <VirtualSystemType>junk</VirtualSystemType>
        <System>
            <ElementName>Virtual Hardware Family</ElementName>
            <InstanceID>0</InstanceID>
            <VirtualSystemIdentifier>centos7</VirtualSystemIdentifier>
            <VirtualSystemType>junk</VirtualSystemType>
        </System>
    </System>`

			if rawObject.StartAndEndLinePrefix() != "    " {
				t.Fatal("Got unexpected start/end prefix of '" + rawObject.RelativeBodyPrefix() + "'")
			}

			if rawObject.BodyPrefix() != "        " {
				t.Fatal("Got unexpected body prefix of '" + rawObject.BodyPrefix() + "'")
			}

			if rawObject.RelativeBodyPrefix() != "    " {
				t.Fatal("Got unexpected relative body prefix of '" + rawObject.RelativeBodyPrefix() + "'")
			}

			if rawObject.Data().String() == expected {
				fmt.Println(rawObject.Data().String())
				return
			} else {
				t.Fatal("Got unexpected result: \n'" + rawObject.Data().String() + "'")
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Fatal("Could not find target object")
}
