// Package engine provides the schedule component for NPC daily routines.
// ScheduleComponent enables NPCs to follow time-based activities, creating
// a living world where characters work, eat, sleep, and socialize.
package engine

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// ActivityType defines the type of scheduled activity.
type ActivityType string

const (
	// ActivityWork represents working at a job location.
	ActivityWork ActivityType = "work"
	// ActivityEat represents eating at a food location.
	ActivityEat ActivityType = "eat"
	// ActivitySleep represents sleeping at home.
	ActivitySleep ActivityType = "sleep"
	// ActivitySocialize represents social interaction at public places.
	ActivitySocialize ActivityType = "socialize"
	// ActivityPatrol represents guard patrol routes.
	ActivityPatrol ActivityType = "patrol"
	// ActivityIdle represents idle time with no specific activity.
	ActivityIdle ActivityType = "idle"
)

// ScheduledActivity represents a single time-based activity in an NPC's day.
type ScheduledActivity struct {
	// ActivityType is the type of activity (work, eat, sleep, etc.)
	ActivityType ActivityType `json:"activity_type"`
	// StartHour is when the activity begins (0-23)
	StartHour int `json:"start_hour"`
	// EndHour is when the activity ends (0-23)
	EndHour int `json:"end_hour"`
	// LocationX is the target X coordinate for this activity
	LocationX float64 `json:"location_x"`
	// LocationY is the target Y coordinate for this activity
	LocationY float64 `json:"location_y"`
	// LocationName is the human-readable name of the location
	LocationName string `json:"location_name"`
}

// ScheduleComponent tracks an NPC's daily schedule of activities.
// Activities are processed by the ScheduleSystem which moves NPCs
// to appropriate locations based on the current game time.
type ScheduleComponent struct {
	// Activities is the ordered list of daily activities
	Activities []ScheduledActivity `json:"activities"`
	// CurrentActivityIdx is the index of the current activity
	CurrentActivityIdx int `json:"current_activity_idx"`
	// HomeX is the NPC's home X coordinate
	HomeX float64 `json:"home_x"`
	// HomeY is the NPC's home Y coordinate
	HomeY float64 `json:"home_y"`
	// MovementSpeed is how fast the NPC moves between locations
	MovementSpeed float64 `json:"movement_speed"`
	// IsMoving indicates if the NPC is currently traveling
	IsMoving bool `json:"is_moving"`
}

// NewScheduleComponent creates a schedule component with default values.
func NewScheduleComponent(homeX, homeY float64) *ScheduleComponent {
	return &ScheduleComponent{
		Activities:         make([]ScheduledActivity, 0),
		CurrentActivityIdx: 0,
		HomeX:              homeX,
		HomeY:              homeY,
		MovementSpeed:      50.0, // pixels per second
		IsMoving:           false,
	}
}

// Type returns the component type identifier.
func (s *ScheduleComponent) Type() string {
	return "schedule"
}

// AddActivity adds a new scheduled activity.
// Activities should be added in chronological order.
func (s *ScheduleComponent) AddActivity(actType ActivityType, startHour, endHour int, locX, locY float64, locName string) {
	s.Activities = append(s.Activities, ScheduledActivity{
		ActivityType: actType,
		StartHour:    startHour,
		EndHour:      endHour,
		LocationX:    locX,
		LocationY:    locY,
		LocationName: locName,
	})
}

// GetCurrentActivity returns the current scheduled activity, or nil if none.
func (s *ScheduleComponent) GetCurrentActivity() *ScheduledActivity {
	if len(s.Activities) == 0 {
		return nil
	}
	if s.CurrentActivityIdx < 0 || s.CurrentActivityIdx >= len(s.Activities) {
		return nil
	}
	return &s.Activities[s.CurrentActivityIdx]
}

// GetActivityForHour returns the activity that should be active at the given hour.
func (s *ScheduleComponent) GetActivityForHour(hour int) *ScheduledActivity {
	for i := range s.Activities {
		act := &s.Activities[i]
		// Handle activities that wrap around midnight
		if act.StartHour <= act.EndHour {
			if hour >= act.StartHour && hour < act.EndHour {
				return act
			}
		} else {
			// Activity spans midnight (e.g., 22:00 - 06:00)
			if hour >= act.StartHour || hour < act.EndHour {
				return act
			}
		}
	}
	return nil
}

// UpdateActivityIndex updates CurrentActivityIdx based on the given hour.
func (s *ScheduleComponent) UpdateActivityIndex(hour int) {
	for i := range s.Activities {
		act := &s.Activities[i]
		if act.StartHour <= act.EndHour {
			if hour >= act.StartHour && hour < act.EndHour {
				s.CurrentActivityIdx = i
				return
			}
		} else {
			if hour >= act.StartHour || hour < act.EndHour {
				s.CurrentActivityIdx = i
				return
			}
		}
	}
	// No matching activity - default to first or -1
	if len(s.Activities) > 0 {
		s.CurrentActivityIdx = 0
	} else {
		s.CurrentActivityIdx = -1
	}
}

// Serialize encodes the component to bytes for persistence.
func (s *ScheduleComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "schedule",
		"activities":     len(s.Activities),
	}).Debug("Serializing schedule component")

	data, err := json.Marshal(s)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "schedule",
			"error":          err.Error(),
		}).Error("Failed to serialize schedule component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (s *ScheduleComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "schedule",
		"bytes":          len(data),
	}).Debug("Deserializing schedule component")

	if err := json.Unmarshal(data, s); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "schedule",
			"error":          err.Error(),
		}).Error("Failed to deserialize schedule component")
		return err
	}
	return nil
}
