//nolint //cgo generates code that doesn't pass linting
package mseed

import (
	"io"
	"os"
	"testing"
)

func TestMSR_Repack(t *testing.T) {
	msr := NewMSRecord()
	defer FreeMSRecord(msr)

	record := make([]byte, 512)

	r, err := os.Open("etc/NZ.ABAZ.10.EHE.D.2016.079")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	_, err = io.ReadFull(r, record)
	if err != nil {
		t.Fatal(err)
	}

	err = msr.Unpack(record, 512, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	d, err := msr.DataSamples()
	if err != nil {
		t.Error(err)
	}

	var samples []int32
	for _, v := range d {
		samples = append(samples, v+100)
	}

	v, err := msr.Repack(samples, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	err = msr.Unpack(v, 512, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	d, err = msr.DataSamples()
	if err != nil {
		t.Error(err)
	}

	if len(d) != 397 {
		t.Errorf("expected 397 data samples got %d", len(d))
	}

	if d[0] != 327 {
		t.Errorf("expected first data value 327 got %d", d[0])
	}
}
