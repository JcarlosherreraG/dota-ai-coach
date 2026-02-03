package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAndEnrichStratz(t *testing.T) {
	mockResponse := `
	{
		"data": {
			"constants": {
				"heroes": [
					{
						"id": 1,
						"displayName": "TestHero",
						"stats": {"strengthBase": 20},
						"facets": [ 
							{"slot": 0, "facetId": 300}, 
							{"slot": 1, "facetId": 301} 
						],
						"talents": [
							{"slot": 0, "abilityId": 100}
						],
						"abilities": [
							{"abilityId": 200}
						]
					}
				],
				"abilities": [
					{
						"id": 100,
						"isTalent": true,
						"language": {"displayName": "Talent +10 Damage"}
					},
					{
						"id": 200,
						"isTalent": false,
						"language": {"displayName": "Super Blink"}
					}
				],
				"facets": [ 
					{
						"id": 300,
						"name": "facet_a",
						"language": {"displayName": "Facet A - Slow"}
					},
					{
						"id": 301,
						"name": "facet_b",
						"language": {"displayName": "Facet B - Damage"}
					}
				]
			}
		}
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	results, err := fetchAndEnrichStratz(ts.URL, "dummy_token")
	if err != nil {
		t.Fatalf("Function returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 hero, got %d", len(results))
	}

	var hero map[string]any
	if err := json.Unmarshal([]byte(results[0]), &hero); err != nil {
		t.Fatalf("Result JSON parse error: %v", err)
	}

	abilities, ok := hero["abilities"].([]any)
	if !ok || len(abilities) != 1 {
		t.Fatalf("Abilities check failed. Found: %v", hero["abilities"])
	}
	ab1 := abilities[0].(map[string]any)
	langAb, _ := ab1["language"].(map[string]any)
	if langAb["displayName"] != "Super Blink" {
		t.Errorf("Ability data mismatch. Expected 'Super Blink', got %v", langAb["displayName"])
	}

	talents, ok := hero["talents"].([]any)
	if !ok || len(talents) != 1 {
		t.Fatalf("Talents check failed. Expected 1 talent.")
	}
	tal1 := talents[0].(map[string]any)
	langTal, _ := tal1["language"].(map[string]any)
	if langTal["displayName"] != "Talent +10 Damage" {
		t.Errorf("Talent displayName mismatch. Got %v", langTal["displayName"])
	}

	facets, ok := hero["facets"].([]any)
	if !ok || len(facets) != 2 {
		t.Fatalf("Expected 2 facets, got %d.", len(facets))
	}

	facetA := facets[0].(map[string]any)
	langFacetA, _ := facetA["language"].(map[string]any)
	if langFacetA["displayName"] != "Facet A - Slow" {
		t.Errorf("Facet A displayName mismatch. Got %v", langFacetA["displayName"])
	}
	if facetA["slot"] != nil {
		t.Errorf("Facet should NOT have 'slot' field.")
	}
}
