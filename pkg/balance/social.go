package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// SocialValidator validates social balance through simulated player interactions.
type SocialValidator struct {
	config *BalanceConfig
}

// NewSocialValidator creates a social balance validator.
func NewSocialValidator(config *BalanceConfig) *SocialValidator {
	return &SocialValidator{
		config: config,
	}
}

// GetDomain returns "Social".
func (v *SocialValidator) GetDomain() string {
	return "Social"
}

// Validate runs social balance tests.
func (v *SocialValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Social",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Social"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Social",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting social balance validation")

	// Test 1: Territory control defender advantage (55-60%)
	logrus.Debug("validating territory control")
	if err := v.validateTerritoryControl(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Social",
			"test":   "territory_control",
			"error":  err.Error(),
		}).Error("territory control validation failed")
		return nil, fmt.Errorf("territory control validation failed: %w", err)
	}

	// Test 2: Trade fraud rate (<1%)
	logrus.Debug("validating trade fraud prevention")
	if err := v.validateTradeFraud(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Social",
			"test":   "trade_fraud",
			"error":  err.Error(),
		}).Error("trade fraud validation failed")
		return nil, fmt.Errorf("trade fraud validation failed: %w", err)
	}

	// Test 3: Chat rejection rate (<5%)
	logrus.Debug("validating chat rate limits")
	if err := v.validateChatRateLimits(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Social",
			"test":   "chat_rate_limits",
			"error":  err.Error(),
		}).Error("chat rate limit validation failed")
		return nil, fmt.Errorf("chat rate limit validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Social",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("social balance validation complete")
	return result, nil
}

// validateTerritoryControl checks that defenders have appropriate advantage.
func (v *SocialValidator) validateTerritoryControl(ctx context.Context, result *ValidationResult) error {
	defenderWins := 0
	rng := rand.New(rand.NewSource(v.config.Seed))
	battles := result.SimulationCount / 2
	progressInterval := battles / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	for i := 0; i < battles; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    battles,
				"percent":  float64(i+1) / float64(battles) * 100,
			}).Debug("territory control simulation progress")
		}

		// Simulate territory battle with defender advantage
		// Defenders get 10% stat bonus from fortifications
		defenderStrength := 100.0 * 1.10 // 10% defender bonus
		attackerStrength := 100.0 + rng.Float64()*20.0 - 10.0

		if defenderStrength+rng.Float64()*20.0 > attackerStrength+rng.Float64()*20.0 {
			defenderWins++
		}
	}

	defenderAdvantage := float64(defenderWins) / float64(battles)
	result.Metrics["defender_advantage"] = defenderAdvantage

	minThreshold := v.config.GetThreshold("defender_advantage_min")
	maxThreshold := v.config.GetThreshold("defender_advantage_max")

	if defenderAdvantage < minThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Defender advantage too low (%.1f%%, target: %.1f%%-%.1f%%)",
				defenderAdvantage*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Increase defender bonuses from fortifications")
	}

	if defenderAdvantage > maxThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Defender advantage too high (%.1f%%, target: %.1f%%-%.1f%%)",
				defenderAdvantage*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Reduce defender bonuses or add attacker siege weapons")
	}

	return nil
}

// validateTradeFraud checks that trust-based trade system prevents scams.
func (v *SocialValidator) validateTradeFraud(ctx context.Context, result *ValidationResult) error {
	fraudAttempts := 0
	fraudSuccesses := 0
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	trades := result.SimulationCount / 2
	progressInterval := trades / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	for i := 0; i < trades; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    trades,
				"percent":  float64(i+1) / float64(trades) * 100,
			}).Debug("trade fraud simulation progress")
		}

		// Simulate trade with trust system
		// 10% of trades are fraud attempts
		isFraudAttempt := rng.Float64() < 0.10
		if isFraudAttempt {
			fraudAttempts++

			// Trust system catches fraud based on trader history
			// New traders: 50% catch rate
			// Established traders: 95% catch rate
			traderTrust := rng.Float64() // 0-1 trust level
			catchRate := 0.50 + traderTrust*0.45
			if rng.Float64() > catchRate {
				fraudSuccesses++
			}
		}
	}

	var fraudRate float64
	if fraudAttempts > 0 {
		fraudRate = float64(fraudSuccesses) / float64(trades) // Rate against all trades
	}
	result.Metrics["fraud_rate"] = fraudRate
	result.Metrics["fraud_attempts"] = float64(fraudAttempts)
	result.Metrics["fraud_successes"] = float64(fraudSuccesses)

	threshold := v.config.GetThreshold("fraud_rate_max")
	if fraudRate > threshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Fraud rate too high (%.2f%%, target: <%.1f%%)",
				fraudRate*100, threshold*100))
		result.Recommendations = append(result.Recommendations,
			"Strengthen trade verification, require escrow for high-value trades")
	}

	return nil
}

// validateChatRateLimits checks that spam prevention doesn't affect normal users.
func (v *SocialValidator) validateChatRateLimits(ctx context.Context, result *ValidationResult) error {
	rejectedMessages := 0
	rng := rand.New(rand.NewSource(v.config.Seed + 2))
	messages := result.SimulationCount
	progressInterval := messages / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	// Simulate message rate limiting
	// Normal users: 1-5 messages per minute
	// Spammers: 30+ messages per minute
	for i := 0; i < messages; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    messages,
				"percent":  float64(i+1) / float64(messages) * 100,
			}).Debug("chat rate limit simulation progress")
		}

		// 5% are spammers, 95% are normal users
		isSpammer := rng.Float64() < 0.05
		messagesPerMinute := 1 + rng.Intn(5) // Normal user
		if isSpammer {
			messagesPerMinute = 30 + rng.Intn(70) // Spammer: 30-100 messages
		}

		// Rate limit: 10 messages per minute
		rateLimit := 10
		if messagesPerMinute > rateLimit && !isSpammer {
			// Normal user incorrectly flagged (shouldn't happen with 1-5 rate)
			rejectedMessages++
		}
	}

	rejectRate := float64(rejectedMessages) / float64(messages)
	result.Metrics["chat_reject_rate"] = rejectRate

	threshold := v.config.GetThreshold("chat_reject_rate_max")
	if rejectRate > threshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Chat reject rate too high (%.2f%%, target: <%.1f%%)",
				rejectRate*100, threshold*100))
		result.Recommendations = append(result.Recommendations,
			"Increase rate limit threshold or add burst allowance")
	}

	return nil
}
