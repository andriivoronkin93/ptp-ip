package main

import (
	"fmt"
	"github.com/malc0mn/ptp-ip/ip"
	"testing"
)

func TestCommandByName(t *testing.T) {
	cmds := map[string]command{
		"info": info,
		"getval": getval,
		"opreq": opreq,
		"state": state,
	}
	for name, want := range cmds {
		got := commandByName(name)
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("commandByName(%s) got = %v; want %v", name, got, want)
		}
	}
}

func TestUnknown(t *testing.T) {
	got := unknown(&ip.Client{}, []string{})
	want := "unknown command"
	if got != want {
		t.Errorf("got = '%s'; want '%s'", got, want)
	}
}