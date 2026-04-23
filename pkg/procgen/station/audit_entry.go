package station

import "github.com/opd-ai/venture/pkg/procgen"

func init() {
procgen.RegisterAuditEntry(procgen.AuditEntry{
Name:      "Station",
Generator: NewStationGenerator(),
Params: procgen.GenerationParams{
Difficulty: 0.5,
Depth:      5,
GenreID:    "fantasy",
},
})
}
