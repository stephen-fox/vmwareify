package vmwareify

import (
	"strings"
	"testing"
)

const (
	basicOvfFileContents = `<?xml version="1.0"?>
<Envelope ovf:version="1.0" xml:lang="en-US" xmlns="http://schemas.dmtf.org/ovf/envelope/1" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:vbox="http://www.virtualbox.org/ovf/machine">
  <References>
    <File ovf:id="file1" ovf:href="centos-0.0.1-disk001.vmdk"/>
  </References>
  <DiskSection>
    <Info>List of the virtual disks used in the package</Info>
    <Disk ovf:capacity="104857600000" ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" vbox:uuid="b3595d90-ffe1-4afb-a341-54b7a46d26e7"/>
  </DiskSection>
  <NetworkSection>
    <Info>Logical networks used in the package</Info>
    <Network ovf:name="NAT">
      <Description>Logical network used by this appliance.</Description>
    </Network>
  </NetworkSection>
  <VirtualSystem ovf:id="centos-0.0.1">
    <Info>A virtual machine</Info>
    <OperatingSystemSection ovf:id="80">
      <Info>The kind of installed guest operating system</Info>
      <Description>RedHat_64</Description>
      <vbox:OSType ovf:required="false">RedHat_64</vbox:OSType>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware requirements for a virtual machine</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemIdentifier>centos-0.0.1</vssd:VirtualSystemIdentifier>
        <vssd:VirtualSystemType>virtualbox-2.2</vssd:VirtualSystemType>
      </System>
      <Item>
        <rasd:Caption>1 virtual CPU</rasd:Caption>
        <rasd:Description>Number of virtual CPUs</rasd:Description>
        <rasd:ElementName>1 virtual CPU</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>1</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:AllocationUnits>MegaBytes</rasd:AllocationUnits>
        <rasd:Caption>512 MB of memory</rasd:Caption>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>512 MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>512</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Caption>ideController0</rasd:Caption>
        <rasd:Description>IDE Controller</rasd:Description>
        <rasd:ElementName>ideController0</rasd:ElementName>
        <rasd:InstanceID>3</rasd:InstanceID>
        <rasd:ResourceSubType>PIIX4</rasd:ResourceSubType>
        <rasd:ResourceType>5</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:Address>1</rasd:Address>
        <rasd:Caption>ideController1</rasd:Caption>
        <rasd:Description>IDE Controller</rasd:Description>
        <rasd:ElementName>ideController1</rasd:ElementName>
        <rasd:InstanceID>4</rasd:InstanceID>
        <rasd:ResourceSubType>PIIX4</rasd:ResourceSubType>
        <rasd:ResourceType>5</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Caption>sataController0</rasd:Caption>
        <rasd:Description>SATA Controller</rasd:Description>
        <rasd:ElementName>sataController0</rasd:ElementName>
        <rasd:InstanceID>5</rasd:InstanceID>
        <rasd:ResourceSubType>AHCI</rasd:ResourceSubType>
        <rasd:ResourceType>20</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:Caption>disk1</rasd:Caption>
        <rasd:Description>Disk Image</rasd:Description>
        <rasd:ElementName>disk1</rasd:ElementName>
        <rasd:HostResource>/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>6</rasd:InstanceID>
        <rasd:Parent>5</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>1</rasd:AddressOnParent>
        <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
        <rasd:Caption>cdrom1</rasd:Caption>
        <rasd:Description>CD-ROM Drive</rasd:Description>
        <rasd:ElementName>cdrom1</rasd:ElementName>
        <rasd:InstanceID>7</rasd:InstanceID>
        <rasd:Parent>5</rasd:Parent>
        <rasd:ResourceType>15</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
        <rasd:Caption>Ethernet adapter on 'NAT'</rasd:Caption>
        <rasd:Connection>NAT</rasd:Connection>
        <rasd:ElementName>Ethernet adapter on 'NAT'</rasd:ElementName>
        <rasd:InstanceID>8</rasd:InstanceID>
        <rasd:ResourceType>10</rasd:ResourceType>
      </Item>
    </VirtualHardwareSection>
    <vbox:Machine ovf:required="false" version="1.16-macosx" uuid="{aaf6485a-eba1-4105-b903-68f9d4ed35fc}" name="centos-0.0.1" OSType="RedHat_64" snapshotFolder="Snapshots" lastStateChange="2019-01-10T16:25:32Z">
      <ovf:Info>Complete VirtualBox machine configuration in VirtualBox format</ovf:Info>
      <Hardware>
        <CPU>
          <PAE enabled="true"/>
          <LongMode enabled="true"/>
          <X2APIC enabled="true"/>
          <HardwareVirtExLargePages enabled="true"/>
        </CPU>
        <Memory RAMSize="512"/>
        <Boot>
          <Order position="1" device="HardDisk"/>
          <Order position="2" device="DVD"/>
          <Order position="3" device="None"/>
          <Order position="4" device="None"/>
        </Boot>
        <RemoteDisplay enabled="true">
          <VRDEProperties>
            <Property name="TCP/Address" value="127.0.0.1"/>
            <Property name="TCP/Ports" value="5938"/>
          </VRDEProperties>
        </RemoteDisplay>
        <BIOS>
          <IOAPIC enabled="true"/>
        </BIOS>
        <Network>
          <Adapter slot="0" enabled="true" MACAddress="08002718A8F8" type="virtio">
            <NAT/>
          </Adapter>
        </Network>
        <AudioAdapter driver="CoreAudio" enabledIn="false" enabledOut="false"/>
      </Hardware>
      <StorageControllers>
        <StorageController name="IDE Controller" type="PIIX4" PortCount="2" useHostIOCache="true" Bootable="true"/>
        <StorageController name="SATA Controller" type="AHCI" PortCount="2" useHostIOCache="false" Bootable="true" IDE0MasterEmulationPort="0" IDE0SlaveEmulationPort="1" IDE1MasterEmulationPort="2" IDE1SlaveEmulationPort="3">
          <AttachedDevice type="HardDisk" hotpluggable="false" port="0" device="0">
            <Image uuid="{b3595d90-ffe1-4afb-a341-54b7a46d26e7}"/>
          </AttachedDevice>
          <AttachedDevice passthrough="false" type="DVD" hotpluggable="false" port="1" device="0"/>
        </StorageController>
      </StorageControllers>
    </vbox:Machine>
  </VirtualSystem>
</Envelope>
`
)

