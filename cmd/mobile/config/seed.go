package config

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// configError represents a configuration error that occurred during parsing or validation
type configError struct {
	key   string
	value string
	err   error
}

func (e *configError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("config error for %s=%q: %v", e.key, e.value, e.err)
	}
	return fmt.Sprintf("config error for %s=%q", e.key, e.value)
}

// GetSeedFromEnv retrieves the world seed from VENTURE_SEED environment variable.
// Returns an error if the environment variable is set but invalid.
// Falls back to time-based seed if not set (no error).
//
// INTENTIONAL EXCEPTION to Coding Guideline #2 (Deterministic Generation):
// Time-based fallback is a documented exception for mobile UX convenience.
// Players get a unique experience each launch unless they explicitly set
// VENTURE_SEED for reproducible gameplay (e.g., testing, sharing worlds).
//
// For reproducible worlds (bug reports, testing, multiplayer coordination):
// Always set VENTURE_SEED environment variable to a specific value.
func GetSeedFromEnv(logger *logrus.Logger) (int64, error) {
	if seedStr := os.Getenv("VENTURE_SEED"); seedStr != "" {
		seed, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"seedStr": seedStr,
					"error":   err.Error(),
				}).Warn("invalid VENTURE_SEED environment variable, using time-based seed")
			}
			return time.Now().UnixNano(), &configError{
				key:   "VENTURE_SEED",
				value: seedStr,
				err:   err,
			}
		}
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"seed":   seed,
				"source": "VENTURE_SEED",
			}).Info("using seed from environment variable")
		}
		return seed, nil
	}

	seed := time.Now().UnixNano()
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"seed":   seed,
			"source": "time-based",
		}).Info("using time-based seed")
	}
	return seed, nil
}

// GetGenreFromEnv retrieves the genre from VENTURE_GENRE environment variable.
// Returns an error if the environment variable is set but invalid.
// Falls back to random genre selection if not set (no error).
func GetGenreFromEnv(genres []string, rng *rand.Rand, logger *logrus.Logger) (string, error) {
	if genreEnv := os.Getenv("VENTURE_GENRE"); genreEnv != "" {
		// Validate that the genre is in the allowed list
		for _, validGenre := range genres {
			if genreEnv == validGenre {
				if logger != nil {
					logger.WithFields(logrus.Fields{
						"genre":  genreEnv,
						"source": "VENTURE_GENRE",
					}).Info("using genre from environment variable")
				}
				return genreEnv, nil
			}
		}
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"genre": genreEnv,
				"valid": genres,
			}).Warn("invalid VENTURE_GENRE environment variable, using random genre")
		}
		genre := genres[rng.Intn(len(genres))]
		return genre, &configError{
			key:   "VENTURE_GENRE",
			value: genreEnv,
			err:   fmt.Errorf("not in valid list %v", genres),
		}
	}

	genre := genres[rng.Intn(len(genres))]
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"genre":  genre,
			"source": "random",
		}).Info("using random genre")
	}
	return genre, nil
}
