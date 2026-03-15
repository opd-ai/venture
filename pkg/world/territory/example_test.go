package territory_test

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/world/territory"
)

// fixedTimeProvider is a deterministic TimeProvider for examples.
type fixedTimeProvider struct{ t time.Time }

func (f *fixedTimeProvider) Now() time.Time { return f.t }

// ExampleManager_CreateTerritory demonstrates basic territory creation and ownership query.
func ExampleManager_CreateTerritory() {
	m := territory.NewManager()
	t, err := m.CreateTerritory("zone-1", territory.TerritoryCoords{ChunkX: 0, ChunkZ: 0})
	if err != nil {
		panic(err)
	}
	fmt.Println(t.Status)
	// Output: Neutral
}

// ExampleManager_DeclareWar demonstrates the war declaration workflow between two guilds.
func ExampleManager_DeclareWar() {
	tp := &fixedTimeProvider{t: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	m := territory.NewManagerWithTimeProvider(tp)

	// Assign territory to a guild before declaring war.
	if _, err := m.CreateTerritory("zone-2", territory.TerritoryCoords{ChunkX: 1, ChunkZ: 0}); err != nil {
		panic(err)
	}
	if err := m.AssignOwner("zone-2", "guild-red"); err != nil {
		panic(err)
	}

	war, err := m.DeclareWar("guild-blue", "guild-red")
	if err != nil {
		panic(err)
	}
	fmt.Println(war.AttackerGuild, "vs", war.DefenderGuild)
	// Output: guild-blue vs guild-red
}
