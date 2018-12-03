# vmwareify

## What is it?
A Go library and application for creating VMWare friendly OVF files.

## Use cases
This library makes it possible to create virtual appliances for VirtualBox and
VMWare a single OVF configuration.

For example, the [packer](https://packer.io/) automation tool can create
VirtualBox and VMWare virtual machine appliances. However, this requires
building the same VM twice; once for VirtualBox, and a second time for VMWare.
Rather than build the same machine twice, this library allows us to build the
VM once (for VirtualBox), and then convert the VirtualBox OVF into a VMWare
friendly copy.

Additionally, while certain VMWare tools are agnostic to the VirtualBox OVF
parameters, not all are. Most notably, vSphere will reject VirtualBox OVFs.
This library allows us to take an existing VirtualBox OVF, and use it with
the more restrictive tools like vSphere.

## API
The library provides several public functions for converting an OVF to a VMWare
friendly copy. The most notable is the `BasicConvert` function, which converts
an existing OVF into a VMWare friendly file. Here is an example application:
```go
package main

import (
    "log"

    "github.com/stephen-fox/vmwareify"
)

func main() {
    err := vmwareify.BasicConvert("/some.ovf", "/some-vmware.ovf")
    if err != nil {
        log.Fatal("Failed to convert .ovf file - " + err.Error())
    }
}
```

## Application usage
The included application can convert an existing OVF file into a VMWare
friendly one like so:
```bash
go run cmd/vmwareify/main.go -f /some.ovf
# Creates '/some-vmware.ovf'.
```

Optionally, you can specify the output file using `-o`:
```bash
go run cmd/vmwareify/main.go -f /some.ovf -o /my-awesome-vmware.ovf
```
