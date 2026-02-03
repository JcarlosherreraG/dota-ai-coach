// Package dota contains types and a parser for Game State Integration (GSI) data from Dota 2.
package dota

import (
	"encoding/json"
	"fmt"
)

// Parser parses JSON data from Dota 2 GSI.
type Parser struct{}

// NewParser creates a new instance of the GSI data parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses JSON data from Dota 2 GSI into a GameState structure.
func (p *Parser) Parse(data []byte) (*GameState, error) {
	var state GameState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("could not parse game state: %w", err)
	}
	return &state, nil
}
