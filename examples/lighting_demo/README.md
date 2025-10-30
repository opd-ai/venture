# Lighting Demo

This example demonstrates the dynamic lighting system in Venture.

## Features Demonstrated

- **Player Torch**: Flickering torch light that follows the player
- **Stationary Torches**: Four corner torches with varying flicker patterns
- **Magical Crystals**: Pulsing colored lights (magenta, cyan, yellow, green)
- **Moving Spell**: Light attached to moving projectile
- **Genre Presets**: Different ambient lighting based on genre
- **Light Culling**: Only processes lights visible on screen
- **Light Limits**: Respects maximum lights per frame

## Usage

```bash
# Run with default (fantasy) genre
go run ./examples/lighting_demo

# Run with different genres
go run ./examples/lighting_demo -genre sci-fi
go run ./examples/lighting_demo -genre horror
go run ./examples/lighting_demo -genre cyberpunk
go run ./examples/lighting_demo -genre post-apocalyptic

# Run with lighting disabled (comparison)
go run ./examples/lighting_demo -no-lighting
```

## Controls

- **WASD / Arrow Keys**: Move player (torch follows)
- **Ctrl+L**: Toggle lighting system on/off
- **Ctrl+P**: Pause animation
- **ESC**: Quit

## What to Observe

1. **Flickering**: Torches flicker independently with different speeds
2. **Pulsing**: Crystals pulse smoothly at different rates
3. **Movement**: Spell projectile's light follows it as it bounces
4. **Intensity**: Top-left shows light intensity at player position
5. **Culling**: Light count shows only visible lights (move to see change)
6. **Genre Differences**: Try different genres to see ambient variations

## Genre Characteristics

- **Fantasy**: Warm tones, moderate ambient (0.4), torches feel medieval
- **Sci-Fi**: Cool blue tint, moderate ambient (0.35), clean lighting
- **Horror**: Very dark (0.15), cold tones, oppressive atmosphere
- **Cyberpunk**: Dark (0.25) with purple tint, neon feel
- **Post-Apocalyptic**: Dusty (0.3), harsh yellow tones

## Implementation Notes

This demo shows the basic lighting system. Full integration would include:
- Lights spawned from terrain generation
- Spell effects automatically creating lights
- Enemy flashlights and glowing effects
- Environmental lights (campfires, streetlights, etc.)

## Performance

The demo maintains 60 FPS with 9 lights (1 player + 4 torches + 4 crystals).
With the default limit of 16 lights, the system can handle much more complex scenes.
