// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package pipeline

import "fmt"

// StageError wraps multiple errors from a pipeline stage.
// The CLI layer can type-assert to format these errors to stderr.
type StageError struct {
	Stage   string
	Message string
	Errors  []error
}

func (e *StageError) Error() string {
	return fmt.Sprintf("stage %s: %s (%d error(s))", e.Stage, e.Message, len(e.Errors))
}
