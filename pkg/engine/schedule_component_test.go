package engine

import (
	"testing"
)

func TestScheduleComponent_Type(t *testing.T) {
	sc := NewScheduleComponent(100, 200)
	if sc.Type() != "schedule" {
		t.Errorf("Expected type 'schedule', got '%s'", sc.Type())
	}
}

func TestScheduleComponent_NewScheduleComponent(t *testing.T) {
	sc := NewScheduleComponent(100.5, 200.5)

	if sc.HomeX != 100.5 {
		t.Errorf("Expected HomeX 100.5, got %f", sc.HomeX)
	}
	if sc.HomeY != 200.5 {
		t.Errorf("Expected HomeY 200.5, got %f", sc.HomeY)
	}
	if sc.MovementSpeed != 50.0 {
		t.Errorf("Expected MovementSpeed 50.0, got %f", sc.MovementSpeed)
	}
	if len(sc.Activities) != 0 {
		t.Errorf("Expected empty activities, got %d", len(sc.Activities))
	}
}

func TestScheduleComponent_AddActivity(t *testing.T) {
	sc := NewScheduleComponent(0, 0)

	sc.AddActivity(ActivityWork, 8, 17, 100, 100, "Workshop")

	if len(sc.Activities) != 1 {
		t.Fatalf("Expected 1 activity, got %d", len(sc.Activities))
	}

	act := sc.Activities[0]
	if act.ActivityType != ActivityWork {
		t.Errorf("Expected ActivityWork, got %s", act.ActivityType)
	}
	if act.StartHour != 8 {
		t.Errorf("Expected StartHour 8, got %d", act.StartHour)
	}
	if act.EndHour != 17 {
		t.Errorf("Expected EndHour 17, got %d", act.EndHour)
	}
	if act.LocationX != 100 {
		t.Errorf("Expected LocationX 100, got %f", act.LocationX)
	}
	if act.LocationName != "Workshop" {
		t.Errorf("Expected LocationName 'Workshop', got '%s'", act.LocationName)
	}
}

func TestScheduleComponent_GetCurrentActivity(t *testing.T) {
	tests := []struct {
		name       string
		activities int
		currentIdx int
		expectNil  bool
	}{
		{"empty schedule", 0, 0, true},
		{"valid index", 2, 1, false},
		{"negative index", 2, -1, true},
		{"index out of bounds", 2, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewScheduleComponent(0, 0)
			for i := 0; i < tt.activities; i++ {
				sc.AddActivity(ActivityWork, 8, 17, 100, 100, "Work")
			}
			sc.CurrentActivityIdx = tt.currentIdx

			result := sc.GetCurrentActivity()
			if tt.expectNil && result != nil {
				t.Errorf("Expected nil, got activity")
			}
			if !tt.expectNil && result == nil {
				t.Errorf("Expected activity, got nil")
			}
		})
	}
}

func TestScheduleComponent_GetActivityForHour(t *testing.T) {
	sc := NewScheduleComponent(0, 0)
	sc.AddActivity(ActivitySleep, 22, 6, 0, 0, "Home")
	sc.AddActivity(ActivityWork, 8, 12, 100, 100, "Work")
	sc.AddActivity(ActivityEat, 12, 13, 100, 100, "Work")
	sc.AddActivity(ActivityWork, 13, 17, 100, 100, "Work")

	tests := []struct {
		hour         int
		expectedType ActivityType
		expectNil    bool
	}{
		{0, ActivitySleep, false},  // Midnight - sleeping
		{3, ActivitySleep, false},  // Early morning - sleeping
		{5, ActivitySleep, false},  // 5am - still sleeping
		{7, "", true},              // 7am - gap in schedule
		{9, ActivityWork, false},   // 9am - working
		{12, ActivityEat, false},   // Noon - eating
		{15, ActivityWork, false},  // 3pm - working
		{18, "", true},             // 6pm - gap
		{23, ActivitySleep, false}, // 11pm - sleeping
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := sc.GetActivityForHour(tt.hour)
			if tt.expectNil && result != nil {
				t.Errorf("Hour %d: expected nil, got %s", tt.hour, result.ActivityType)
			}
			if !tt.expectNil {
				if result == nil {
					t.Errorf("Hour %d: expected %s, got nil", tt.hour, tt.expectedType)
				} else if result.ActivityType != tt.expectedType {
					t.Errorf("Hour %d: expected %s, got %s", tt.hour, tt.expectedType, result.ActivityType)
				}
			}
		})
	}
}

func TestScheduleComponent_UpdateActivityIndex(t *testing.T) {
	sc := NewScheduleComponent(0, 0)
	sc.AddActivity(ActivitySleep, 22, 6, 0, 0, "Home")
	sc.AddActivity(ActivityWork, 8, 17, 100, 100, "Work")

	tests := []struct {
		hour        int
		expectedIdx int
	}{
		{0, 0},  // Midnight -> sleep (idx 0)
		{10, 1}, // 10am -> work (idx 1)
		{23, 0}, // 11pm -> sleep (idx 0)
		{7, 0},  // 7am -> gap, defaults to first (idx 0)
	}

	for _, tt := range tests {
		sc.UpdateActivityIndex(tt.hour)
		if sc.CurrentActivityIdx != tt.expectedIdx {
			t.Errorf("Hour %d: expected idx %d, got %d", tt.hour, tt.expectedIdx, sc.CurrentActivityIdx)
		}
	}
}

func TestScheduleComponent_Serialize_Deserialize(t *testing.T) {
	original := NewScheduleComponent(100, 200)
	original.AddActivity(ActivityWork, 8, 17, 150, 250, "Shop")
	original.AddActivity(ActivitySleep, 22, 6, 100, 200, "Home")
	original.CurrentActivityIdx = 1
	original.IsMoving = true

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize returned empty data")
	}

	// Deserialize into new component
	restored := &ScheduleComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify values
	if restored.HomeX != original.HomeX {
		t.Errorf("HomeX mismatch: %f vs %f", restored.HomeX, original.HomeX)
	}
	if restored.HomeY != original.HomeY {
		t.Errorf("HomeY mismatch: %f vs %f", restored.HomeY, original.HomeY)
	}
	if restored.CurrentActivityIdx != original.CurrentActivityIdx {
		t.Errorf("CurrentActivityIdx mismatch: %d vs %d", restored.CurrentActivityIdx, original.CurrentActivityIdx)
	}
	if restored.IsMoving != original.IsMoving {
		t.Errorf("IsMoving mismatch: %v vs %v", restored.IsMoving, original.IsMoving)
	}
	if len(restored.Activities) != len(original.Activities) {
		t.Errorf("Activities count mismatch: %d vs %d", len(restored.Activities), len(original.Activities))
	}
}

func TestActivityType_Constants(t *testing.T) {
	// Verify all activity type constants are defined
	types := []ActivityType{
		ActivityWork,
		ActivityEat,
		ActivitySleep,
		ActivitySocialize,
		ActivityPatrol,
		ActivityIdle,
	}

	for _, at := range types {
		if at == "" {
			t.Error("Activity type constant is empty")
		}
	}
}
