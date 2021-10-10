package time

import (
	"testing"
	"time"
)

func TestDurationText(t *testing.T) {
	var inputs = make([][]byte, 3)
	inputs[0] = []byte("1s")
	inputs[1] = []byte("1m")
	inputs[2] = []byte("1h")

	var outputs = make([]time.Duration, 3)
	outputs[0] = time.Second
	outputs[1] = time.Minute
	outputs[2] = time.Hour

	var d Duration
	for i := 0; i< len(inputs); i++ {
		input := inputs[i]
		output := outputs[i]
		if err := d.UnmarshalText(input); err != nil {
			t.FailNow()
		}
		if int64(output) != int64(d) {
			t.FailNow()
		}
	}
}