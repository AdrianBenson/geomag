package geomag

import (
	"testing"
)

var obs = `NZ_APIM_51_LFF, 2018-06-23 04:59:56.000, 3877896.0, 99.0
NZ_APIM_51_LFF, 2018-06-23 04:59:57.000, 3877895.0, 99.0
NZ_APIM_51_LFF, 2018-06-23 04:59:58.000, 3877900.0, 99.0
NZ_APIM_51_LFF, 2018-06-23 04:59:59.000, 3877896.0, 99.0
`

var name = "2018/2018.174/2018.174.0400.00.gsm-test.raw"

func TestGsm_Coding(t *testing.T) {

	var gsm Gsm

	err := gsm.Decode([]byte(obs))
	if err != nil {
		t.Fatalf("unable to decode gsm data: %v", err)
	}

	res, err := gsm.Encode()
	if err != nil {
		t.Fatalf("unable to encode gsm data: %v", err)
	}

	if string(obs) != string(res) {
		t.Errorf("gsm data mismatch:\n->\n%s<- != ->\n%s<-\n", string(obs), string(res))
	}

}

func TestGsm_Filename(t *testing.T) {

	var gsm Gsm

	err := gsm.Decode([]byte(obs))
	if err != nil {
		t.Fatalf("unable to decode gsm data: %v", err)
	}

	x := NewGsm("NZ_APIM_51_LFF", gsm.Readings)
	if len(x) != 1 {
		t.Fatalf("unable to extract gsm data")
	}

	if s := x[0].Filename("test"); s != name {
		t.Errorf("gsm name mismatch: \"%s\" != \"%s\"\n", name, s)
	}
}
