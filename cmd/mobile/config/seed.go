package config

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// GetSeedFromEnv retrieves the world seed from VENTURE_SEED environment variable.
// Falls back to time-based seed if not set or invalid.
func GetSeedFromEnv(logger *logrus.Logger) int64 {
	if seedStr := os.Getenv("VENTURE_SEED"); seedStr != "" {
		if seed, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"seed":   seed,
					"source": "VENTURE_SEED",
				}).Info("using seed from environment variable")
			}
			return seed
		}
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"seedStr": seedStr,
			}).Warn("invalid VENTURE_SEED environment variable, using time-based seed")
		}
	}

	seed := time.Now().UnixNano()
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"seed":   seed,
			"source": "time-based",
		}).Info("using time-based seed")
	}
	return seed
}

// GetGenreFromEnv retrieves the genre from VENTURE_GENRE environment variable.
// Falls back to random genre selection if not set or invalid.
func GetGenreFromEnv(genres []string, rng *rand.Rand, logger *logrus.Logger) string {
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
				return genreEnv
			}
		}
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"genre": genreEnv,
				"valid": genres,
			}).Warn("invalid VENTURE_GENRE environment variable, using random genre")
		}
	}

	genre := genres[rng.Intn(len(genres))]
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"genre":  genre,
			"source": "random",
		}).Info("using random genre")
	}
	return genre
}
