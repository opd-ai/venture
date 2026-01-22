package advanced

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.players == nil {
		t.Error("NewManager() players map is nil")
	}
	if m.talentTrees == nil {
		t.Error("NewManager() talentTrees map is nil")
	}
	if len(m.synergies) == 0 {
		t.Error("NewManager() synergies is empty")
	}
}

func TestSetPrimaryClass(t *testing.T) {
	m := NewManager()

	err := m.SetPrimaryClass("player1", ClassWarrior)
	if err != nil {
		t.Fatalf("SetPrimaryClass() error = %v", err)
	}

	player, err := m.GetPlayerClass("player1")
	if err != nil {
		t.Fatalf("GetPlayerClass() error = %v", err)
	}
	if player.PrimaryClass != ClassWarrior {
		t.Errorf("PrimaryClass = %v, want %v", player.PrimaryClass, ClassWarrior)
	}

	err = m.SetPrimaryClass("player2", ClassID("invalid"))
	if err == nil {
		t.Error("SetPrimaryClass() with invalid class should error")
	}
}

func TestSetSecondaryClass(t *testing.T) {
	m := NewManager()

	m.SetPrimaryClass("player1", ClassWarrior)

	err := m.SetSecondaryClass("player1", ClassMage)
	if err != nil {
		t.Fatalf("SetSecondaryClass() error = %v", err)
	}

	player, _ := m.GetPlayerClass("player1")
	if player.SecondaryClass != ClassMage {
		t.Errorf("SecondaryClass = %v, want %v", player.SecondaryClass, ClassMage)
	}

	err = m.SetSecondaryClass("player1", ClassWarrior)
	if err == nil {
		t.Error("SetSecondaryClass() with same as primary should error")
	}

	err = m.SetSecondaryClass("nonexistent", ClassRogue)
	if err == nil {
		t.Error("SetSecondaryClass() with nonexistent player should error")
	}
}

func TestSetPrestigeClass(t *testing.T) {
	m := NewManager()

	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 25)

	err := m.SetPrestigeClass("player1", PrestigeBladeMaster)
	if err != nil {
		t.Fatalf("SetPrestigeClass() error = %v", err)
	}

	player, _ := m.GetPlayerClass("player1")
	if player.PrestigeClass != PrestigeBladeMaster {
		t.Errorf("PrestigeClass = %v, want %v", player.PrestigeClass, PrestigeBladeMaster)
	}

	m.SetPrimaryClass("player2", ClassMage)
	m.SetLevel("player2", 15)
	err = m.SetPrestigeClass("player2", PrestigeArchmage)
	if err == nil {
		t.Error("SetPrestigeClass() with insufficient level should error")
	}

	m.SetPrimaryClass("player3", ClassRogue)
	m.SetLevel("player3", 25)
	err = m.SetPrestigeClass("player3", PrestigeBladeMaster)
	if err == nil {
		t.Error("SetPrestigeClass() with wrong primary class should error")
	}
}

func TestSetLevel(t *testing.T) {
	m := NewManager()

	err := m.SetLevel("player1", 10)
	if err != nil {
		t.Fatalf("SetLevel() error = %v", err)
	}

	player, _ := m.GetPlayerClass("player1")
	if player.Level != 10 {
		t.Errorf("Level = %v, want 10", player.Level)
	}
	if player.TalentPoints.PointsTotal != 10 {
		t.Errorf("PointsTotal = %v, want 10", player.TalentPoints.PointsTotal)
	}

	err = m.SetLevel("player1", 15)
	if err != nil {
		t.Fatalf("SetLevel() increase error = %v", err)
	}

	player, _ = m.GetPlayerClass("player1")
	if player.Level != 15 {
		t.Errorf("Level = %v, want 15", player.Level)
	}
	if player.TalentPoints.PointsTotal != 15 {
		t.Errorf("PointsTotal = %v, want 15", player.TalentPoints.PointsTotal)
	}

	err = m.SetLevel("player1", 10)
	if err == nil {
		t.Error("SetLevel() decrease should error")
	}
}

func TestAllocateTalent(t *testing.T) {
	m := NewManager()

	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 10)

	err := m.AllocateTalent("player1", "warrior_power_strike")
	if err != nil {
		t.Fatalf("AllocateTalent() error = %v", err)
	}

	player, _ := m.GetPlayerClass("player1")
	if player.TalentPoints.Talents["warrior_power_strike"] != 1 {
		t.Errorf("Talent rank = %v, want 1", player.TalentPoints.Talents["warrior_power_strike"])
	}
	if player.TalentPoints.PointsSpent != 1 {
		t.Errorf("PointsSpent = %v, want 1", player.TalentPoints.PointsSpent)
	}

	for i := 0; i < 4; i++ {
		m.AllocateTalent("player1", "warrior_power_strike")
	}

	player, _ = m.GetPlayerClass("player1")
	if player.TalentPoints.Talents["warrior_power_strike"] != 5 {
		t.Errorf("Talent rank = %v, want 5", player.TalentPoints.Talents["warrior_power_strike"])
	}

	err = m.AllocateTalent("player1", "warrior_power_strike")
	if err == nil {
		t.Error("AllocateTalent() beyond max rank should error")
	}

	err = m.AllocateTalent("player1", "warrior_weapon_mastery")
	if err != nil {
		t.Errorf("AllocateTalent() with prerequisite met error = %v", err)
	}

	// berserker_rage requires weapon_mastery (now met)
	err = m.AllocateTalent("player1", "warrior_berserker_rage")
	if err != nil {
		t.Errorf("AllocateTalent() berserker_rage with prerequisite met error = %v", err)
	}

	// execute requires critical_hit (not allocated yet)
	err = m.AllocateTalent("player1", "warrior_execute")
	if err == nil {
		t.Error("AllocateTalent() without prerequisite should error")
	}
}

