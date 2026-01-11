package dota

type GameState struct {
	Map       Map                `json:"map"`
	Player    Player             `json:"player"`
	Hero      Hero               `json:"hero"`
	Abilities map[string]Ability `json:"abilities"`
	Items     map[string]Item    `json:"items"`
}

type Map struct {
	Name      string `json:"name"`
	MatchID   string `json:"matchid"`
	GameTime  int    `json:"game_time"`
	ClockTime int    `json:"clock_time"`
	GameState string `json:"game_state"`
	DayTime   bool   `json:"daytime"`
}

type Player struct {
	TeamName string `json:"team_name"`
	Name     string `json:"name"`
	Gold     int    `json:"gold"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
	Assists  int    `json:"assists"`
	LastHits int    `json:"last_hits"`
	GPM      int    `json:"gpm"`
	XPM      int    `json:"xpm"`
}

type Hero struct {
	Name      string `json:"name"`
	Level     int    `json:"level"`
	Health    int    `json:"health"`
	MaxHealth int    `json:"max_health"`
	Mana      int    `json:"mana"`
	MaxMana   int    `json:"max_mana"`
	XPos      int    `json:"xpos"`
	YPos      int    `json:"ypos"`
	Alive     bool   `json:"alive"`
	Stunned   bool   `json:"stunned"`
	Silenced  bool   `json:"silenced"`
	Talent1   bool   `json:"talent_1"`
	Talent2   bool   `json:"talent_2"`
	Talent3   bool   `json:"talent_3"`
	Talent4   bool   `json:"talent_4"`
	Talent5   bool   `json:"talent_5"`
	Talent6   bool   `json:"talent_6"`
	Talent7   bool   `json:"talent_7"`
	Talent8   bool   `json:"talent_8"`
	Facet     int    `json:"facet"` // Аспект
}

type Ability struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	CanCast     bool   `json:"can_cast"`
	Cooldown    int    `json:"cooldown"`
	MaxCooldown int    `json:"max_cooldown"`
	IsUltimate  bool   `json:"ultimate"`
}

type Item struct {
	Name    string `json:"name"`
	CanCast bool   `json:"can_cast"`
	Charges int    `json:"charges"`
}
