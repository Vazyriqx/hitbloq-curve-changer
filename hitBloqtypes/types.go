package hitBloqtypes

type RankedList struct {
	ID                   string   `json:"_id"`
	AccumulationConstant float64  `json:"accumulation_constant"`
	CRCurve              CRCurve  `json:"cr_curve"`
	LeaderboardIDList    []string `json:"leaderboard_id_list"`
}

type CRCurve struct {
	Type        string      `json:"type"`
	Baseline    float64     `json:"baseline"`
	Cutoff      float64     `json:"cutoff"`
	Exponential float64     `json:"exponential"`
	Points      [][]float64 `json:"points"`
}

type Score struct {
	CR    map[string]float64 `json:"cr"`
	Score int                `json:"score"`
}

type LeaderboardInfo struct {
	Name             string             `json:"name"`
	Notes            int                `json:"notes"`
	ForcedStarRating map[string]float64 `json:"forced_star_rating"`
}
