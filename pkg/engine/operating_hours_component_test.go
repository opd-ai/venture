package engine

import (
	"testing"
)

func TestOperatingHoursComponent_Type(t *testing.T) {
	comp := NewOperatingHoursComponent()
	if comp.Type() != "operatingHours" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "operatingHours")
	}
}

func TestNewOperatingHoursComponent_Defaults(t *testing.T) {
	comp := NewOperatingHoursComponent()

	if comp.OpenHour != 8 {
		t.Errorf("OpenHour = %d, want 8", comp.OpenHour)
	}
	if comp.CloseHour != 18 {
		t.Errorf("CloseHour = %d, want 18", comp.CloseHour)
	}
	if comp.IsAlwaysOpen {
		t.Error("IsAlwaysOpen should be false by default")
	}
	if comp.ServiceType != "shop" {
		t.Errorf("ServiceType = %q, want %q", comp.ServiceType, "shop")
	}

	// Check days: Mon-Sat should be open, Sunday closed
	if comp.DaysOpen[Sunday] {
		t.Error("Sunday should be closed by default")
	}
	for d := Monday; d <= Saturday; d++ {
		if !comp.DaysOpen[d] {
			t.Errorf("%s should be open by default", d.String())
		}
	}
}

func TestNewAlwaysOpenComponent(t *testing.T) {
	comp := NewAlwaysOpenComponent("inn")

	if !comp.IsAlwaysOpen {
		t.Error("IsAlwaysOpen should be true")
	}
	if comp.ServiceType != "inn" {
		t.Errorf("ServiceType = %q, want %q", comp.ServiceType, "inn")
	}
	// All days should be open
	for d := Sunday; d <= Saturday; d++ {
		if !comp.DaysOpen[d] {
			t.Errorf("%s should be open for always-open component", d.String())
		}
	}
}

func TestOperatingHoursComponent_IsOpenAt(t *testing.T) {
	tests := []struct {
		name       string
		openHour   int
		closeHour  int
		daysOpen   [7]bool
		alwaysOpen bool
		checkHour  int
		checkDay   DayOfWeek
		wantOpen   bool
	}{
		{
			name:      "within normal hours on open day",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{false, true, true, true, true, true, true},
			checkHour: 12,
			checkDay:  Monday,
			wantOpen:  true,
		},
		{
			name:      "before opening on open day",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{false, true, true, true, true, true, true},
			checkHour: 6,
			checkDay:  Monday,
			wantOpen:  false,
		},
		{
			name:      "after closing on open day",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{false, true, true, true, true, true, true},
			checkHour: 20,
			checkDay:  Monday,
			wantOpen:  false,
		},
		{
			name:      "within hours on closed day",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{false, true, true, true, true, true, true},
			checkHour: 12,
			checkDay:  Sunday,
			wantOpen:  false,
		},
		{
			name:      "overnight hours - late night",
			openHour:  22,
			closeHour: 6,
			daysOpen:  [7]bool{true, true, true, true, true, true, true},
			checkHour: 23,
			checkDay:  Friday,
			wantOpen:  true,
		},
		{
			name:      "overnight hours - early morning",
			openHour:  22,
			closeHour: 6,
			daysOpen:  [7]bool{true, true, true, true, true, true, true},
			checkHour: 3,
			checkDay:  Saturday,
			wantOpen:  true,
		},
		{
			name:      "overnight hours - midday closed",
			openHour:  22,
			closeHour: 6,
			daysOpen:  [7]bool{true, true, true, true, true, true, true},
			checkHour: 12,
			checkDay:  Saturday,
			wantOpen:  false,
		},
		{
			name:       "always open",
			alwaysOpen: true,
			checkHour:  3,
			checkDay:   Sunday,
			wantOpen:   true,
		},
		{
			name:      "at exact opening time",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{true, true, true, true, true, true, true},
			checkHour: 8,
			checkDay:  Monday,
			wantOpen:  true,
		},
		{
			name:      "at exact closing time",
			openHour:  8,
			closeHour: 18,
			daysOpen:  [7]bool{true, true, true, true, true, true, true},
			checkHour: 18,
			checkDay:  Monday,
			wantOpen:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &OperatingHoursComponent{
				OpenHour:     tt.openHour,
				CloseHour:    tt.closeHour,
				DaysOpen:     tt.daysOpen,
				IsAlwaysOpen: tt.alwaysOpen,
			}
			got := comp.IsOpenAt(tt.checkHour, tt.checkDay)
			if got != tt.wantOpen {
				t.Errorf("IsOpenAt(%d, %s) = %v, want %v",
					tt.checkHour, tt.checkDay.String(), got, tt.wantOpen)
			}
		})
	}
}

func TestOperatingHoursComponent_SetHours(t *testing.T) {
	comp := NewOperatingHoursComponent()
	comp.SetHours(10, 22)

	if comp.OpenHour != 10 {
		t.Errorf("OpenHour = %d, want 10", comp.OpenHour)
	}
	if comp.CloseHour != 22 {
		t.Errorf("CloseHour = %d, want 22", comp.CloseHour)
	}

	// Test clamping
	comp.SetHours(-5, 30)
	if comp.OpenHour != 0 {
		t.Errorf("OpenHour should be clamped to 0, got %d", comp.OpenHour)
	}
	if comp.CloseHour != 23 {
		t.Errorf("CloseHour should be clamped to 23, got %d", comp.CloseHour)
	}
}

