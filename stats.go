package eznode

// ChainNodeStats is the stats of a chain node
type ChainNodeStats struct {
	Name          string         `json:"name"`
	CurrentHits   uint           `json:"current_hits"`
	TotalHits     uint64         `json:"total_hits"`
	Limits        uint           `json:"limits"`
	ResponseStats map[int]uint64 `json:"response_stats"`
	Priority      int            `json:"priority"`
	Disabled      bool           `json:"disabled"`
	Fails         uint           `json:"fails"`
}

// ChainStats is the stats of a chain
type ChainStats struct {
	Id    string           `json:"id"`
	Nodes []ChainNodeStats `json:"nodes"`
}
