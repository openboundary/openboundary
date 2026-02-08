// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

import "fmt"

// HTTPServerSchema validates http.server component specs.
type HTTPServerSchema struct{}

// Kind returns the component kind.
func (s *HTTPServerSchema) Kind() Kind {
	return KindHTTPServer
}

// Validate validates the http.server spec.
func (s *HTTPServerSchema) Validate(spec map[string]interface{}) error {
	// TODO: Implement validation
	// Required fields: port
	// Optional fields: host, middleware, routes

	if _, ok := spec["port"]; !ok {
		return fmt.Errorf("http.server requires 'port' field")
	}

	return nil
}
