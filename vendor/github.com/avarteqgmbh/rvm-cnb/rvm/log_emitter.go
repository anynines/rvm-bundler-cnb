package rvm

import (
	"io"

	"github.com/paketo-buildpacks/packit/scribe"
)

// LogEmitter allows to write logs using the packing lib
type LogEmitter struct {
	scribe.Logger
}

// NewLogEmitter returns a new LogEmitter instance
func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Logger: scribe.NewLogger(output),
	}
}
