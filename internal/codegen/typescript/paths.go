// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"fmt"
	"strings"
)

// sanitizeFilename converts a component ID to a safe filename.
func sanitizeFilename(id string) string {
	// Replace dots and other special chars with dashes
	result := strings.ReplaceAll(id, ".", "-")
	result = strings.ReplaceAll(result, "/", "-")
	return result
}

func componentIDSlug(id string) string {
	return sanitizeFilename(id)
}

func serverSourcePath(id string) string {
	return fmt.Sprintf("src/components/%s.server.ts", componentIDSlug(id))
}

func serverContextPath(id string) string {
	return fmt.Sprintf("src/components/%s.context.ts", componentIDSlug(id))
}

func serverOpenAPIPath(id string) string {
	return fmt.Sprintf("src/components/%s.openapi.yaml", componentIDSlug(id))
}

func serverTestPath(id string) string {
	return fmt.Sprintf("src/components/%s.server.test.ts", componentIDSlug(id))
}

func middlewareSourcePath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.ts", componentIDSlug(id))
}

func middlewareConfigPath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.config.ts", componentIDSlug(id))
}

func middlewareSchemaPath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.schema.ts", componentIDSlug(id))
}

func middlewareModelPath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.model.conf", componentIDSlug(id))
}

func middlewarePolicyPath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.policy.csv", componentIDSlug(id))
}

func middlewareTestPath(id string) string {
	return fmt.Sprintf("src/components/%s.middleware.test.ts", componentIDSlug(id))
}

func postgresSourcePath(id string) string {
	return fmt.Sprintf("src/components/%s.postgres.ts", componentIDSlug(id))
}

func postgresSchemaPath(id string) string {
	return fmt.Sprintf("src/components/%s.postgres.schema.ts", componentIDSlug(id))
}

func postgresClientPath() string {
	return "src/components/postgres.client.ts"
}

func postgresClientImportPath() string {
	return "./postgres.client"
}

func usecaseSourcePath(id string) string {
	return fmt.Sprintf("src/components/%s.usecase.ts", componentIDSlug(id))
}

func usecaseTestPath(id string) string {
	return fmt.Sprintf("src/components/%s.usecase.test.ts", componentIDSlug(id))
}

func usecaseIndexPath() string {
	return "src/components/usecases.ts"
}

func usecaseSchemasPath() string {
	return "src/components/usecase.schemas.ts"
}
