package item

import "testing"

// TestItem_CanBeUsedByClass tests the class restriction checking.
// Phase 25.2: Class-specific equipment restrictions.
func TestItem_CanBeUsedByClass(t *testing.T) {
	tests := []struct {
		name              string
		classRestrictions []string
		checkClass        string
		want              bool
	}{
		{
			name:              "no restrictions - any class",
			classRestrictions: []string{},
			checkClass:        "warrior",
			want:              true,
		},
		{
			name:              "no restrictions - mage",
			classRestrictions: []string{},
			checkClass:        "mage",
			want:              true,
		},
		{
			name:              "warrior only - warrior checks",
			classRestrictions: []string{"warrior"},
			checkClass:        "warrior",
			want:              true,
		},
		{
			name:              "warrior only - mage checks",
			classRestrictions: []string{"warrior"},
			checkClass:        "mage",
			want:              false,
		},
		{
			name:              "multiple classes - warrior allowed",
			classRestrictions: []string{"warrior", "rogue", "ranger"},
			checkClass:        "warrior",
			want:              true,
		},
		{
			name:              "multiple classes - rogue allowed",
			classRestrictions: []string{"warrior", "rogue", "ranger"},
			checkClass:        "rogue",
			want:              true,
		},
		{
			name:              "multiple classes - mage not allowed",
			classRestrictions: []string{"warrior", "rogue", "ranger"},
			checkClass:        "mage",
			want:              false,
		},
		{
			name:              "caster classes - mage allowed",
			classRestrictions: []string{"mage", "cleric", "necromancer"},
			checkClass:        "mage",
			want:              true,
		},
		{
			name:              "caster classes - warrior not allowed",
			classRestrictions: []string{"mage", "cleric", "necromancer"},
			checkClass:        "warrior",
			want:              false,
		},
		{
			name:              "single class - exact match",
			classRestrictions: []string{"ranger"},
			checkClass:        "ranger",
			want:              true,
		},
		{
			name:              "single class - no match",
			classRestrictions: []string{"ranger"},
			checkClass:        "cleric",
			want:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{
				ClassRestrictions: tt.classRestrictions,
			}
			got := item.CanBeUsedByClass(tt.checkClass)
			if got != tt.want {
				t.Errorf("CanBeUsedByClass(%q) = %v, want %v (restrictions: %v)",
					tt.checkClass, got, tt.want, tt.classRestrictions)
			}
		})
	}
}

// TestItem_ClassRestrictions_AllClasses tests that all valid class names work.
func TestItem_ClassRestrictions_AllClasses(t *testing.T) {
	validClasses := []string{"warrior", "mage", "rogue", "ranger", "cleric", "necromancer"}

	for _, class := range validClasses {
		t.Run(class, func(t *testing.T) {
			// Create item restricted to this class
			item := &Item{
				ClassRestrictions: []string{class},
			}

			// Should be usable by this class
			if !item.CanBeUsedByClass(class) {
				t.Errorf("Item restricted to %s should be usable by %s", class, class)
			}

			// Should not be usable by other classes
			for _, otherClass := range validClasses {
				if otherClass == class {
					continue
				}
				if item.CanBeUsedByClass(otherClass) {
					t.Errorf("Item restricted to %s should not be usable by %s", class, otherClass)
				}
			}
		})
	}
}