func TestOperatingHoursComponent_SetDaysOpen(t *testing.T) {
	comp := NewOperatingHoursComponent()
	weekendOnly := [7]bool{true, false, false, false, false, false, true}
	comp.SetDaysOpen(weekendOnly)

	if !comp.DaysOpen[Sunday] || !comp.DaysOpen[Saturday] {
		t.Error("Weekend days should be open")
	}
	for d := Monday; d <= Friday; d++ {
		if comp.DaysOpen[d] {
			t.Errorf("%s should be closed", d.String())
		}
	}
}

func TestOperatingHoursComponent_SetOpenOnDay(t *testing.T) {
	comp := NewOperatingHoursComponent()

	// Close Saturday
	comp.SetOpenOnDay(Saturday, false)
	if comp.DaysOpen[Saturday] {
		t.Error("Saturday should be closed after SetOpenOnDay(Saturday, false)")
	}

	// Open Sunday
	comp.SetOpenOnDay(Sunday, true)
	if !comp.DaysOpen[Sunday] {
		t.Error("Sunday should be open after SetOpenOnDay(Sunday, true)")
	}
}

func TestOperatingHoursComponent_GetNextOpenTime(t *testing.T) {
	tests := []struct {
		name        string
		comp        *OperatingHoursComponent
		currentHour int
		currentDay  DayOfWeek
		wantHour    int
		wantDay     DayOfWeek
	}{
		{
			name: "next open is today",
			comp: &OperatingHoursComponent{
				OpenHour:  8,
				CloseHour: 18,
				DaysOpen:  [7]bool{false, true, true, true, true, true, true},
			},
			currentHour: 6,
			currentDay:  Monday,
			wantHour:    8,
			wantDay:     Monday,
		},
		{
			name: "next open is tomorrow",
			comp: &OperatingHoursComponent{
				OpenHour:  8,
				CloseHour: 18,
				DaysOpen:  [7]bool{false, true, true, true, true, true, true},
			},
			currentHour: 20,
			currentDay:  Monday,
			wantHour:    8,
			wantDay:     Tuesday,
		},
		{
			name: "Sunday closed, next is Monday",
			comp: &OperatingHoursComponent{
				OpenHour:  8,
				CloseHour: 18,
				DaysOpen:  [7]bool{false, true, true, true, true, true, true},
			},
			currentHour: 12,
			currentDay:  Sunday,
			wantHour:    8,
			wantDay:     Monday,
		},
		{
			name: "always open returns current time",
			comp: &OperatingHoursComponent{
				IsAlwaysOpen: true,
			},
			currentHour: 3,
			currentDay:  Sunday,
			wantHour:    3,
			wantDay:     Sunday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHour, gotDay := tt.comp.GetNextOpenTime(tt.currentHour, tt.currentDay)
			if gotHour != tt.wantHour || gotDay != tt.wantDay {
				t.Errorf("GetNextOpenTime(%d, %s) = (%d, %s), want (%d, %s)",
					tt.currentHour, tt.currentDay.String(),
					gotHour, gotDay.String(),
					tt.wantHour, tt.wantDay.String())
			}
		})
	}
}

func TestOperatingHoursComponent_GetStatusMessage(t *testing.T) {
	comp := NewOperatingHoursComponent()
	comp.ClosedMessage = "We're closed for lunch!"

	if msg := comp.GetStatusMessage(true); msg != "Open" {
		t.Errorf("GetStatusMessage(true) = %q, want %q", msg, "Open")
	}
	if msg := comp.GetStatusMessage(false); msg != "We're closed for lunch!" {
		t.Errorf("GetStatusMessage(false) = %q, want %q", msg, "We're closed for lunch!")
	}

	// Test default closed message
	comp.ClosedMessage = ""
	if msg := comp.GetStatusMessage(false); msg != "Closed" {
		t.Errorf("GetStatusMessage(false) with empty message = %q, want %q", msg, "Closed")
	}
}

func TestOperatingHoursComponent_Serialize(t *testing.T) {
	comp := NewOperatingHoursComponent()
	comp.SetHours(10, 20)
	comp.ServiceType = "blacksmith"
	comp.ClosedMessage = "Gone fishing"

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("Serialize() returned empty data")
	}

	// Deserialize and verify
	restored := &OperatingHoursComponent{}
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if restored.OpenHour != 10 {
		t.Errorf("restored.OpenHour = %d, want 10", restored.OpenHour)
	}
	if restored.CloseHour != 20 {
		t.Errorf("restored.CloseHour = %d, want 20", restored.CloseHour)
	}
	if restored.ServiceType != "blacksmith" {
		t.Errorf("restored.ServiceType = %q, want %q", restored.ServiceType, "blacksmith")
	}
	if restored.ClosedMessage != "Gone fishing" {
		t.Errorf("restored.ClosedMessage = %q, want %q", restored.ClosedMessage, "Gone fishing")
	}
}

func TestDayOfWeek_String(t *testing.T) {
	tests := []struct {
		day  DayOfWeek
		want string
	}{
		{Sunday, "Sunday"},
		{Monday, "Monday"},
		{Tuesday, "Tuesday"},
		{Wednesday, "Wednesday"},
		{Thursday, "Thursday"},
		{Friday, "Friday"},
		{Saturday, "Saturday"},
		{DayOfWeek(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.day.String(); got != tt.want {
			t.Errorf("DayOfWeek(%d).String() = %q, want %q", int(tt.day), got, tt.want)
		}
	}
}

func TestClampHour(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-5, 0},
		{0, 0},
		{12, 12},
		{23, 23},
		{30, 23},
		{100, 23},
	}

	for _, tt := range tests {
		if got := clampHour(tt.input); got != tt.want {
			t.Errorf("clampHour(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