func TestRespecTalents(t *testing.T) {
	m := NewManager()

	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 10)

	m.AllocateTalent("player1", "warrior_power_strike")
	m.AllocateTalent("player1", "warrior_power_strike")

	cost := m.GetRespecCost("player1")
	if cost != 1000 {
		t.Errorf("GetRespecCost() = %v, want 1000", cost)
	}

	err := m.RespecTalents("player1", 500)
	if err == nil {
		t.Error("RespecTalents() with insufficient gold should error")
	}

	err = m.RespecTalents("player1", 1000)
	if err != nil {
		t.Fatalf("RespecTalents() error = %v", err)
	}

	player, _ := m.GetPlayerClass("player1")
	if player.TalentPoints.PointsSpent != 0 {
		t.Errorf("PointsSpent after respec = %v, want 0", player.TalentPoints.PointsSpent)
	}
	if len(player.TalentPoints.Talents) != 0 {
		t.Errorf("Talents after respec = %v, want empty", player.TalentPoints.Talents)
	}

	cost = m.GetRespecCost("player1")
	if cost != 1500 {
		t.Errorf("GetRespecCost() after respec = %v, want 1500", cost)
	}
}

func TestCalculateTotalStats(t *testing.T) {
	m := NewManager()

	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 10)

	stats, err := m.CalculateTotalStats("player1")
	if err != nil {
		t.Fatalf("CalculateTotalStats() error = %v", err)
	}

	if stats.Health <= 0 {
		t.Error("CalculateTotalStats() Health should be positive")
	}
	if stats.Strength <= 0 {
		t.Error("CalculateTotalStats() Strength should be positive")
	}

	m.SetSecondaryClass("player1", ClassMage)
	stats, err = m.CalculateTotalStats("player1")
	if err != nil {
		t.Fatalf("CalculateTotalStats() with secondary error = %v", err)
	}

	if stats.Mana <= 0 {
		t.Error("CalculateTotalStats() Mana from secondary should be positive")
	}

	m.AllocateTalent("player1", "warrior_power_strike")
	stats2, _ := m.CalculateTotalStats("player1")
	if stats2.Strength <= stats.Strength {
		t.Error("CalculateTotalStats() Strength should increase with talent")
	}

	// Level up to 20 to unlock prestige class
	m.SetLevel("player1", 20)
	m.SetPrestigeClass("player1", PrestigeBladeMaster)
	stats3, _ := m.CalculateTotalStats("player1")
	if stats3.Strength <= stats2.Strength {
		t.Error("CalculateTotalStats() Strength should increase with prestige class")
	}
}

func TestGetSynergyBonus(t *testing.T) {
	m := NewManager()

	synergy := m.getSynergyBonus(ClassWarrior, ClassMage)
	if synergy == nil {
		t.Error("getSynergyBonus() for Warrior+Mage should not be nil")
	}
	if synergy != nil {
		if synergy.Name != "Spellsword" {
			t.Errorf("getSynergyBonus() Name = %v, want Spellsword", synergy.Name)
		}
	}

	synergy = m.getSynergyBonus(ClassMage, ClassWarrior)
	if synergy == nil {
		t.Error("getSynergyBonus() should work in reverse order")
	}

	synergy = m.getSynergyBonus(ClassWarrior, "")
	if synergy != nil {
		t.Error("getSynergyBonus() with empty secondary should be nil")
	}

	synergy = m.getSynergyBonus(ClassWarrior, ClassWarrior)
	if synergy != nil {
		t.Error("getSynergyBonus() for same classes should be nil")
	}
}

func TestGetTalentTree(t *testing.T) {
	m := NewManager()

	tree, err := m.GetTalentTree(ClassWarrior)
	if err != nil {
		t.Fatalf("GetTalentTree() error = %v", err)
	}
	if tree == nil {
		t.Fatal("GetTalentTree() returned nil")
	}

	if len(tree.Offensive) != 10 {
		t.Errorf("Offensive talents = %v, want 10", len(tree.Offensive))
	}
	if len(tree.Defensive) != 10 {
		t.Errorf("Defensive talents = %v, want 10", len(tree.Defensive))
	}
	if len(tree.Utility) != 10 {
		t.Errorf("Utility talents = %v, want 10", len(tree.Utility))
	}

	_, err = m.GetTalentTree(ClassID("invalid"))
	if err == nil {
		t.Error("GetTalentTree() with invalid class should error")
	}
}

