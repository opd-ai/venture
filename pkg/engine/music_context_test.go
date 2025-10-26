package engine

import (
	"testing"
	"time"
)

// TestMusicContext_String verifies context string conversion
func TestMusicContext_String(t *testing.T) {
	tests := []struct {
		name    string
		context MusicContext
		want    string
	}{
		{"exploration", MusicContextExploration, "exploration"},
		{"combat", MusicContextCombat, "combat"},
		{"boss", MusicContextBoss, "boss"},
		{"danger", MusicContextDanger, "danger"},
		{"victory", MusicContextVictory, "victory"},
		{"death", MusicContextDeath, "death"},
		{"unknown", MusicContext(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.context.String(); got != tt.want {
				t.Errorf("MusicContext.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMusicContext_Priority verifies priority ordering
func TestMusicContext_Priority(t *testing.T) {
	tests := []struct {
		name    string
		context MusicContext
		want    int
	}{
		{"boss highest", MusicContextBoss, 100},
		{"combat high", MusicContextCombat, 80},
		{"danger medium", MusicContextDanger, 60},
		{"exploration low", MusicContextExploration, 40},
		{"victory lower", MusicContextVictory, 20},
		{"death lowest", MusicContextDeath, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.context.Priority(); got != tt.want {
				t.Errorf("MusicContext.Priority() = %v, want %v", got, tt.want)
			}
		})
	}

	// Verify ordering
	if MusicContextBoss.Priority() <= MusicContextCombat.Priority() {
		t.Error("Boss priority should be higher than combat")
	}
	if MusicContextCombat.Priority() <= MusicContextDanger.Priority() {
		t.Error("Combat priority should be higher than danger")
	}
}

// TestNewMusicContextDetector verifies default configuration
func TestNewMusicContextDetector(t *testing.T) {
	detector := NewMusicContextDetector()

	if detector.CombatRadius != 300.0 {
		t.Errorf("CombatRadius = %v, want 300.0", detector.CombatRadius)
	}
	if detector.BossAttackThreshold != 20.0 {
		t.Errorf("BossAttackThreshold = %v, want 20.0", detector.BossAttackThreshold)
	}
	if detector.DangerHealthPercent != 0.2 {
		t.Errorf("DangerHealthPercent = %v, want 0.2", detector.DangerHealthPercent)
	}
}

// TestMusicContextDetector_DetectContext tests context detection logic
func TestMusicContextDetector_DetectContext(t *testing.T) {
	detector := NewMusicContextDetector()

	tests := []struct {
		name         string
		setupPlayer  func() *Entity
		setupEnemies func() []*Entity
		want         MusicContext
	}{
		{
			name: "no player returns exploration",
			setupPlayer: func() *Entity {
				return nil
			},
			setupEnemies: func() []*Entity {
				return []*Entity{}
			},
			want: MusicContextExploration,
		},
		{
			name: "dead player returns death",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 0, Max: 100})
				player.AddComponent(&DeadComponent{TimeOfDeath: 0})
				return player
			},
			setupEnemies: func() []*Entity {
				return []*Entity{}
			},
			want: MusicContextDeath,
		},
		{
			name: "healthy player alone returns exploration",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				return []*Entity{}
			},
			want: MusicContextExploration,
		},
		{
			name: "nearby enemy returns combat",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				world := NewWorld()
				enemy := world.CreateEntity()
				enemy.AddComponent(&PositionComponent{X: 200, Y: 100}) // 100 pixels away
				enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
				enemy.AddComponent(&TeamComponent{TeamID: 2}) // Different team
				enemy.AddComponent(&StatsComponent{Attack: 10})
				return []*Entity{enemy}
			},
			want: MusicContextCombat,
		},
		{
			name: "far enemy returns exploration",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				world := NewWorld()
				enemy := world.CreateEntity()
				enemy.AddComponent(&PositionComponent{X: 500, Y: 500}) // 565 pixels away
				enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
				enemy.AddComponent(&TeamComponent{TeamID: 2})
				return []*Entity{enemy}
			},
			want: MusicContextExploration,
		},
		{
			name: "nearby boss returns boss",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				world := NewWorld()
				boss := world.CreateEntity()
				boss.AddComponent(&PositionComponent{X: 150, Y: 150}) // Close
				boss.AddComponent(&HealthComponent{Current: 200, Max: 200})
				boss.AddComponent(&TeamComponent{TeamID: 2})
				boss.AddComponent(&StatsComponent{Attack: 25}) // Boss-level attack
				return []*Entity{boss}
			},
			want: MusicContextBoss,
		},
		{
			name: "low health returns danger",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 15, Max: 100}) // 15% health
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				return []*Entity{}
			},
			want: MusicContextDanger,
		},
		{
			name: "boss overrides danger",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 15, Max: 100}) // Low health
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				world := NewWorld()
				boss := world.CreateEntity()
				boss.AddComponent(&PositionComponent{X: 150, Y: 150})
				boss.AddComponent(&HealthComponent{Current: 200, Max: 200})
				boss.AddComponent(&TeamComponent{TeamID: 2})
				boss.AddComponent(&StatsComponent{Attack: 30})
				return []*Entity{boss}
			},
			want: MusicContextBoss,
		},
		{
			name: "dead enemy not counted",
			setupPlayer: func() *Entity {
				world := NewWorld()
				player := world.CreateEntity()
				player.AddComponent(&PositionComponent{X: 100, Y: 100})
				player.AddComponent(&HealthComponent{Current: 100, Max: 100})
				player.AddComponent(&TeamComponent{TeamID: 1})
				return player
			},
			setupEnemies: func() []*Entity {
				world := NewWorld()
				enemy := world.CreateEntity()
				enemy.AddComponent(&PositionComponent{X: 150, Y: 150})
				enemy.AddComponent(&HealthComponent{Current: 0, Max: 50})
				enemy.AddComponent(&DeadComponent{TimeOfDeath: 0})
				enemy.AddComponent(&TeamComponent{TeamID: 2})
				return []*Entity{enemy}
			},
			want: MusicContextExploration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setupPlayer()
			enemies := tt.setupEnemies()
			allEntities := append([]*Entity{player}, enemies...)

			got := detector.DetectContext(allEntities, player)
			if got != tt.want {
				t.Errorf("DetectContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewMusicTransitionManager verifies initial state
func TestNewMusicTransitionManager(t *testing.T) {
	manager := NewMusicTransitionManager()

	if manager.GetCurrentContext() != MusicContextExploration {
		t.Errorf("Initial context = %v, want exploration", manager.GetCurrentContext())
	}
	if manager.IsTransitioning() {
		t.Error("Should not be transitioning initially")
	}
	if manager.transitionCooldown != 10*time.Second {
		t.Errorf("Cooldown = %v, want 10s", manager.transitionCooldown)
	}
}

// TestMusicTransitionManager_ShouldTransition tests transition logic
func TestMusicTransitionManager_ShouldTransition(t *testing.T) {
	tests := []struct {
		name           string
		currentContext MusicContext
		newContext     MusicContext
		cooldownActive bool
		want           bool
	}{
		{
			name:           "same context no transition",
			currentContext: MusicContextExploration,
			newContext:     MusicContextExploration,
			cooldownActive: false,
			want:           false,
		},
		{
			name:           "boss interrupts immediately",
			currentContext: MusicContextExploration,
			newContext:     MusicContextBoss,
			cooldownActive: true,
			want:           true,
		},
		{
			name:           "death interrupts immediately",
			currentContext: MusicContextCombat,
			newContext:     MusicContextDeath,
			cooldownActive: true,
			want:           false, // Death priority is 10, combat is 80
		},
		{
			name:           "higher priority during cooldown",
			currentContext: MusicContextExploration,
			newContext:     MusicContextCombat,
			cooldownActive: true,
			want:           true,
		},
		{
			name:           "lower priority during cooldown blocked",
			currentContext: MusicContextCombat,
			newContext:     MusicContextExploration,
			cooldownActive: true,
			want:           false,
		},
		{
			name:           "after cooldown allows transition",
			currentContext: MusicContextCombat,
			newContext:     MusicContextExploration,
			cooldownActive: false,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMusicTransitionManager()
			manager.currentContext = tt.currentContext

			if tt.cooldownActive {
				manager.lastTransitionTime = time.Now()
			} else {
				manager.lastTransitionTime = time.Now().Add(-15 * time.Second)
			}

			got := manager.ShouldTransition(tt.newContext)
			if got != tt.want {
				t.Errorf("ShouldTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMusicTransitionManager_Lifecycle tests transition state management
func TestMusicTransitionManager_Lifecycle(t *testing.T) {
	manager := NewMusicTransitionManager()

	// Initial state
	if manager.IsTransitioning() {
		t.Error("Should not be transitioning initially")
	}

	// Begin transition
	manager.BeginTransition(MusicContextCombat)
	if !manager.IsTransitioning() {
		t.Error("Should be transitioning after BeginTransition")
	}
	if manager.GetCurrentContext() != MusicContextCombat {
		t.Errorf("Context = %v, want combat", manager.GetCurrentContext())
	}

	// Complete transition
	manager.CompleteTransition()
	if manager.IsTransitioning() {
		t.Error("Should not be transitioning after CompleteTransition")
	}
	if manager.GetCurrentContext() != MusicContextCombat {
		t.Errorf("Context = %v, want combat", manager.GetCurrentContext())
	}
}

// TestMusicTransitionManager_SetCooldown tests cooldown configuration
func TestMusicTransitionManager_SetCooldown(t *testing.T) {
	manager := NewMusicTransitionManager()

	newCooldown := 5 * time.Second
	manager.SetCooldown(newCooldown)

	if manager.transitionCooldown != newCooldown {
		t.Errorf("Cooldown = %v, want %v", manager.transitionCooldown, newCooldown)
	}
}

// TestMusicContextDetector_calculateDistance tests distance calculation
func TestMusicContextDetector_calculateDistance(t *testing.T) {
	detector := NewMusicContextDetector()

	tests := []struct {
		name string
		x1   float64
		y1   float64
		x2   float64
		y2   float64
		want float64
	}{
		{"same point", 0, 0, 0, 0, 0},
		{"horizontal", 0, 0, 3, 0, 3},
		{"vertical", 0, 0, 0, 4, 4},
		{"diagonal", 0, 0, 3, 4, 5}, // 3-4-5 triangle
		{"negative coords", -10, -10, 0, 0, 14.142135623730951},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.calculateDistance(tt.x1, tt.y1, tt.x2, tt.y2)
			diff := got - tt.want
			if diff < -0.0001 || diff > 0.0001 {
				t.Errorf("calculateDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
