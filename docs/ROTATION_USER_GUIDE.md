# 360° Rotation & Mouse Aim - User Guide

## Overview

Version 2.0 introduces **full 360° rotation** and **independent aim control**, transforming Venture's combat from 4-directional to dual-stick shooter mechanics. This guide explains the new control scheme and how to use it effectively.

**Status:** Foundation Complete (Phase 10.1)  
**Available In:** Version 2.0 Alpha (Coming January 2026)

---

## What's New?

### Before (Version 1.x)
- **Movement = Facing Direction:** Pressing W makes you face up and move up
- **4 Directions Only:** Can only face right, down, left, or up
- **Attack Direction:** Always attacks in the direction you're moving

### After (Version 2.0)
- **Independent Movement & Aim:** Move with WASD, aim with mouse/touch
- **Full 360° Rotation:** Face and shoot in any direction
- **Strafe Mechanics:** Move one direction while shooting another
- **Precise Aiming:** Mouse cursor controls exactly where you aim

---

## Controls

### Desktop (Mouse + Keyboard)

**Movement:**
- `W` - Move up
- `A` - Move left
- `S` - Move down
- `D` - Move right

**Aiming:**
- `Mouse` - Move cursor to aim direction
  - Your character rotates to face the cursor
  - Attacks fire toward the cursor position
  - Smooth rotation follows cursor movement

**Combat:**
- `Space` - Attack in aim direction
- `1-5` - Cast spells in aim direction

**Example Gameplay:**
1. Press `D` to move right
2. Move mouse cursor to upper-left
3. Character moves right while facing upper-left
4. Press `Space` to shoot upper-left while moving right (strafe)

### Mobile (Touch)

**Movement:**
- **Left Virtual Joystick** (bottom-left of screen)
  - Touch and drag to move in that direction
  - Release to stop moving

**Aiming:**
- **Right Virtual Joystick** (bottom-right of screen)
  - Touch and drag to aim in that direction
  - Your character rotates to face the aim direction
  - Attacks fire in the aim direction

**Combat:**
- **Attack Button** - Touch to attack in aim direction
- **Spell Buttons** - Touch to cast spell in aim direction

**Tips:**
- Both joysticks work simultaneously for true dual-stick control
- Visual indicators show joystick positions
- Adjust joystick opacity in settings if they obstruct view

### Gamepad (Future Support)

**Planned Controls:**
- `Left Stick` - Movement (8-way)
- `Right Stick` - Aiming (360°)
- `A Button` / `RT` - Attack in aim direction

---

## Advanced Techniques

### Strafing

**What Is It:** Moving in one direction while aiming/shooting in another.

**How To Use:**
1. Hold `D` to move right
2. Aim mouse cursor to the left
3. Press `Space` to attack left while moving right

**When To Use:**
- **Kiting:** Retreat from enemies while shooting at them
- **Dodging:** Avoid projectiles while maintaining aim on target
- **Flanking:** Move around cover while keeping aim on enemy

### Circle Strafing

**What Is It:** Moving in a circle around an enemy while maintaining aim at them.

**How To Use:**
1. Stand near an enemy
2. Hold `A` + `W` to circle counter-clockwise
3. Keep mouse cursor on the enemy
4. Attack while circling to avoid their attacks

**Best Against:**
- Slow-moving melee enemies
- Bosses with telegraphed attacks
- Stationary turrets

### Snap Aiming

**What Is It:** Quickly flicking mouse cursor to new target.

**How To Use:**
1. Aim at enemy #1 and attack
2. Quickly move cursor to enemy #2
3. Character smoothly rotates to new target
4. Attack when rotation completes

**Rotation Speed:**
- **Default:** 172°/second (smooth but responsive)
- **Fast Mode:** 286°/second (arcade feel) - _Coming in later build_
- **Instant Mode:** No interpolation (for experts) - _Settings option_

### Precision Aiming

**What Is It:** Using subtle mouse movements for exact aim.

**How To Use:**
1. Get close to target for easier precision
2. Move cursor slowly to fine-tune aim
3. Character rotation follows with smooth interpolation
4. Fire when crosshair is exactly on target

**Best For:**
- Long-range attacks
- Hitting weak points on bosses
- Navigating projectiles through gaps

---

## Mobile-Specific Features

### Auto-Aim Assist

**What Is It:** Subtle aim correction that pulls your aim toward nearby enemies.

**How It Works:**
- When aiming near an enemy (within 100 pixels by default)
- Your aim automatically adjusts toward that enemy
- Correction is partial (30% by default), not a full snap
- Makes touch controls competitive with mouse aiming

**Configuration:**
- **Strength:** 0% (off) to 100% (full snap)
  - 30% (default): Subtle assistance
  - 50%: Moderate help for new players
  - 100%: Strong assistance (recommended for very small screens)
- **Range:** 50-200 pixels
  - 100 (default): Standard range
  - 150: Forgiving mode
  - 200: Very forgiving (for casual play)

**Note:** Auto-aim only works against enemies, not interactive objects or terrain.

### Virtual Joystick Customization

**Settings (Coming in later build):**
- **Opacity:** 30% - 100%
  - Lower = less screen obstruction
  - Higher = easier to see during intense gameplay
- **Size:** Small, Medium, Large
  - Adjust based on screen size and finger size
- **Position:** Fixed corners or draggable
  - Fixed: Always bottom-left/right corners
  - Draggable: Place anywhere on first touch

