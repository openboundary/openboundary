// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

// MiddlewareSchema validates middleware component specs.
type MiddlewareSchema struct{}

// Kind returns the component kind.
func (s *MiddlewareSchema) Kind() Kind {
	return KindMiddleware
}

// Validate validates the middleware spec.
func (s *MiddlewareSchema) Validate(spec map[string]interface{}) error {
	// TODO: Implement validation
	// Required fields: type
	// Optional fields vary by middleware type
	return nil
}
