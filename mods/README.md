# Mods Directory

This directory contains example mod configuration files for Venture.

## Mod Format

Each mod is a JSON file with the following structure:

```json
{
  "id": "unique-mod-id",
  "name": "Human-Readable Name",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "What this mod does",
  "type": "rule",
  "rules": {
    "rule.name": "value"
  },
  "enabled": true
}
```

## Available Rules

### Combat Rules
- `combat.enemy_health_multiplier` (float) - Enemy health scaling
- `combat.player_damage_multiplier` (float) - Player damage scaling
- `difficulty.multiplier` (float) - Overall difficulty scaling

### Spawning Rules
- `spawn.rate_multiplier` (float) - Enemy spawn rate
- `loot.drop_multiplier` (float) - Loot drop rate

### Player Rules
- `player.permadeath` (bool) - Enable permadeath mode

### Territory Rules
- `territory.capture_time_seconds` (int) - Base time to capture territory (default: 60)
- `territory.defender_time_bonus_seconds` (int) - Additional time per defender (default: 30)
- `territory.resource_bonus_percent` (int) - Resource bonus per territory (default: 10)
- `territory.xp_bonus_percent` (int) - XP bonus per territory (default: 5)
- `territory.war_declaration_cost` (int) - Gold to declare war (default: 1000)
- `territory.peace_declaration_cost` (int) - Gold to declare peace (default: 500)
- `territory.war_duration_days` (int) - War duration in days (default: 7)

## Example Mods

- `hardcore-mode.json` - Increased difficulty with permadeath
- `fast-sieges.json` - Reduced capture time and war costs
- `pvp-zones.json` - PvP zone configuration
- `custom-spawns.json` - Custom spawn point configuration

## Enabling Mods

Set `"enabled": true` in the mod file to activate it. Mods are loaded when the game starts with the `--enable-mods` flag (enabled by default).

## Server Configuration

Mods can be applied server-side to affect all connected players. Use `SetConfig()` on the territory manager to apply territory rule changes programmatically.
