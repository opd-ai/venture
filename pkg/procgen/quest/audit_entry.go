package quest

import "github.com/opd-ai/venture/pkg/procgen"

func init() {
	procgen.RegisterAuditEntry(procgen.AuditEntry{
		Name:      "Quest",
		Generator: NewQuestGenerator(),
		Params: procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      5,
			GenreID:    "fantasy",
		},
	})
}