func TestBasicConvert(t *testing.T) {
	b, err := basicConvert(strings.NewReader(basicOvfFileContents))
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := `<?xml version="1.0"?>
<Envelope ovf:version="1.0" xml:lang="en-US" xmlns="http://schemas.dmtf.org/ovf/envelope/1" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:vbox="http://www.virtualbox.org/ovf/machine">
  <References>
    <File ovf:id="file1" ovf:href="centos-0.0.1-disk001.vmdk"/>
  </References>
  <DiskSection>
    <Info>List of the virtual disks used in the package</Info>
    <Disk ovf:capacity="104857600000" ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" vbox:uuid="b3595d90-ffe1-4afb-a341-54b7a46d26e7"/>
  </DiskSection>
  <NetworkSection>
    <Info>Logical networks used in the package</Info>
    <Network ovf:name="NAT">
      <Description>Logical network used by this appliance.</Description>
    </Network>
  </NetworkSection>
  <VirtualSystem ovf:id="centos-0.0.1">
    <Info>A virtual machine</Info>
    <OperatingSystemSection ovf:id="80">
      <Info>The kind of installed guest operating system</Info>
      <Description>RedHat_64</Description>
      <vbox:OSType ovf:required="false">RedHat_64</vbox:OSType>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware requirements for a virtual machine</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemIdentifier>centos-0.0.1</vssd:VirtualSystemIdentifier>
        <vssd:VirtualSystemType>vmx-10</vssd:VirtualSystemType>
      </System>
      <Item>
        <rasd:Caption>1 virtual CPU</rasd:Caption>
        <rasd:Description>Number of virtual CPUs</rasd:Description>
        <rasd:ElementName>1 virtual CPU</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>1</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:AllocationUnits>MegaBytes</rasd:AllocationUnits>
        <rasd:Caption>512 MB of memory</rasd:Caption>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>512 MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>512</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Caption>SATA Controller</rasd:Caption>
        <rasd:Description>SATAController</rasd:Description>
        <rasd:ElementName>SATAController0</rasd:ElementName>
        <rasd:InstanceID>5</rasd:InstanceID>
        <rasd:ResourceSubType>vmware.sata.ahci</rasd:ResourceSubType>
        <rasd:ResourceType>20</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:Caption>disk1</rasd:Caption>
        <rasd:Description>Disk Image</rasd:Description>
        <rasd:ElementName>disk1</rasd:ElementName>
        <rasd:HostResource>/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>6</rasd:InstanceID>
        <rasd:Parent>5</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>1</rasd:AddressOnParent>
        <rasd:Caption>cdrom1</rasd:Caption>
        <rasd:Description>CD-ROM Drive</rasd:Description>
        <rasd:ElementName>cdrom1</rasd:ElementName>
        <rasd:InstanceID>7</rasd:InstanceID>
        <rasd:Parent>5</rasd:Parent>
        <rasd:ResourceType>15</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
        <rasd:Caption>Ethernet adapter on 'NAT'</rasd:Caption>
        <rasd:Connection>NAT</rasd:Connection>
        <rasd:ElementName>Ethernet adapter on 'NAT'</rasd:ElementName>
        <rasd:InstanceID>8</rasd:InstanceID>
        <rasd:ResourceType>10</rasd:ResourceType>
      </Item>
    </VirtualHardwareSection>
    <vbox:Machine ovf:required="false" version="1.16-macosx" uuid="{aaf6485a-eba1-4105-b903-68f9d4ed35fc}" name="centos-0.0.1" OSType="RedHat_64" snapshotFolder="Snapshots" lastStateChange="2019-01-10T16:25:32Z">
      <ovf:Info>Complete VirtualBox machine configuration in VirtualBox format</ovf:Info>
      <Hardware>
        <CPU>
          <PAE enabled="true"/>
          <LongMode enabled="true"/>
          <X2APIC enabled="true"/>
          <HardwareVirtExLargePages enabled="true"/>
        </CPU>
        <Memory RAMSize="512"/>
        <Boot>
          <Order position="1" device="HardDisk"/>
          <Order position="2" device="DVD"/>
          <Order position="3" device="None"/>
          <Order position="4" device="None"/>
        </Boot>
        <RemoteDisplay enabled="true">
          <VRDEProperties>
            <Property name="TCP/Address" value="127.0.0.1"/>
            <Property name="TCP/Ports" value="5938"/>
          </VRDEProperties>
        </RemoteDisplay>
        <BIOS>
          <IOAPIC enabled="true"/>
        </BIOS>
        <Network>
          <Adapter slot="0" enabled="true" MACAddress="08002718A8F8" type="virtio">
            <NAT/>
          </Adapter>
        </Network>
        <AudioAdapter driver="CoreAudio" enabledIn="false" enabledOut="false"/>
      </Hardware>
      <StorageControllers>
        <StorageController name="IDE Controller" type="PIIX4" PortCount="2" useHostIOCache="true" Bootable="true"/>
        <StorageController name="SATA Controller" type="AHCI" PortCount="2" useHostIOCache="false" Bootable="true" IDE0MasterEmulationPort="0" IDE0SlaveEmulationPort="1" IDE1MasterEmulationPort="2" IDE1SlaveEmulationPort="3">
          <AttachedDevice type="HardDisk" hotpluggable="false" port="0" device="0">
            <Image uuid="{b3595d90-ffe1-4afb-a341-54b7a46d26e7}"/>
          </AttachedDevice>
          <AttachedDevice passthrough="false" type="DVD" hotpluggable="false" port="1" device="0"/>
        </StorageController>
      </StorageControllers>
    </vbox:Machine>
  </VirtualSystem>
</Envelope>
`

	result := b.String()
	if result != expected {
		t.Fatal("Did not get expected result:\n'" + result + "'")
	}
}