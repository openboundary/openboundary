// Copyright 2026 Open Boundary Contributors
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
	return fmt.Sprintf("%s (%d error(s))", e.Message, len(e.Errors))
}
