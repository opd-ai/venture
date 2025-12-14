// Package engine provides the schedule system for NPC daily routines.
// ScheduleSystem processes entities with ScheduleComponent, moving them
// between locations based on game time and their daily schedules.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ScheduleSystem processes NPC schedules and movements.
// It uses the game clock to determine current activities and
// moves NPCs to their scheduled locations.
type ScheduleSystem struct {
	world *World
	clock GameClock
}

// NewScheduleSystem creates a new schedule system.
func NewScheduleSystem(world *World, clock GameClock) *ScheduleSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "schedule",
	}).Debug("Creating schedule system")

	return &ScheduleSystem{
		world: world,
		clock: clock,
	}
}

// Update processes all entities with schedule components.
// It updates activity states and moves NPCs toward their destinations.
func (s *ScheduleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.clock == nil {
		return
	}

	currentTime := s.clock.Now()
	currentHour := currentTime.Hour()

	for _, entity := range entities {
		if !entity.HasComponent("schedule") {
			continue
		}

		schedComp, ok := entity.GetComponent("schedule")
		if !ok || schedComp == nil {
			continue
		}
		schedule := schedComp.(*ScheduleComponent)

		// Update current activity based on time
		schedule.UpdateActivityIndex(currentHour)

		// Get current activity
		activity := schedule.GetCurrentActivity()
		if activity == nil {
			continue
		}

		// Get position component
		posComp, ok := entity.GetComponent("position")
		if !ok || posComp == nil {
			continue
		}
		pos := posComp.(*PositionComponent)

		// Calculate distance to target location
		dx := activity.LocationX - pos.X
		dy := activity.LocationY - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If we're already at the location, stop moving
		const arrivalThreshold = 5.0
		if distance < arrivalThreshold {
			schedule.IsMoving = false
			// Clear velocity if present
			if velComp, ok := entity.GetComponent("velocity"); ok && velComp != nil {
				vel := velComp.(*VelocityComponent)
				vel.VX = 0
				vel.VY = 0
			}
			continue
		}

		// Move toward target location
		schedule.IsMoving = true
		moveDistance := schedule.MovementSpeed * deltaTime
		if moveDistance > distance {
			moveDistance = distance
		}

		// Normalize direction and apply movement
		if distance > 0 {
			pos.X += (dx / distance) * moveDistance
			pos.Y += (dy / distance) * moveDistance
		}

		logrus.WithFields(logrus.Fields{
			"system_name":  "schedule",
			"entity_id":    entity.ID,
			"activity":     activity.ActivityType,
			"location":     activity.LocationName,
			"distance":     distance,
			"current_hour": currentHour,
		}).Trace("NPC moving to scheduled location")
	}
}

// GenerateDefaultSchedule creates a deterministic daily schedule for an NPC.
// The schedule is based on the seed and NPC role for reproducibility.
func GenerateDefaultSchedule(seed int64, role string, homeX, homeY, workX, workY float64) *ScheduleComponent {
	rng := rand.New(rand.NewSource(seed))
	schedule := NewScheduleComponent(homeX, homeY)

	// Generate schedule based on role
	switch role {
	case "merchant":
		// Merchants: sleep 22-6, work 8-20, eat 12-13 and 19-20
		schedule.AddActivity(ActivitySleep, 22, 6, homeX, homeY, "Home")
		schedule.AddActivity(ActivityEat, 6, 7, homeX, homeY, "Home")
		schedule.AddActivity(ActivityWork, 8, 12, workX, workY, "Shop")
		schedule.AddActivity(ActivityEat, 12, 13, workX, workY, "Shop")
		schedule.AddActivity(ActivityWork, 13, 19, workX, workY, "Shop")
		schedule.AddActivity(ActivityEat, 19, 20, homeX, homeY, "Home")
		schedule.AddActivity(ActivityIdle, 20, 22, homeX, homeY, "Home")

	case "guard":
		// Guards: 8-hour patrol shifts with breaks
		shiftStart := rng.Intn(3) * 8 // 0, 8, or 16
		schedule.AddActivity(ActivityPatrol, shiftStart, shiftStart+4, workX, workY, "Patrol Route")
		schedule.AddActivity(ActivityEat, shiftStart+4, shiftStart+5, homeX, homeY, "Barracks")
		schedule.AddActivity(ActivityPatrol, shiftStart+5, shiftStart+8, workX, workY, "Patrol Route")
		schedule.AddActivity(ActivitySleep, (shiftStart+10)%24, (shiftStart+18)%24, homeX, homeY, "Barracks")

	case "villager":
		// Villagers: varied schedules with socialization
		wakeHour := 5 + rng.Intn(3) // Wake 5-7
		schedule.AddActivity(ActivitySleep, 22, wakeHour, homeX, homeY, "Home")
		schedule.AddActivity(ActivityEat, wakeHour, wakeHour+1, homeX, homeY, "Home")
		schedule.AddActivity(ActivityWork, wakeHour+1, 12, workX, workY, "Workplace")
		schedule.AddActivity(ActivityEat, 12, 13, workX, workY, "Workplace")
		schedule.AddActivity(ActivityWork, 13, 17, workX, workY, "Workplace")
		// Add socialization based on seed
		if rng.Float64() > 0.5 {
			tavernX := homeX + float64(rng.Intn(200)-100)
			tavernY := homeY + float64(rng.Intn(200)-100)
			schedule.AddActivity(ActivitySocialize, 18, 21, tavernX, tavernY, "Tavern")
		} else {
			schedule.AddActivity(ActivityIdle, 18, 21, homeX, homeY, "Home")
		}
		schedule.AddActivity(ActivityEat, 17, 18, homeX, homeY, "Home")

	default:
		// Default: basic schedule
		schedule.AddActivity(ActivitySleep, 22, 6, homeX, homeY, "Home")
		schedule.AddActivity(ActivityIdle, 6, 22, homeX, homeY, "Home")
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "schedule",
		"role":        role,
		"seed":        seed,
		"activities":  len(schedule.Activities),
	}).Debug("Generated NPC schedule")

	return schedule
}
