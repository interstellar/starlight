// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GOLICENSE file.

package xlm

import (
	"testing"
)

func TestDurationString(t *testing.T) {
	var cases = []struct {
		str string
		n   Amount
	}{
		{"0 XLM", 0},
		{"1 stroop", 1 * Stroop},
		{"2 stroops", 2 * Stroop},
		{"110 µXLM", 1100 * Stroop},
		{"110.1 µXLM", 1101 * Stroop},
		{"2.2 mXLM", 2200 * Microlumen},
		{"2.02 mXLM", 2020 * Microlumen},
		{"2.2001 mXLM", 2200*Microlumen + 1*Stroop},
		{"3.3 XLM", 3300 * Millilumen},
		{"3.03 XLM", 3030 * Millilumen},
		{"3.3000001 XLM", 3300*Millilumen + 1*Stroop},
		{"245 XLM", 245 * Lumen},
		{"245.001 XLM", 240*Lumen + 5001*Millilumen},
		{"18367.001 XLM", 18360*Lumen + 7001*Millilumen},
		{"480.0000001 XLM", 480*Lumen + 1*Stroop},
		{"922337203685.4775807 XLM", 1<<63 - 1},
		{"-922337203685.4775808 XLM", -1 << 63},
	}

	for _, tt := range cases {
		if str := tt.n.String(); str != tt.str {
			t.Errorf("Amount(%d).String() = %s, want %s", int64(tt.n), str, tt.str)
		}
		if tt.n > 0 {
			if str := (-tt.n).String(); str != "-"+tt.str {
				t.Errorf("Amount(%d).String() = %s, want %s", int64(-tt.n), str, "-"+tt.str)
			}
		}
	}
}

func TestHorizonString(t *testing.T) {
	var cases = []struct {
		str string
		n   Amount
	}{
		{"0", 0},
		{"8.7012201", 1*Lumen + 7701*Millilumen + 220*Microlumen + 1*Stroop},
		{"0.000724", 724 * Microlumen},
		{"0.0000001", 1 * Stroop},
		{"2500.1234567", 2500*Lumen + 123*Millilumen + 456*Microlumen + 7*Stroop},
		{"-250", -250 * Lumen},
		{"-2500.1234567", -1 * (2500*Lumen + 123*Millilumen + 456*Microlumen + 7*Stroop)},
		{"-0.0000001", -1 * Stroop},
	}

	for _, tt := range cases {
		if str := tt.n.HorizonString(); str != tt.str {
			t.Errorf("Amount(%d).HorizonString() = %s, want %s", int64(tt.n), str, tt.str)
		}
	}
}

func TestParse(t *testing.T) {
	var cases = []struct {
		str     string
		n       Amount
		wantErr error
	}{
		{"0", 0, nil},
		{"2002", 2002 * Lumen, nil},
		{"8.7012201", 1*Lumen + 7701*Millilumen + 220*Microlumen + 1*Stroop, nil},
		{"0.000724", 724 * Microlumen, nil},
		{"0.0000001", 1 * Stroop, nil},
		{"2500.1234567", 2500*Lumen + 123*Millilumen + 456*Microlumen + 7*Stroop, nil},
		{"", 0, errInvalidHorizonStr},
		{"127.0.0.1", 0, errInvalidHorizonStr},
		{"1.234567890", 0, errInvalidHorizonStr},
		{"1A", 0, errInvalidHorizonStr},
		{"11.2A", 0, errInvalidHorizonStr},
		{"11.-1", 0, errInvalidHorizonStr},
		{"-1", -1 * Lumen, nil},
		{"-0.0000001", -1 * Stroop, nil},
		{"-1000.701", -1 * (1000*Lumen + 701*Millilumen), nil},
	}

	for _, tt := range cases {
		n, err := Parse(tt.str)
		if err != tt.wantErr {
			t.Errorf("got %s, want %s", err, tt.wantErr)
		}
		if n != tt.n {
			t.Errorf("HorizonAmount(%s) = %d, want %d", tt.str, int64(n), int64(tt.n))
		}
	}
}
