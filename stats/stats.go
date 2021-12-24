package stats

type ChainNodeStats struct {
	Name        string `json:"name"`
	CurrentHits uint   `json:"current_hits"`
	TotalHits   uint64 `json:"total_hits"`
	Limits      uint   `json:"limits"`
	Priority    int    `json:"priority"`
}

type ChainStats struct {
	Id    string           `json:"id"`
	Nodes []ChainNodeStats `json:"nodes"`
}
