package main

import (
	"context"
	"testing"
)

type checkIpTest struct {
	ctx                     context.Context
	toCheck, ipStart, ipEnd string
	want                    bool
}

var checkIpTests = []checkIpTest{
	{context.TODO(), "10.80.21.32", "10.80.21.31/21", "10.80.21.51/21", true},
	{context.TODO(), "10.80.21.35", "10.80.21.31/21", "10.80.21.51/21", true},
	{context.TODO(), "25.82.21.32", "10.80.21.31/21", "10.80.21.51/21", false},
	{context.TODO(), "100.100.100.1000", "10.80.21.31/21", "10.80.21.51/21", false},
}

func TestCheckIP(t *testing.T) {
	n := new(Netbox)
	for _, test := range checkIpTests {
		if output := n.CheckIp(test.ctx, test.toCheck, test.ipStart, test.ipEnd); output != test.want {
			t.Errorf("output %v not equal to expected %v", test.toCheck, test.want)
		}
	}
}
