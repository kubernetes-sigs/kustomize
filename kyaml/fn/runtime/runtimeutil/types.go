package runtimeutil

type DeferFailureFunction interface {
	GetExit() error
}
