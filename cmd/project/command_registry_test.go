package main

import "testing"

func TestRegisteredCommandProviders_HasExpectedKeys(t *testing.T) {
	providers := registeredCommandProviders()
	got := make([]commandKey, 0, len(providers))
	for _, provider := range providers {
		got = append(got, provider.key())
	}

	want := []commandKey{
		commandKeyNew,
		commandKeyInit,
		commandKeyList,
		commandKeyInspect,
		commandKeyVersion,
		commandKeyCompletion,
	}

	if len(got) != len(want) {
		t.Fatalf("registered command count = %d, want %d (got keys: %v)", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("registered command keys = %v, want %v", got, want)
		}
	}
}
