package model

type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description,omitempty"`
	RolloutPercent int    `json:"rollout_percent"`
}
