package balance

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// VehicleValidator validates vehicle balance through simulated trips.
type VehicleValidator struct {
	config *BalanceConfig
}

// NewVehicleValidator creates a vehicle balance validator.
func NewVehicleValidator(config *BalanceConfig) *VehicleValidator {
	return &VehicleValidator{
		config: config,
	}
}

// GetDomain returns "Vehicle".
func (v *VehicleValidator) GetDomain() string {
	return "Vehicle"
}

// Validate runs vehicle balance tests.
func (v *VehicleValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Vehicle",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Vehicle"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Vehicle",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting vehicle balance validation")

	// Test 1: Speed/durability trade-offs are meaningful
	logrus.Debug("validating speed/durability trade-offs")
	if err := v.validateSpeedDurabilityTradeoff(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Vehicle",
			"test":   "speed_durability",
			"error":  err.Error(),
		}).Error("speed/durability validation failed")
		return nil, fmt.Errorf("speed/durability validation failed: %w", err)
	}

	// Test 2: Fuel consumption is balanced
	logrus.Debug("validating fuel consumption")
	if err := v.validateFuelConsumption(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Vehicle",
			"test":   "fuel_consumption",
			"error":  err.Error(),
		}).Error("fuel consumption validation failed")
		return nil, fmt.Errorf("fuel consumption validation failed: %w", err)
	}

	// Test 3: Terrain compatibility creates interesting choices
	logrus.Debug("validating terrain compatibility")
	if err := v.validateTerrainCompatibility(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Vehicle",
			"test":   "terrain_compatibility",
			"error":  err.Error(),
		}).Error("terrain compatibility validation failed")
		return nil, fmt.Errorf("terrain compatibility validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Vehicle",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("vehicle balance validation complete")
	return result, nil
}

// VehicleType represents different vehicle categories.
type VehicleType struct {
	Name       string
	Speed      float64 // 0-100
	Durability float64 // 0-100
	FuelRate   float64 // fuel per distance unit
}

// validateSpeedDurabilityTradeoff checks that fast vehicles are fragile and vice versa.
func (v *VehicleValidator) validateSpeedDurabilityTradeoff(ctx context.Context, result *ValidationResult) error {
	vehicles := []VehicleType{
		{Name: "Scout", Speed: 90, Durability: 30, FuelRate: 0.8},
		{Name: "Transport", Speed: 50, Durability: 80, FuelRate: 1.2},
		{Name: "Battle", Speed: 60, Durability: 70, FuelRate: 1.5},
		{Name: "Racing", Speed: 100, Durability: 20, FuelRate: 0.6},
		{Name: "Heavy", Speed: 30, Durability: 100, FuelRate: 2.0},
	}

	speeds := make([]float64, len(vehicles))
	durabilities := make([]float64, len(vehicles))

	for i, veh := range vehicles {
		speeds[i] = veh.Speed
		durabilities[i] = veh.Durability
		result.Metrics[fmt.Sprintf("vehicle_%s_speed", veh.Name)] = veh.Speed
		result.Metrics[fmt.Sprintf("vehicle_%s_durability", veh.Name)] = veh.Durability
	}

	// Calculate negative correlation (speed and durability should be inversely related)
	correlation := v.calculateCorrelation(speeds, durabilities)
	result.Metrics["speed_durability_correlation"] = correlation

	// Target: negative correlation (-0.5 to -0.9) for meaningful trade-offs
	if correlation > -0.3 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Speed/durability trade-off weak (r=%.2f, target: <-0.3)",
				correlation))
		result.Recommendations = append(result.Recommendations,
			"Increase stat differentiation between vehicle types")
	}

	return nil
}

// validateFuelConsumption checks that fuel costs are balanced.
func (v *VehicleValidator) validateFuelConsumption(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed))
	trips := result.SimulationCount
	progressInterval := trips / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	completedTrips := 0
	totalFuelCost := 0.0

	for i := 0; i < trips; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    trips,
				"percent":  float64(i+1) / float64(trips) * 100,
			}).Debug("fuel consumption simulation progress")
		}

		// Simulate trip with random vehicle and distance
		fuelRate := 0.6 + rng.Float64()*1.4   // 0.6-2.0
		distance := 100.0 + rng.Float64()*900 // 100-1000 units
		fuelCapacity := 100.0 + rng.Float64()*100.0

		fuelNeeded := distance * fuelRate
		fuelCost := fuelNeeded * 10.0 // 10 gold per fuel unit
		totalFuelCost += fuelCost

		// Can complete trip without refuel?
		if fuelNeeded <= fuelCapacity {
			completedTrips++
		}
	}

	completionRate := float64(completedTrips) / float64(trips)
	avgFuelCost := totalFuelCost / float64(trips)

	result.Metrics["trip_completion_rate"] = completionRate
	result.Metrics["avg_fuel_cost"] = avgFuelCost

	// Target: 60-80% of trips completable without refuel
	if completionRate < 0.50 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Fuel consumption too high (%.1f%% trips complete, target: ≥50%%)",
				completionRate*100))
		result.Recommendations = append(result.Recommendations,
			"Increase vehicle fuel capacity or reduce consumption rates")
	}

	if completionRate > 0.95 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Fuel consumption too low (%.1f%% trips complete, target: ≤95%%)",
				completionRate*100))
		result.Recommendations = append(result.Recommendations,
			"Fuel management should be a meaningful consideration")
	}

	return nil
}

