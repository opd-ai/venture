// Package engine provides the core game engine functionality including the Entity-Component-System (ECS)
// framework, game loop management, and fundamental game mechanics.
//
// # Architecture Overview
//
// The engine follows an Entity-Component-System (ECS) architectural pattern for maximum
// flexibility, performance, and maintainability:
//
//   - Entities: Lightweight containers with unique IDs and component collections
//   - Components: Pure data structures (no behavior) describing entity properties
//   - Systems: Pure logic (stateless or minimal state) operating on entities with specific components
//
// This separation enables data-oriented design, easy testing, and efficient caching.
//
// # Core Concepts
//
// ## World
//
// The World is the central ECS container managing entities and systems. It maintains
// the entity registry, system execution order, and provides entity queries.
//
//	world := engine.NewWorld()
//	world.AddSystem(engine.NewMovementSystem(200.0))  // Max speed 200
//	world.AddSystem(engine.NewCollisionSystem(32.0))  // Cell size 32
//	world.Update(0.016)  // 60 FPS delta time
//
// ## Entities
//
// Entities are identified by unique IDs and contain components. They have no behavior,
// only data via their component collections.
//
//	player := engine.NewEntity(1)
//	player.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
//	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
//
// ## Components
//
// Components are pure data structures with a Type() method for identification.
// They should never contain logic, only data fields and simple getters.
//
//	type PositionComponent struct {
//	    X, Y float64
//	}
//	
//	func (p PositionComponent) Type() string { return "position" }
//
// ## Systems
//
// Systems implement game logic by operating on entities with specific components.
// They implement the System interface with an Update(deltaTime float64) method.
//
//	type MovementSystem struct {
//	    world    *World
//	    maxSpeed float64
//	}
//	
//	func (s *MovementSystem) Update(deltaTime float64) {
//	    entities := s.world.GetEntitiesWith("position", "velocity")
//	    for _, entity := range entities {
//	        // Update logic here
//	    }
//	}
//
// # System Categories
//
// The engine organizes systems into logical categories:
//
// ## Core Systems
//
//   - InputSystem: Processes player input (keyboard, mouse, touch)
//   - MovementSystem: Updates entity positions based on velocity
//   - CollisionSystem: Detects and resolves entity collisions using spatial partitioning
//   - RenderSystem: Renders entities to screen (sprites, animations, effects)
//
// ## Combat Systems
//
//   - CombatSystem: Handles attack logic, damage calculation, hit detection
//   - HealthSystem: Manages health, damage application, death/revival
//   - SpellCastingSystem: Manages spell casting, mana costs, cooldowns
//   - ManaRegenSystem: Regenerates mana over time
//
// ## Progression Systems
//
//   - ProgressionSystem: Tracks XP, levels, stat increases
//   - SkillProgressionSystem: Manages skill trees, unlocks, progression
//   - QuestTrackerSystem: Tracks quest progress, objective completion
//   - ObjectiveTrackerSystem: Updates quest objectives based on player actions
//
// ## AI Systems
//
//   - AISystem: Controls NPC/enemy behavior (patrol, chase, attack, flee)
//   - SquadSystem: Coordinates enemy groups for tactical behavior
//   - FactionSystem: Manages NPC allegiances and reputation
//
// ## Inventory & Economy
//
//   - InventorySystem: Manages item storage, equipment slots
//   - ItemPickupSystem: Handles item collection from world
//   - CommerceSystem: Merchant trading, buy/sell transactions
//   - CraftingSystem: Item crafting, recipe management
//
// ## UI Systems
//
//   - TutorialSystem: Displays step-by-step tutorial for new players
//   - HelpSystem: Context-sensitive help and control reference
//   - InventoryUI: Inventory management interface
//   - CharacterUI: Character stats and equipment display
//   - QuestUI: Quest log and progress tracking
//   - MapUI: World map with fog of war exploration
//   - SkillsUI: Skill tree interface
//   - CraftingUI: Crafting interface with recipe selection
//   - ShopUI: Merchant trading interface
//
// ## Visual Systems
//
//   - AnimationSystem: Updates sprite animations based on state
//   - VisualFeedbackSystem: Hit flashes, damage tints, status effects
//   - EquipmentVisualSystem: Updates equipment rendering layers
//   - ParticleSystem: Particle effects for spells, impacts, environment
//
// ## Audio Systems
//
//   - AudioManagerSystem: Coordinates sound playback (music, SFX, ambient)
//
// ## Interaction Systems
//
//   - InteractionSystem: Handles puzzle interactions, switches, doors
//   - DialogSystem: NPC conversation management
//
// # Key Interfaces
//
// ## System Interface
//
// All systems must implement this interface to participate in the game loop:
//
//	type System interface {
//	    Update(deltaTime float64)
//	}
//
// ## Component Interface
//
// All components must implement this interface for type identification:
//
//	type Component interface {
//	    Type() string
//	}
//
// ## UISystem Interface
//
// UI systems implement additional methods for visibility control:
//
//	type UISystem interface {
//	    System
//	    IsActive() bool
//	    SetActive(active bool)
//	    Draw(screen interface{})
//	    Update(entities []*Entity, deltaTime float64)
//	}
//
// # Usage Examples
//
// ## Basic Game Loop
//
//	world := engine.NewWorld()
//	world.AddSystem(engine.NewMovementSystem(200.0))
//	world.AddSystem(engine.NewCollisionSystem(32.0))
//	
//	player := world.CreateEntity()
//	player.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
//	player.AddComponent(&engine.VelocityComponent{VX: 10, VY: 0})
//	world.AddEntity(player)
//	
//	// Game loop
//	for {
//	    deltaTime := calculateDeltaTime()
//	    world.Update(deltaTime)
//	}
//
// ## Entity Queries
//
//	// Get all entities with position and health
//	livingEntities := world.GetEntitiesWith("position", "health")
//	
//	// Get specific entity by ID
//	player, exists := world.GetEntity(1)
//	if exists {
//	    pos, ok := player.GetComponent("position")
//	    if ok {
//	        position := pos.(*engine.PositionComponent)
//	        fmt.Printf("Player at %.2f, %.2f\n", position.X, position.Y)
//	    }
//	}
//
// ## Creating Custom Components
//
//	type StaminaComponent struct {
//	    Current float64
//	    Max     float64
//	}
//	
//	func (s StaminaComponent) Type() string { return "stamina" }
//	
//	// Use in entity
//	player.AddComponent(&StaminaComponent{Current: 100, Max: 100})
//
// ## Creating Custom Systems
//
//	type StaminaRegenSystem struct {
//	    world    *World
//	    regenRate float64  // per second
//	}
//	
//	func NewStaminaRegenSystem(world *World, rate float64) *StaminaRegenSystem {
//	    return &StaminaRegenSystem{world: world, regenRate: rate}
//	}
//	
//	func (s *StaminaRegenSystem) Update(deltaTime float64) {
//	    entities := s.world.GetEntitiesWith("stamina")
//	    for _, entity := range entities {
//	        comp, ok := entity.GetComponent("stamina")
//	        if !ok {
//	            continue
//	        }
//	        stamina := comp.(*StaminaComponent)
//	        stamina.Current += s.regenRate * deltaTime
//	        if stamina.Current > stamina.Max {
//	            stamina.Current = stamina.Max
//	        }
//	    }
//	}
//
// # Performance Considerations
//
// ## Entity Queries
//
// Entity queries use component filtering for efficient lookups:
//
//   - GetEntitiesWith() returns cached results when possible
//   - Queries are O(n) where n is total entities (optimized with early exit)
//   - Use specific component queries rather than iterating all entities
//
// ## Spatial Partitioning
//
// The CollisionSystem uses spatial partitioning (quadtree) for efficient collision detection:
//
//   - Reduces collision checks from O(n²) to O(n log n)
//   - Cell size parameter (default 32.0) affects granularity
//   - Smaller cells = more spatial queries, fewer collision checks per cell
//
// ## Component Access
//
// Component lookups are O(1) hash map operations:
//
//   - Cache component references when possible (within a single frame)
//   - Avoid repeated GetComponent() calls for same component
//   - Use type assertions efficiently (check ok before asserting)
//
// # Testing
//
// The engine package achieves 50.0% test coverage. Testing uses interface-based
// dependency injection with stub implementations (StubInput, StubSprite, etc.) to
// avoid Ebiten runtime dependencies in CI environments.
//
// Example test pattern:
//
//	func TestMovementSystem(t *testing.T) {
//	    world := engine.NewWorld()
//	    system := engine.NewMovementSystem(200.0)
//	    world.AddSystem(system)
//	    
//	    entity := world.CreateEntity()
//	    entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
//	    entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 0})
//	    world.AddEntity(entity)
//	    
//	    world.Update(1.0)  // 1 second
//	    
//	    pos, _ := entity.GetComponent("position")
//	    position := pos.(*engine.PositionComponent)
//	    if position.X != 10.0 {
//	        t.Errorf("Expected X=10.0, got %f", position.X)
//	    }
//	}
//
// # References
//
// For more information:
//   - Architecture: docs/ARCHITECTURE.md
//   - API Reference: docs/API_REFERENCE.md  
//   - Development Guide: docs/DEVELOPMENT.md
//   - Technical Spec: docs/TECHNICAL_SPEC.md
package engine
