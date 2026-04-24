//go:build !android && !ios
// +build !android,!ios

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/config"
	"github.com/opd-ai/venture/pkg/hostplay"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

// Note: All system wrappers have been consolidated in system_wrappers.go for better code organization.

var (
	width            = flag.Int("width", 1920, "Screen width (1280, 1920, 2560, 3840)")
	height           = flag.Int("height", 1080, "Screen height (720, 1080, 1440, 2160)")
	fullscreen       = flag.Bool("fullscreen", false, "Start in fullscreen mode")
	seed             = flag.Int64("seed", seededRandom(), "World generation seed")
	genreID          = flag.String("genre", "random", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc, random)")
	weatherType      = flag.String("weather", "", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation) - empty for genre-appropriate random")
	weatherIntensity = flag.String("weather-intensity", "heavy", "Weather intensity (light, medium, heavy, extreme)")

	// Post-processing configuration (all enabled by default for maximum visual quality)
	postprocessPreset          = flag.String("postprocess-preset", "cinematic", "Post-processing preset (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic, neutral, cinematic)")
	postprocessColorGrading    = flag.Bool("postprocess-color-grading", true, "Enable color grading effect")
	postprocessVignette        = flag.Bool("postprocess-vignette", true, "Enable vignette effect")
	postprocessChromaticAber   = flag.Bool("postprocess-chromatic", true, "Enable chromatic aberration effect")
	postprocessSaturation      = flag.Float64("postprocess-saturation", 1.1, "Color grading saturation (0.0-2.0)")
	postprocessContrast        = flag.Float64("postprocess-contrast", 1.05, "Color grading contrast (0.0-2.0)")
	postprocessBrightness      = flag.Float64("postprocess-brightness", 0.02, "Color grading brightness (-1.0 to 1.0)")
	postprocessVignetteIntens  = flag.Float64("postprocess-vignette-intensity", 0.6, "Vignette intensity (0.0-1.0)")
	postprocessVignetteSoft    = flag.Float64("postprocess-vignette-softness", 0.4, "Vignette softness (0.0-1.0)")
	postprocessChromaticIntens = flag.Float64("postprocess-chromatic-intensity", 0.3, "Chromatic aberration intensity (0.0-1.0)")

	// Phase 5.4 (PLAN.md): Genre Palette (enhanced defaults for best visual quality)
	paletteHarmony = flag.String("palette-harmony", "triadic", "Color harmony type (complementary, analogous, triadic, tetradic, split-complementary, monochromatic)")
	paletteMood    = flag.String("palette-mood", "vibrant", "Palette mood (normal, bright, dark, saturated, muted, vibrant, pastel, tense, calm, victorious, melancholic, energetic, mystical, ominous, serene, aggressive, playful, somber, ethereal, dangerous, peaceful, chaotic, regal, desolate)")
	paletteRarity  = flag.String("palette-rarity", "epic", "Palette rarity/intensity (common, uncommon, rare, epic, legendary)")

	verbose       = flag.Bool("verbose", true, "Enable verbose logging")
	profile       = flag.Bool("profile", true, "Enable performance profiling with frame time tracking")
	multiplayer   = flag.Bool("multiplayer", false, "Connect to remote multiplayer server (default: starts local server)")
	server        = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
	highLatency   = flag.Bool("high-latency", false, "Use high-latency configuration optimized for Tor/onion services (200-5000ms latency)")
	hostAndPlay   = flag.Bool("host-and-play", false, "Explicitly enable host-and-play mode (default behavior when --multiplayer not specified)")
	hostLAN       = flag.Bool("host-lan", false, "Bind server to 0.0.0.0 for LAN access instead of 127.0.0.1 (requires host-and-play mode)")
	serverPort    = flag.Int("port", 8080, "Server port for --host-and-play mode (will try next 10 ports if occupied)")
	serverPlayers = flag.Int("max-players", 4, "Maximum players for --host-and-play mode")
	serverTick    = flag.Int("tick-rate", 30, "Server tick rate for --host-and-play mode (updates per second)")
	noTutorial    = flag.Bool("no-tutorial", false, "Disable tutorial for experienced players")
	enableVR      = flag.Bool("vr", false, "Enable VR mode (requires VR headset, auto-detects hardware)")
	forceVR       = flag.Bool("force-vr", false, "Force VR mode even without detected hardware (for testing)")
	showVersion   = flag.Bool("version", false, "Print version information and exit")
	spriteCacheMB = flag.Int("sprite-cache-mb", 0, "Sprite cache size in MB (0 = use platform default: 400 desktop, 150 WASM; max 300)")

	// G7 (AUDIT.md): opt-in client Prometheus/health endpoint for host-and-play observability.
	clientMetricsPort   = flag.String("metrics-port", "9091", "Port for client Prometheus metrics HTTP endpoint (only active when --enable-metrics is set)")
	clientEnableMetrics = flag.Bool("enable-metrics", false, "Start a Prometheus /metrics endpoint on --metrics-port (opt-in; disabled by default for desktop)")
)

