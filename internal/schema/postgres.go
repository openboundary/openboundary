// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

// PostgresSchema validates postgres component specs.
type PostgresSchema struct{}

// Kind returns the component kind.
func (s *PostgresSchema) Kind() Kind {
	return KindPostgres
}

// Validate validates the postgres spec.
func (s *PostgresSchema) Validate(spec map[string]interface{}) error {
	// TODO: Implement validation
	// Required fields: connection or connectionRef
	// Optional fields: migrations, models
	return nil
}