---

## Settings & Configuration

### Rotation Settings

**Smooth Rotation** (Default: On)
- **On:** Character smoothly rotates to aim direction (172°/s)
- **Off:** Character instantly snaps to aim direction (0ms)

**When To Use:**
- **Smooth On:** Normal gameplay, looks more natural
- **Smooth Off:** Competitive play, instant response, twitchy gameplay

### Mouse Sensitivity

**Coming in later build** - Adjust how fast character follows cursor movement.

### Auto-Aim (Mobile Only)

**Strength:** 0% - 100% (Default: 30%)
**Range:** 50 - 200 pixels (Default: 100)

---

## Gameplay Strategy

### Enemy Types

**Melee Enemies:**
- **Strategy:** Strafe around them in circles while attacking
- **Technique:** Circle strafing keeps you out of their range
- **Aim:** Keep cursor on enemy, let rotation track them

**Ranged Enemies:**
- **Strategy:** Dodge projectiles while returning fire
- **Technique:** Perpendicular movement to dodge, snap aim to counterattack
- **Aim:** Quick flicks between dodge direction and enemy position

**Fast Enemies:**
- **Strategy:** Backpedal while maintaining aim
- **Technique:** Hold `S` to retreat, cursor tracks enemy approaching
- **Aim:** Continuous tracking, fire when aligned

**Bosses:**
- **Strategy:** Orbit at safe distance, focus fire on weak points
- **Technique:** Wide circle strafe, precision aim at specific body parts
- **Aim:** Adjust aim for weak point while maintaining orbit

### Multiplayer Tactics

**Cooperative Play:**
- **Crossfire:** One player distracts, other flanks from different angle
- **Focus Fire:** Multiple players aim at same target from different positions
- **Cover Fire:** Retreating player aims behind while advancing player pushes

**Combat Awareness:**
- Aim independent from movement = check all directions while retreating
- Mouse allows quick 180° turns to check behind
- Strafe sideways to keep both allies and enemies in view

---

## Troubleshooting

### "Character rotation feels sluggish"
**Solution:** Rotation is intentionally smooth (172°/s) for visual quality. To disable:
1. Open Settings
2. Find "Smooth Rotation"
3. Toggle to "Off" for instant rotation

### "I keep walking away from where I'm aiming"
**Solution:** This is correct behavior! Movement (WASD) and aim (mouse) are independent.
- WASD controls where you **move**
- Mouse controls where you **face/aim**
- To move toward aim: Press WASD in same direction as cursor

### "Mobile joysticks are in the way"
**Solution:**
1. Open Settings → Touch Controls
2. Reduce joystick opacity (try 50%)
3. Or adjust joystick size to "Small"

### "Auto-aim is pulling me off target"
**Solution:**
1. Open Settings → Touch Controls → Auto-Aim
2. Reduce strength to 10-15% for subtle assistance
3. Or disable entirely for full manual control

### "Character won't rotate to where I'm aiming"
**Check:**
- Ensure you have RotationComponent (all players do by default)
- Check if smooth rotation is off and you're moving very slowly
- Verify cursor is not outside game window

---

## Performance Notes

**Frame Rate Impact:** <1% (rotation is highly optimized)

**Network Bandwidth:** +2 bytes per entity per update (negligible)

**Mobile Battery:** No significant impact (efficient sprite caching)

---

## Comparison to Other Games

**Similar To:**
- **Enter the Gungeon:** Dual-stick, strafe while shooting
- **Nuclear Throne:** 360° aim, independent movement
- **Hotline Miami:** Mouse aim, instant rotation
- **Binding of Isaac:** Twin-stick shooter mechanics

**Different From:**
- **Classic Zelda:** 4-directional, no independent aim
- **Diablo:** Click-to-move, can't strafe
- **Minecraft:** First-person, different perspective

---

## Feedback & Suggestions

**We want to hear from you!**

- Does rotation feel smooth enough?
- Is auto-aim too strong/weak on mobile?
- Do you prefer smooth or instant rotation?
- Any control scheme improvements?

**Report Issues:**
- GitHub Issues: [opd-ai/venture](https://github.com/opd-ai/venture/issues)
- Include: Platform (desktop/mobile), input method, specific scenario

---

## Coming Soon

### Phase 10.2: Projectile Physics
- Arrows curve with gravity
- Bullets bounce off walls
- Grenades arc based on aim angle

### Phase 10.3: Screen Shake & Impact
- Screen shake intensity based on rotation snap speed
- Camera recoil on attacks
- Dramatic hit feedback

### Phase 11: Advanced Level Design
- Rotation-sensitive puzzles (aim beams at mirrors)
- Reflect projectiles using rotation angle
- Diagonal walls for interesting geometry

---

**Document Version:** 1.0  
**For Game Version:** 2.0 Alpha  
**Last Updated:** October 2025

**Quick Reference Card:**
```
Desktop Controls:
  WASD     = Move (independent)
  Mouse    = Aim direction
  Space    = Attack in aim direction
  1-5      = Spells in aim direction

Mobile Controls:
  Left Joystick  = Move
  Right Joystick = Aim
  Attack Button  = Attack in aim direction

Advanced Techniques:
  Strafe        = Move one way, aim another
  Circle Strafe = Orbit enemy while shooting
  Snap Aim      = Quick cursor flicks between targets
  Kiting        = Retreat while shooting at pursuers
```
