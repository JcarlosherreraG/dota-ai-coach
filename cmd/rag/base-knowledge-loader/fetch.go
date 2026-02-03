package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

type StratzResponse struct {
	Data struct {
		Constants struct {
			Heroes    []HeroRaw        `json:"heroes"`
			Abilities []map[string]any `json:"abilities"`
			Facets    []map[string]any `json:"facets"`
		} `json:"constants"`
	} `json:"data"`
}

type Abilities struct {
	AbilityID int `json:"abilityId"`
}

type Talents struct {
	Slot      int `json:"slot"`
	AbilityID int `json:"abilityId"`
}

type Facets struct {
	FacetID int `json:"facetId"`
}

type HeroRaw struct {
	ID          int             `json:"id"`
	DisplayName string          `json:"displayName"`
	Stats       json.RawMessage `json:"stats"`
	Talents     []Talents       `json:"talents"`
	Abilities   []Abilities     `json:"abilities"`
	Facets      []Facets        `json:"facets"`
}

// fetchAndSplitJSON downloads JSON from a URL and splits it into individual stringified objects.
func fetchAndSplitJSON(url string, token string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad HTTP status: %d %s for url: %s", resp.StatusCode, resp.Status, url)
	}

	body, _ := io.ReadAll(resp.Body)
	var container any
	if err := json.Unmarshal(body, &container); err != nil {
		preview := string(body)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return nil, fmt.Errorf("invalid JSON from source: %v. Body preview: %s", err, preview)
	}

	var results []string
	// Handle both map and array JSON structures
	if dataMap, ok := container.(map[string]any); ok {
		keys := make([]string, 0, len(dataMap))
		for k := range dataMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val := dataMap[key]
			if valMap, ok := val.(map[string]any); ok {
				valMap["_system_name"] = key
			}
			bytes, _ := json.Marshal(val)
			results = append(results, string(bytes))
		}
	} else if dataArray, ok := container.([]any); ok {
		for _, val := range dataArray {
			bytes, _ := json.Marshal(val)
			results = append(results, string(bytes))
		}
	}
	return results, nil
}

func fetchAndEnrichStratz(url string, token string) ([]string, error) {
	// 1. Query preparing
	query := `
		query {
		  constants {
		    heroes {
		      id
		      displayName
		      stats {
		        attackType
		        attackRate
		        strengthBase
		        agilityBase
		        intelligenceBase
		      }
		      facets {
		        slot
		        facetId
		      }
		      talents {
		        slot
		        abilityId 
		      }
		      abilities {
		        abilityId
		      }
		    }
		   
		    abilities {
		      id
		      isTalent
		      attributes {
		        name
		        value
		      }
		      language {
		        displayName
		        description
		        notes
		      }
		    }
		    facets {
		      id
		      name
		      language {
		        displayName
		        description
		      }
		    }
		  }
		}
    `
	reqBody, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	}

	// 2. Process request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(b))
	}

	// 3. Decode response
	var result StratzResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 4. Create ability map: ID -> Ablitity Object
	abilitiesMap := make(map[int]map[string]any)
	for _, ab := range result.Data.Constants.Abilities {
		if idVal, ok := ab["id"].(float64); ok {
			abilitiesMap[int(idVal)] = ab
		}
	}
	// 5. Create Facets map: ID -> Facet Object
	facetsMap := make(map[int]map[string]any)
	for _, facet := range result.Data.Constants.Facets {
		if idVal, ok := facet["id"].(float64); ok {
			facetsMap[int(idVal)] = facet
		}
	}

	var jsonStrings []string

	for _, hero := range result.Data.Constants.Heroes {
		// Create the final hero map
		heroFinal := map[string]any{
			"id":          hero.ID,
			"displayName": hero.DisplayName,
			"stats":       hero.Stats,
		}

		// Set abilities to map
		fullAbilities := make([]map[string]any, 0)
		for _, ref := range hero.Abilities {
			if ab, exists := abilitiesMap[ref.AbilityID]; exists {
				fullAbilities = append(fullAbilities, ab)
			}
		}
		heroFinal["abilities"] = fullAbilities

		// Set talents to map
		fullTalents := make([]map[string]any, 0)
		for _, ref := range hero.Talents {
			if ab, exists := abilitiesMap[ref.AbilityID]; exists {
				talentWithSlot := make(map[string]any)
				for k, v := range ab {
					talentWithSlot[k] = v
				}
				talentWithSlot["slot"] = ref.Slot
				fullTalents = append(fullTalents, talentWithSlot)
			}
		}
		heroFinal["talents"] = fullTalents

		// Set facets to map
		fullFacets := make([]map[string]any, 0)
		for _, ref := range hero.Facets {
			if facet, exists := facetsMap[ref.FacetID]; exists {
				facetWithSlot := make(map[string]any)
				for k, v := range facet {
					facetWithSlot[k] = v
				}
				fullFacets = append(fullFacets, facetWithSlot)
			}
		}
		heroFinal["facets"] = fullFacets

		bytes, _ := json.Marshal(heroFinal)
		jsonStrings = append(jsonStrings, string(bytes))
	}

	return jsonStrings, nil
}
