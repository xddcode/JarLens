package decompile

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJavaNotFound   = errors.New("java runtime was not found")
	ErrEngineFailed   = errors.New("decompiler engine failed")
	ErrOutputTooLarge = errors.New("decompiled source exceeds the 32 MiB limit")
)

type Request struct {
	ClassName   string
	Data        []byte
	ArchivePath string
}

type Result struct {
	Source        string
	Engine        string
	EngineVersion string
	Duration      time.Duration
	Diagnostics   []string
}

type Engine interface {
	Name() string
	Version() string
	Decompile(context.Context, Request) (Result, error)
}