// initializeLogger creates and configures the logger based on environment variables and flags.
func initializeLogger() (*logrus.Logger, *logrus.Entry) {
	logConfig := logging.DefaultConfig()

	// Check for JSON format from environment (default to text for client)
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat == "json" {
		logConfig.Format = logging.JSONFormat
	} else {
		logConfig.Format = logging.TextFormat
		logConfig.EnableColor = true
	}

	// Set log level from environment or use Info as default
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		logConfig.Level = logging.LogLevel(logLevel)
	} else if *verbose {
		logConfig.Level = logging.DebugLevel
	} else {
		logConfig.Level = logging.InfoLevel
	}

	logger := logging.NewLogger(logConfig)
	clientLogger := logger.WithFields(logrus.Fields{
		"component": "client",
		"genre":     *genreID,
		"seed":      *seed,
	})

	clientLogger.Infof("Starting Venture %s", version.FullVersion)
	clientLogger.WithFields(logrus.Fields{
		"width":  *width,
		"height": *height,
		"seed":   *seed,
		"genre":  *genreID,
	}).Info("client configuration")

	return logger, clientLogger
}

// initializeNetworkClient sets up network connection for multiplayer.
// With auto-enable logic, host-and-play always sets *multiplayer=true before this is called,
// so this function always creates a network client (either to embedded or remote server).
func initializeNetworkClient(logger *logrus.Logger, clientLogger *logrus.Entry) network.ClientConnection {
	clientLogger.WithField("server", *server).Info("connecting to server")

	var clientConfig network.ClientConfig
	if *highLatency {
		clientConfig = network.TorClientConfig()
		clientLogger.Info("using high-latency client configuration (Tor/onion service optimized)")
	} else {
		clientConfig = network.DefaultClientConfig()
	}
	clientConfig.ServerAddress = *server
	networkClient := network.NewClientWithLogger(clientConfig, logger)

	// Connect to server
	if err := networkClient.Connect(); err != nil {
		clientLogger.WithError(err).Fatal("failed to connect to server")
	}

	clientLogger.Info("connected to server successfully")

	// Handle network errors in background
	go func() {
		for err := range networkClient.ReceiveError() {
			// Check if error is due to normal disconnection (EOF, connection closed)
			if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "use of closed") {
				clientLogger.WithError(err).Debug("connection closed")
			} else {
				clientLogger.WithError(err).Error("network error")
			}
		}
	}()

	return networkClient
}

// canonicalGenreID normalizes genre ID aliases to the canonical form used in map keys.
// The CLI flag uses "postapoc" as the canonical ID; some older code used "postapocalyptic".
func canonicalGenreID(genreID string) string {
	if genreID == "postapocalyptic" {
		return "postapoc"
	}
	return genreID
}

// seededRandom returns a random seed based on time for CLI flag default values.
//
// IMPORTANT: This function is called ONLY at program initialization time via
// flag.Int64("seed", seededRandom(), ...) in the var() block. It is NOT called
// during gameplay. The time-based randomness is appropriate here because:
//
//  1. It provides a unique default seed for each program run when users don't
//     specify --seed explicitly
//  2. Once the seed is chosen (either from this default or from --seed flag),
//     ALL procedural generation uses that deterministic seed
//  3. This is a one-time initialization, not ongoing randomness
//
// The time.Now() usage here is explicitly EXEMPT from the "no time.Now() in
// procgen" rule because it's for CLI initialization, not for procedural content
// generation during gameplay.
func seededRandom() int64 {
	nowNano := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(nowNano))
	return rng.Int63()
}

// getGenreTheme returns the genre theme configuration for all generators.
// Uses the world seed for deterministic random genre selection.
func getGenreTheme() *genre.Genre {
	return genre.GetThemeWithSeed(*genreID, *seed)
}

// startEmbeddedServer starts a server in a background goroutine for host-and-play mode
// Design: Uses ServerManager from pkg/hostplay for lifecycle management
// Why: Reuses server implementation with proper resource management
//
// Returns: (serverAddress, cleanupFunction, error)
func startEmbeddedServer(logger *logrus.Logger, seed int64, genreID string) (string, func(), error) {
	serverLogger := logger.WithFields(logrus.Fields{
		"component": "embedded-server",
		"seed":      seed,
		"genre":     genreID,
	})

	serverLogger.Info("starting server in background")

	// Create server configuration
	serverConfig := &hostplay.ServerConfig{
		Port:       *serverPort,
		MaxPlayers: *serverPlayers,
		BindLAN:    *hostLAN,
		WorldSeed:  seed,
		GenreID:    genreID,
		Difficulty: 0.5,
		TickRate:   *serverTick,
	}

	// Create server manager
	manager, err := hostplay.NewServerManager(serverConfig, logger)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create server manager: %w", err)
	}

	// Start the server (blocks until listening or error)
	if err := manager.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	serverAddr := manager.Address()
	port := manager.Port()

	if *hostLAN {
		serverLogger.WithField("bindAddr", "0.0.0.0").Warn("server accessible on LAN - firewall may block connections")

		// Try to get LAN IP for display
		if lanAddr := manager.GetLANAddress(); lanAddr != "" {
			serverLogger.WithField("lanAddress", lanAddr).Info("LAN players can connect to this address")
		}
	} else {
		serverLogger.WithField("bindAddr", "127.0.0.1").Info("server bound to localhost only")
	}

	serverLogger.WithFields(logrus.Fields{
		"address":    serverAddr,
		"port":       port,
		"maxPlayers": *serverPlayers,
		"tickRate":   *serverTick,
	}).Info("server ready for connections")

	// Return cleanup function
	cleanup := func() {
		serverLogger.Info("initiating graceful shutdown")
		if err := manager.Stop(); err != nil {
			serverLogger.WithError(err).Error("shutdown error")
		}
	}

	return serverAddr, cleanup, nil
}

