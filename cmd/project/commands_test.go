package main

import "testing"

func TestSubcommands_HasExpectedKeys(t *testing.T) {
	commands := subcommands(newDependencies())
	got := make([]commandKey, 0, len(commands))
	for _, cmd := range commands {
		got = append(got, commandKey(cmd.Name()))
	}

	want := []commandKey{
		commandKeyNew,
		commandKeyInit,
		commandKeyList,
		commandKeyInspect,
		commandKeyConfig,
		commandKeyVersion,
		commandKeyCompletion,
	}

	if len(got) != len(want) {
		t.Fatalf("subcommand count = %d, want %d (got keys: %v)", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("subcommand keys = %v, want %v", got, want)
		}
	}
}

func TestSubcommands_AreParentless(t *testing.T) {
	for _, cmd := range subcommands(newDependencies()) {
		if cmd.HasParent() {
			t.Fatalf("subcommand %q already has a parent; newRootCmd owns attachment", cmd.Name())
		}
	}
}