// validateTerrainCompatibility checks that terrain affects vehicle choice.
func (v *VehicleValidator) validateTerrainCompatibility(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	terrainTests := result.SimulationCount
	progressInterval := terrainTests / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	terrainTypes := []string{"Road", "Grass", "Sand", "Snow", "Water"}
	vehicleTypes := []string{"Scout", "Transport", "Battle", "Amphibious", "All-Terrain"}

	// Compatibility matrix (vehicle x terrain effectiveness 0-1)
	compatibility := map[string]map[string]float64{
		"Scout":       {"Road": 1.0, "Grass": 0.8, "Sand": 0.4, "Snow": 0.3, "Water": 0.0},
		"Transport":   {"Road": 1.0, "Grass": 0.6, "Sand": 0.3, "Snow": 0.2, "Water": 0.0},
		"Battle":      {"Road": 0.9, "Grass": 0.8, "Sand": 0.5, "Snow": 0.4, "Water": 0.0},
		"Amphibious":  {"Road": 0.7, "Grass": 0.6, "Sand": 0.5, "Snow": 0.3, "Water": 0.9},
		"All-Terrain": {"Road": 0.7, "Grass": 0.8, "Sand": 0.7, "Snow": 0.6, "Water": 0.3},
	}

	bestChoiceCount := 0
	for i := 0; i < terrainTests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    terrainTests,
				"percent":  float64(i+1) / float64(terrainTests) * 100,
			}).Debug("terrain compatibility simulation progress")
		}

		// Random terrain and vehicle
		terrain := terrainTypes[rng.Intn(len(terrainTypes))]
		vehicle := vehicleTypes[rng.Intn(len(vehicleTypes))]

		// Was this a good choice?
		effectiveness := compatibility[vehicle][terrain]

		// Find best vehicle for this terrain
		bestEffectiveness := 0.0
		for _, vt := range vehicleTypes {
			if compatibility[vt][terrain] > bestEffectiveness {
				bestEffectiveness = compatibility[vt][terrain]
			}
		}

		// If player picked a vehicle within 20% of optimal, count as good choice
		if effectiveness >= bestEffectiveness*0.8 {
			bestChoiceCount++
		}
	}

	goodChoiceRate := float64(bestChoiceCount) / float64(terrainTests)
	result.Metrics["terrain_choice_rate"] = goodChoiceRate

	// With 5 vehicles and 5 terrains, random choice has ~25% optimal rate
	// Target: terrain should matter but not be punishing (30-60% good choices)
	if goodChoiceRate > 0.80 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Terrain compatibility too uniform (%.1f%% optimal, target: 30%%-80%%)",
				goodChoiceRate*100))
		result.Recommendations = append(result.Recommendations,
			"Increase terrain penalties to make vehicle choice matter")
	}

	return nil
}

// calculateCorrelation computes the Pearson product-moment correlation
// coefficient (r) between two data series. Pearson's r measures the
// strength and direction of the linear relationship between x and y.
//
// Formula: r = Σ(xᵢ - x̄)(yᵢ - ȳ) / √(Σ(xᵢ - x̄)² · Σ(yᵢ - ȳ)²)
//
// Returns a value in [-1, 1] where:
//
//	 1 = perfect positive linear correlation
//	 0 = no linear correlation
//	-1 = perfect negative linear correlation
//
// Returns 0.0 if fewer than 2 data points are provided, if x and y
// have different lengths, or if either series has zero variance.
func (v *VehicleValidator) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	meanX, meanY := 0.0, 0.0
	for i := range x {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(len(x))
	meanY /= float64(len(y))

	var numerator, denomX, denomY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}

	if denomX == 0 || denomY == 0 {
		return 0.0
	}

	return numerator / math.Sqrt(denomX*denomY)
}
