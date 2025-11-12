package engine

import (
	"testing"
)

func TestCompanionComponent(t *testing.T) {
	comp := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Experience:    0.0,
		Level:         1,
		Behavior:      BehaviorPassive,
		Commands:      []CommandType{CommandFollow, CommandStay},
	}
	
	if comp.Type() != "companion" {
		t.Errorf("Expected type 'companion', got '%s'", comp.Type())
	}
	
	if comp.Loyalty != 50.0 {
		t.Errorf("Expected loyalty 50.0, got %f", comp.Loyalty)
	}
}

func TestCompanionStatsComponent(t *testing.T) {
	comp := &CompanionStatsComponent{
		Attack:  10.0,
		Defense: 8.0,
		Speed:   100.0,
		HP:      50.0,
		MaxHP:   50.0,
	}
	
	if comp.Type() != "companionstats" {
		t.Errorf("Expected type 'companionstats', got '%s'", comp.Type())
	}
}
