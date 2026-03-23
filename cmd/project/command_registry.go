package main

import (
	"cmp"
	"slices"

	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type commandDependencies struct {
	creator *appcreate.Creator
}

type commandProvider interface {
	key() commandKey
	buildCommand(commandDependencies) *cobra.Command
	order() int
}

type baseCommandProvider struct {
	commandKey commandKey
	index      int
}

func (p baseCommandProvider) key() commandKey {
	return p.commandKey
}

func (p baseCommandProvider) order() int {
	return p.index
}

type commandKey string

const (
	commandKeyNew        commandKey = "new"
	commandKeyInit       commandKey = "init"
	commandKeyList       commandKey = "list"
	commandKeyInspect    commandKey = "inspect"
	commandKeyVersion    commandKey = "version"
	commandKeyCompletion commandKey = "completion"
)

var commandRegistry []commandProvider

func registerCommand(provider commandProvider) {
	for _, existing := range commandRegistry {
		if existing.key() == provider.key() {
			panic("duplicate command registration: " + string(provider.key()))
		}
	}

	commandRegistry = append(commandRegistry, provider)
}

func registerOrderedCommand(key commandKey, index int, build func(commandDependencies) *cobra.Command) {
	registerCommand(functionCommandProvider{
		baseCommandProvider: baseCommandProvider{commandKey: key, index: index},
		build:               build,
	})
}

func registeredCommandProviders() []commandProvider {
	providers := append([]commandProvider(nil), commandRegistry...)
	slices.SortStableFunc(providers, func(left, right commandProvider) int {
		return cmp.Compare(left.order(), right.order())
	})
	return providers
}

func registeredCommandProvider(key commandKey) (commandProvider, bool) {
	for _, provider := range registeredCommandProviders() {
		if provider.key() == key {
			return provider, true
		}
	}

	return nil, false
}

func buildRegisteredCommand(deps commandDependencies, key commandKey) (*cobra.Command, bool) {
	provider, ok := registeredCommandProvider(key)
	if !ok {
		return nil, false
	}

	return provider.buildCommand(deps), true
}

func addRegisteredCommands(root *cobra.Command, deps commandDependencies) {
	for _, provider := range registeredCommandProviders() {
		root.AddCommand(provider.buildCommand(deps))
	}
}

type functionCommandProvider struct {
	baseCommandProvider
	build func(commandDependencies) *cobra.Command
}

func (p functionCommandProvider) buildCommand(deps commandDependencies) *cobra.Command {
	return p.build(deps)
}

const (
	commandOrderNew = iota + 1
	commandOrderInit
	commandOrderList
	commandOrderInspect
	commandOrderVersion
	commandOrderCompletion
)