// parsePaletteOptions parses command-line flags into palette.GenerationOptions (Phase 5.4).
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (36) is intentional—exhaustively
// validates and maps CLI flag strings to typed enum values for harmony (6), mood (24),
// and rarity (5) options. The switch provides compile-time enum coverage and clear error messages.
func parsePaletteOptions() (*palette.GenerationOptions, error) {
	opts := &palette.GenerationOptions{}

	// Parse harmony type
	switch *paletteHarmony {
	case "complementary":
		opts.Harmony = palette.HarmonyComplementary
	case "analogous":
		opts.Harmony = palette.HarmonyAnalogous
	case "triadic":
		opts.Harmony = palette.HarmonyTriadic
	case "tetradic":
		opts.Harmony = palette.HarmonyTetradic
	case "split-complementary":
		opts.Harmony = palette.HarmonySplitComplementary
	case "monochromatic":
		opts.Harmony = palette.HarmonyMonochromatic
	default:
		return nil, fmt.Errorf("invalid harmony type: %s", *paletteHarmony)
	}

	// Parse mood type
	switch *paletteMood {
	case "normal":
		opts.Mood = palette.MoodNormal
	case "bright":
		opts.Mood = palette.MoodBright
	case "dark":
		opts.Mood = palette.MoodDark
	case "saturated":
		opts.Mood = palette.MoodSaturated
	case "muted":
		opts.Mood = palette.MoodMuted
	case "vibrant":
		opts.Mood = palette.MoodVibrant
	case "pastel":
		opts.Mood = palette.MoodPastel
	case "tense":
		opts.Mood = palette.MoodTense
	case "calm":
		opts.Mood = palette.MoodCalm
	case "victorious":
		opts.Mood = palette.MoodVictorious
	case "melancholic":
		opts.Mood = palette.MoodMelancholic
	case "energetic":
		opts.Mood = palette.MoodEnergetic
	case "mystical":
		opts.Mood = palette.MoodMystical
	case "ominous":
		opts.Mood = palette.MoodOminous
	case "serene":
		opts.Mood = palette.MoodSerene
	case "aggressive":
		opts.Mood = palette.MoodAggressive
	case "playful":
		opts.Mood = palette.MoodPlayful
	case "somber":
		opts.Mood = palette.MoodSomber
	case "ethereal":
		opts.Mood = palette.MoodEthereal
	case "dangerous":
		opts.Mood = palette.MoodDangerous
	case "peaceful":
		opts.Mood = palette.MoodPeaceful
	case "chaotic":
		opts.Mood = palette.MoodChaotic
	case "regal":
		opts.Mood = palette.MoodRegal
	case "desolate":
		opts.Mood = palette.MoodDesolate
	default:
		return nil, fmt.Errorf("invalid mood type: %s", *paletteMood)
	}

	// Parse rarity type
	switch *paletteRarity {
	case "common":
		opts.Rarity = palette.RarityCommon
	case "uncommon":
		opts.Rarity = palette.RarityUncommon
	case "rare":
		opts.Rarity = palette.RarityRare
	case "epic":
		opts.Rarity = palette.RarityEpic
	case "legendary":
		opts.Rarity = palette.RarityLegendary
	default:
		return nil, fmt.Errorf("invalid rarity type: %s", *paletteRarity)
	}

	// Default minimum colors (can be extended later with flag)
	opts.MinColors = 12

	return opts, nil
}

// validateClientConfiguration validates client configuration flags.
// Returns an error if any configuration is invalid, with a helpful message.
func validateClientConfiguration() error {
	validator := config.NewValidator()

	// Build configuration to validate
	cfg := &config.Config{
		Genre: *genreID,
	}

	// Validate host-and-play server configuration if enabled
	if *hostAndPlay {
		cfg.Port = fmt.Sprintf("%d", *serverPort)
		cfg.MaxPlayers = *serverPlayers
		cfg.ValidateMaxPlayers = true
		cfg.TickRate = *serverTick
		cfg.ValidateTickRate = true
	}

	if err := validator.ValidateAll(cfg); err != nil {
		return fmt.Errorf("%w\n\nRun with -help to see all available options", err)
	}

	return nil
}
