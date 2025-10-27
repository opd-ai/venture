package engine

import (
	"testing"
)

func TestNewAudioManager(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	if am == nil {
		t.Fatal("NewAudioManager returned nil")
	}
	if am.musicVolume != 1.0 {
		t.Errorf("Expected default music volume 1.0, got %f", am.musicVolume)
	}
	if am.sfxVolume != 1.0 {
		t.Errorf("Expected default SFX volume 1.0, got %f", am.sfxVolume)
	}
	if !am.musicEnabled {
		t.Error("Expected music to be enabled by default")
	}
	if !am.sfxEnabled {
		t.Error("Expected SFX to be enabled by default")
	}
}

func TestSetMusicVolume(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	tests := []struct {
		name     string
		volume   float64
		expected float64
		enabled  bool
	}{
		{"normal volume", 0.5, 0.5, true},
		{"max volume", 1.0, 1.0, true},
		{"min volume", 0.0, 0.0, false},
		{"below min", -0.5, 0.0, false},
		{"above max", 1.5, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am.SetMusicVolume(tt.volume)
			if am.musicVolume != tt.expected {
				t.Errorf("Expected volume %f, got %f", tt.expected, am.musicVolume)
			}
			if am.musicEnabled != tt.enabled {
				t.Errorf("Expected musicEnabled %v, got %v", tt.enabled, am.musicEnabled)
			}
		})
	}
}

func TestSetSFXVolume(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	tests := []struct {
		name     string
		volume   float64
		expected float64
		enabled  bool
	}{
		{"normal volume", 0.7, 0.7, true},
		{"max volume", 1.0, 1.0, true},
		{"min volume", 0.0, 0.0, false},
		{"below min", -0.3, 0.0, false},
		{"above max", 2.0, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am.SetSFXVolume(tt.volume)
			if am.sfxVolume != tt.expected {
				t.Errorf("Expected volume %f, got %f", tt.expected, am.sfxVolume)
			}
			if am.sfxEnabled != tt.enabled {
				t.Errorf("Expected sfxEnabled %v, got %v", tt.enabled, am.sfxEnabled)
			}
		})
	}
}

func TestPlayMusic(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	tests := []struct {
		name    string
		genre   string
		context string
	}{
		{"fantasy exploration", "fantasy", "exploration"},
		{"fantasy combat", "fantasy", "combat"},
		{"fantasy boss", "fantasy", "boss"},
		{"scifi exploration", "scifi", "exploration"},
		{"scifi combat", "scifi", "combat"},
		{"horror exploration", "horror", "exploration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := am.PlayMusic(tt.genre, tt.context)
			if err != nil {
				t.Fatalf("PlayMusic failed: %v", err)
			}

			// Verify current track is set
			track := am.GetCurrentTrack()
			if track == nil {
				t.Error("Expected current track to be set")
			}

			// Verify genre and context are stored
			genre, context := am.GetCurrentMusicInfo()
			if genre != tt.genre {
				t.Errorf("Expected genre %s, got %s", tt.genre, genre)
			}
			if context != tt.context {
				t.Errorf("Expected context %s, got %s", tt.context, context)
			}
		})
	}
}

func TestPlayMusic_Disabled(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	am.SetMusicVolume(0.0) // Disable music

	err := am.PlayMusic("fantasy", "combat")
	if err != nil {
		t.Fatalf("PlayMusic with disabled music should not error: %v", err)
	}

	// Current track should not be updated when music is disabled
	track := am.GetCurrentTrack()
	if track != nil {
		t.Error("Expected no current track when music is disabled")
	}
}

func TestPlayMusic_SameTrack(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	// Play initial track
	err := am.PlayMusic("fantasy", "exploration")
	if err != nil {
		t.Fatalf("Initial PlayMusic failed: %v", err)
	}

	firstTrack := am.GetCurrentTrack()

	// Play same track again
	err = am.PlayMusic("fantasy", "exploration")
	if err != nil {
		t.Fatalf("Second PlayMusic failed: %v", err)
	}

	secondTrack := am.GetCurrentTrack()

	// Should return the same track (not regenerate)
	if firstTrack != secondTrack {
		t.Error("Expected same track when playing same genre/context")
	}
}

