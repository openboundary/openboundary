// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser handles YAML parsing with position tracking.
type Parser struct {
	filename string
}

// NewParser creates a new Parser for the given file.
func NewParser(filename string) *Parser {
	return &Parser{filename: filename}
}

// Parse reads and parses the YAML specification file.
func (p *Parser) Parse() (*Spec, error) {
	data, err := os.ReadFile(p.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseBytes(data)
}

// ParseBytes parses YAML specification from bytes.
func (p *Parser) ParseBytes(data []byte) (*Spec, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	spec, err := p.parseSpec(&node)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

// parseSpec parses the root node into a Spec.
func (p *Parser) parseSpec(node *yaml.Node) (*Spec, error) {
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("expected document node")
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping at root")
	}

	spec := &Spec{
		position: WithPosition(p.filename, root.Line, root.Column),
	}

	// TODO: Implement full position-aware parsing
	// For now, use simple unmarshal
	if err := root.Decode(spec); err != nil {
		return nil, fmt.Errorf("failed to decode spec: %w", err)
	}

	return spec, nil
}
