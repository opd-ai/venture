package book

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func init() {
	procgen.RegisterAuditEntry(procgen.AuditEntry{
		Name:      "Book",
		Generator: NewGenerator(),
		Params: procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      5,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"book_type": engine.BookTypeLore,
			},
		},
	})
}
