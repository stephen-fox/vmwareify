package ovf

import (
	"fmt"
	"os"
	"path"
	"testing"
)

const (
	testFilename = "virtualbox-vm.ovf"
)

func TestToOvf(t *testing.T) {
	f, err := os.Open(testDataFilePath(testFilename))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer f.Close()

	r, err := ToOvf(f)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(r.Envelope)

	// TODO: Test field values.
}

func testDataFilePath(filename string) string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return path.Join(path.Dir(p), ".testdata", "ovf", filename)
}