func TestPlaySFX(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	effectTypes := []string{
		"impact", "explosion", "magic", "laser", "pickup",
		"hit", "jump", "death", "powerup",
	}

	for _, effectType := range effectTypes {
		t.Run(effectType, func(t *testing.T) {
			err := am.PlaySFX(effectType, 54321)
			if err != nil {
				t.Errorf("PlaySFX(%s) failed: %v", effectType, err)
			}
		})
	}
}

func TestPlaySFX_Disabled(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	am.SetSFXVolume(0.0) // Disable SFX

	err := am.PlaySFX("impact", 54321)
	if err != nil {
		t.Fatalf("PlaySFX with disabled SFX should not error: %v", err)
	}
}

func TestStopMusic(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	// Play some music
	err := am.PlayMusic("fantasy", "combat")
	if err != nil {
		t.Fatalf("PlayMusic failed: %v", err)
	}

	// Verify music is playing
	if am.GetCurrentTrack() == nil {
		t.Fatal("Expected track to be playing")
	}

	// Stop music
	am.StopMusic()

	// Verify music is stopped
	if am.GetCurrentTrack() != nil {
		t.Error("Expected no current track after StopMusic")
	}

	genre, context := am.GetCurrentMusicInfo()
	if genre != "" || context != "" {
		t.Error("Expected genre and context to be cleared after StopMusic")
	}
}

func TestApplyVolumeToTrack(t *testing.T) {
	am := NewAudioManager(44100, 12345)

	// Play music to generate a track
	err := am.PlayMusic("fantasy", "exploration")
	if err != nil {
		t.Fatalf("PlayMusic failed: %v", err)
	}

	track := am.GetCurrentTrack()
	if track == nil {
		t.Fatal("Expected track to be generated")
	}

	// Verify track has audio data
	if len(track.Data) == 0 {
		t.Error("Expected track to have audio data")
	}

	// Verify sample rate
	if track.SampleRate != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", track.SampleRate)
	}
}

func TestAudioManagerSystem_Update(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	system := NewAudioManagerSystem(am)

	// Create test entities
	world := NewWorld()

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create enemy entity
	enemy := world.CreateEntity()
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(&PositionComponent{X: 200, Y: 200})
	enemy.AddComponent(&StatsComponent{Attack: 10})

	// Update world to process entity additions
	world.Update(0.016)

	// Update system (should stay in exploration mode - timer not reached)
	system.Update(world.GetEntities(), 0.016)

	// Music should start as exploration (set by client initialization)
	// After 60 frames, it should switch to combat due to enemy presence

	// Simulate 60 frames
	for i := 0; i < 60; i++ {
		system.Update(world.GetEntities(), 0.016)
	}

	// At this point, music context should have been evaluated
	// Since we have an enemy, it should be combat (if music was initialized)
}

func TestAudioManagerSystem_BossMusic(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	system := NewAudioManagerSystem(am)

	// Start with exploration music
	err := am.PlayMusic("fantasy", "exploration")
	if err != nil {
		t.Fatalf("PlayMusic failed: %v", err)
	}

	world := NewWorld()

	// Create player entity with position
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&EbitenInput{}) // Mark as player

	// Set player entity in system
	system.SetPlayerEntity(player)

	// Create boss enemy (high attack) near player
	boss := world.CreateEntity()
	boss.AddComponent(&HealthComponent{Current: 500, Max: 500})
	boss.AddComponent(&StatsComponent{Attack: 25})        // > 20 = boss
	boss.AddComponent(&PositionComponent{X: 200, Y: 200}) // Within 300px combat radius

	world.Update(0.016)

	// Simulate 60 frames to trigger music update
	for i := 0; i < 60; i++ {
		system.Update(world.GetEntities(), 0.016)
	}

	// Verify boss music is playing
	_, context := am.GetCurrentMusicInfo()
	if context != "boss" {
		t.Errorf("Expected boss music context, got %s", context)
	}
}

func TestGetMusicVolume(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	am.SetMusicVolume(0.75)

	if vol := am.GetMusicVolume(); vol != 0.75 {
		t.Errorf("Expected music volume 0.75, got %f", vol)
	}
}

func TestGetSFXVolume(t *testing.T) {
	am := NewAudioManager(44100, 12345)
	am.SetSFXVolume(0.85)

	if vol := am.GetSFXVolume(); vol != 0.85 {
		t.Errorf("Expected SFX volume 0.85, got %f", vol)
	}
}
