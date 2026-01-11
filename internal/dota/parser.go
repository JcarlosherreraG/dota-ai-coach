// Package dota содержит типы и парсер для Game State Integration (GSI) данных от Dota 2.
package dota

import (
	"encoding/json"
	"fmt"
)

// Parser парсит JSON данные от GSI Dota 2.
type Parser struct{}

// NewParser создаёт новый парсер GSI данных.
func NewParser() *Parser {
	return &Parser{}
}

// Parse парсит JSON данные от Dota 2 GSI в структуру GameState.
func (p *Parser) Parse(data []byte) (*GameState, error) {
	var state GameState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("could not parse game state: %w", err)
	}
	return &state, nil
}