func TestGetAllSynergies(t *testing.T) {
	m := NewManager()

	synergies := m.GetAllSynergies()
	if len(synergies) == 0 {
		t.Error("GetAllSynergies() returned empty slice")
	}

	for _, synergy := range synergies {
		if synergy.Name == "" {
			t.Error("GetAllSynergies() returned synergy with empty name")
		}
		if synergy.Primary == "" || synergy.Secondary == "" {
			t.Error("GetAllSynergies() returned synergy with empty class")
		}
	}
}

func TestConcurrency(t *testing.T) {
	m := NewManager()

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			m.SetPrimaryClass("player1", ClassWarrior)
			m.CalculateTotalStats("player1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			m.SetLevel("player2", 10)
			m.AllocateTalent("player2", "warrior_power_strike")
		}
		done <- true
	}()

	<-done
	<-done
}

func BenchmarkSetPrimaryClass(b *testing.B) {
	m := NewManager()
	for i := 0; i < b.N; i++ {
		m.SetPrimaryClass("player1", ClassWarrior)
	}
}

func BenchmarkAllocateTalent(b *testing.B) {
	m := NewManager()
	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.AllocateTalent("player1", "warrior_power_strike")
		if i%5 == 4 {
			m.RespecTalents("player1", 10000)
		}
	}
}

func BenchmarkCalculateTotalStats(b *testing.B) {
	m := NewManager()
	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetSecondaryClass("player1", ClassMage)
	m.SetLevel("player1", 25)
	m.SetPrestigeClass("player1", PrestigeBladeMaster)
	m.AllocateTalent("player1", "warrior_power_strike")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CalculateTotalStats("player1")
	}
}

func BenchmarkRespecTalents(b *testing.B) {
	m := NewManager()
	m.SetPrimaryClass("player1", ClassWarrior)
	m.SetLevel("player1", 100)

	for i := 0; i < 10; i++ {
		m.AllocateTalent("player1", "warrior_power_strike")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RespecTalents("player1", 10000)
	}
}

// TestAllClassesTalentTrees verifies all 15 base classes have complete talent trees
func TestAllClassesTalentTrees(t *testing.T) {
	m := NewManager()

	allClasses := []ClassID{
		ClassWarrior, ClassBerserker, ClassPaladin, ClassKnight,
		ClassRogue, ClassAssassin, ClassRanger, ClassNinja,
		ClassMage, ClassElementalist, ClassNecromancer, ClassEnchanter,
		ClassCleric, ClassBard, ClassDruid,
	}

	for _, classID := range allClasses {
		t.Run(string(classID), func(t *testing.T) {
			tree, err := m.GetTalentTree(classID)
			if err != nil {
				t.Fatalf("GetTalentTree(%s) error = %v", classID, err)
			}
			if tree == nil {
				t.Fatalf("GetTalentTree(%s) returned nil", classID)
			}

			// Each class should have 30 talents total (10 per category)
			if len(tree.Offensive) != 10 {
				t.Errorf("%s Offensive talents = %d, want 10", classID, len(tree.Offensive))
			}
			if len(tree.Defensive) != 10 {
				t.Errorf("%s Defensive talents = %d, want 10", classID, len(tree.Defensive))
			}
			if len(tree.Utility) != 10 {
				t.Errorf("%s Utility talents = %d, want 10", classID, len(tree.Utility))
			}

			// Verify tree name and class ID are set correctly
			if tree.ClassID != classID {
				t.Errorf("%s tree.ClassID = %s, want %s", classID, tree.ClassID, classID)
			}
			if tree.Name == "" {
				t.Errorf("%s tree.Name is empty", classID)
			}
		})
	}
}

// TestTotalTalentCount verifies we have the documented 450 talents (15 classes * 30 talents)
func TestTotalTalentCount(t *testing.T) {
	m := NewManager()

	allClasses := []ClassID{
		ClassWarrior, ClassBerserker, ClassPaladin, ClassKnight,
		ClassRogue, ClassAssassin, ClassRanger, ClassNinja,
		ClassMage, ClassElementalist, ClassNecromancer, ClassEnchanter,
		ClassCleric, ClassBard, ClassDruid,
	}

	totalTalents := 0
	for _, classID := range allClasses {
		tree, err := m.GetTalentTree(classID)
		if err != nil {
			t.Fatalf("GetTalentTree(%s) error = %v", classID, err)
		}
		totalTalents += len(tree.Offensive) + len(tree.Defensive) + len(tree.Utility)
	}

	expectedTotal := 15 * 30 // 15 classes * 30 talents each
	if totalTalents != expectedTotal {
		t.Errorf("Total talents = %d, want %d", totalTalents, expectedTotal)
	}
}
