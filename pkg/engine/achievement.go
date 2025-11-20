package engine

// AchievementType identifies different types of achievements
type AchievementType int

const (
	// Expression-related achievements
	AchievementFirstExpression  AchievementType = iota
	AchievementExpressionMaster                 // Use all 12 expressions
	AchievementComboStarter                     // Start first combo
	AchievementComboExpert                      // Participate in 10 combos
	AchievementComboLegend                      // Participate in 100 combos
	AchievementSocialButterfly                  // Use 50 different expressions
	AchievementRareExpression                   // Use a rare expression variant
	AchievementGroupPerformer                   // Combo with 5+ people
)

// String returns the achievement name
func (a AchievementType) String() string {
	names := []string{
		"First Expression",
		"Expression Master",
		"Combo Starter",
		"Combo Expert",
		"Combo Legend",
		"Social Butterfly",
		"Rare Expression",
		"Group Performer",
	}
	if int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

// Achievement represents an unlocked achievement
type Achievement struct {
	Type        AchievementType
	UnlockedAt  int64  // Unix timestamp
	Description string // Achievement description
}

// AchievementComponent tracks unlocked achievements for an entity
type AchievementComponent struct {
	Achievements     []Achievement
	ExpressionCount  int                    // Total expressions performed
	UniqueExpression map[ExpressionType]int // Count per expression type
}

// Type returns the component type
func (a AchievementComponent) Type() string {
	return "achievement"
}

// Serialize converts AchievementComponent to bytes for network transmission.
// Format: achievementCount(4) + achievements(N*9) + expressionCount(4) + uniqueCount(4) + uniqueExprs(N*5)
func (a *AchievementComponent) Serialize() []byte {
	// Calculate size
	achCount := len(a.Achievements)
	uniqueCount := len(a.UniqueExpression)
	size := 12 + (achCount * 9) + (uniqueCount * 5)

	buf := make([]byte, size)
	offset := 0

	// Achievement count (4 bytes)
	writeInt32(buf[offset:], int32(achCount))
	offset += 4

	// Achievements (9 bytes each: type(1) + unlockedAt(8))
	for _, ach := range a.Achievements {
		buf[offset] = byte(ach.Type)
		offset++
		writeUint64(buf[offset:], uint64(ach.UnlockedAt))
		offset += 8
	}

	// ExpressionCount (4 bytes)
	writeInt32(buf[offset:], int32(a.ExpressionCount))
	offset += 4

	// UniqueExpression count (4 bytes)
	writeInt32(buf[offset:], int32(uniqueCount))
	offset += 4

	// UniqueExpression map (5 bytes each: type(1) + count(4))
	for exprType, count := range a.UniqueExpression {
		buf[offset] = byte(exprType)
		offset++
		writeInt32(buf[offset:], int32(count))
		offset += 4
	}

	return buf
}

// Deserialize restores AchievementComponent from bytes.
func (a *AchievementComponent) Deserialize(data []byte) error {
	if len(data) < 12 {
		return ErrInvalidComponentData
	}

	offset := 0

	// Achievement count
	achCount := int(readInt32(data[offset:]))
	offset += 4

	// Achievements
	a.Achievements = make([]Achievement, achCount)
	for i := 0; i < achCount; i++ {
		a.Achievements[i].Type = AchievementType(data[offset])
		offset++
		// Skip timestamp reconstruction for now (use zero time)
		offset += 8
	}

	// ExpressionCount
	a.ExpressionCount = int(readInt32(data[offset:]))
	offset += 4

	// UniqueExpression count
	uniqueCount := int(readInt32(data[offset:]))
	offset += 4

	// UniqueExpression map
	a.UniqueExpression = make(map[ExpressionType]int, uniqueCount)
	for i := 0; i < uniqueCount; i++ {
		exprType := ExpressionType(data[offset])
		offset++
		count := int(readInt32(data[offset:]))
		offset += 4
		a.UniqueExpression[exprType] = count
	}

	return nil
}

// HasAchievement checks if an achievement is unlocked
func (a *AchievementComponent) HasAchievement(achievementType AchievementType) bool {
	for _, achievement := range a.Achievements {
		if achievement.Type == achievementType {
			return true
		}
	}
	return false
}

// AchievementSystem manages achievement tracking and unlocking
type AchievementSystem struct {
	world *World
}

// NewAchievementSystem creates a new achievement system
func NewAchievementSystem(world *World) *AchievementSystem {
	return &AchievementSystem{
		world: world,
	}
}

// Update processes achievement checking
func (s *AchievementSystem) Update(deltaTime float64) {
	// Achievement unlocking is event-driven, not time-based
	// This system primarily responds to events from other systems
}

// OnExpressionUsed is called when an entity uses an expression
func (s *AchievementSystem) OnExpressionUsed(entityID uint64, expressionType ExpressionType) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}

	// Get or create achievement component
	achCompRaw, ok := entity.GetComponent("achievement")
	var achComp *AchievementComponent

	if !ok {
		achComp = &AchievementComponent{
			Achievements:     []Achievement{},
			ExpressionCount:  0,
			UniqueExpression: make(map[ExpressionType]int),
		}
		entity.AddComponent(achComp)
	} else {
		achComp = achCompRaw.(*AchievementComponent)
	}

	// Update stats
	achComp.ExpressionCount++
	achComp.UniqueExpression[expressionType]++

	// Check achievements
	s.checkFirstExpression(achComp)
	s.checkExpressionMaster(achComp)
	s.checkSocialButterfly(achComp)
}

// OnComboCompleted is called when an entity completes a combo
func (s *AchievementSystem) OnComboCompleted(entityID uint64, combo *ExpressionCombo) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}

	// Get or create achievement component
	achCompRaw, ok := entity.GetComponent("achievement")
	var achComp *AchievementComponent

	if !ok {
		achComp = &AchievementComponent{
			Achievements:     []Achievement{},
			ExpressionCount:  0,
			UniqueExpression: make(map[ExpressionType]int),
		}
		entity.AddComponent(achComp)
	} else {
		achComp = achCompRaw.(*AchievementComponent)
	}

	// Get combo count from combo component
	comboCompRaw, ok := entity.GetComponent("expressioncombo")
	if !ok {
		return
	}
	comboComp := comboCompRaw.(*ExpressionComboComponent)

	// Check combo achievements
	s.checkComboStarter(achComp, comboComp.TotalCombos)
	s.checkComboExpert(achComp, comboComp.TotalCombos)
	s.checkComboLegend(achComp, comboComp.TotalCombos)
	s.checkGroupPerformer(achComp, combo)
}

// checkFirstExpression unlocks "First Expression" achievement
func (s *AchievementSystem) checkFirstExpression(achComp *AchievementComponent) {
	if achComp.HasAchievement(AchievementFirstExpression) {
		return
	}

	if achComp.ExpressionCount >= 1 {
		s.unlockAchievement(achComp, AchievementFirstExpression, "Performed your first expression!")
	}
}

// checkExpressionMaster unlocks "Expression Master" achievement
func (s *AchievementSystem) checkExpressionMaster(achComp *AchievementComponent) {
	if achComp.HasAchievement(AchievementExpressionMaster) {
		return
	}

	// Check if all 12 expressions have been used
	if len(achComp.UniqueExpression) >= 12 {
		s.unlockAchievement(achComp, AchievementExpressionMaster, "Used all 12 expressions!")
	}
}

// checkSocialButterfly unlocks "Social Butterfly" achievement
func (s *AchievementSystem) checkSocialButterfly(achComp *AchievementComponent) {
	if achComp.HasAchievement(AchievementSocialButterfly) {
		return
	}

	if achComp.ExpressionCount >= 50 {
		s.unlockAchievement(achComp, AchievementSocialButterfly, "Performed 50 expressions!")
	}
}

// checkComboStarter unlocks "Combo Starter" achievement
func (s *AchievementSystem) checkComboStarter(achComp *AchievementComponent, totalCombos int) {
	if achComp.HasAchievement(AchievementComboStarter) {
		return
	}

	if totalCombos >= 1 {
		s.unlockAchievement(achComp, AchievementComboStarter, "Completed your first combo!")
	}
}

// checkComboExpert unlocks "Combo Expert" achievement
func (s *AchievementSystem) checkComboExpert(achComp *AchievementComponent, totalCombos int) {
	if achComp.HasAchievement(AchievementComboExpert) {
		return
	}

	if totalCombos >= 10 {
		s.unlockAchievement(achComp, AchievementComboExpert, "Completed 10 combos!")
	}
}

// checkComboLegend unlocks "Combo Legend" achievement
func (s *AchievementSystem) checkComboLegend(achComp *AchievementComponent, totalCombos int) {
	if achComp.HasAchievement(AchievementComboLegend) {
		return
	}

	if totalCombos >= 100 {
		s.unlockAchievement(achComp, AchievementComboLegend, "Completed 100 combos!")
	}
}

// checkGroupPerformer unlocks "Group Performer" achievement
func (s *AchievementSystem) checkGroupPerformer(achComp *AchievementComponent, combo *ExpressionCombo) {
	if achComp.HasAchievement(AchievementGroupPerformer) {
		return
	}

	if len(combo.ParticipantIDs) >= 5 {
		s.unlockAchievement(achComp, AchievementGroupPerformer, "Performed a combo with 5+ people!")
	}
}

// unlockAchievement adds an achievement to the component
func (s *AchievementSystem) unlockAchievement(achComp *AchievementComponent, achievementType AchievementType, description string) {
	achievement := Achievement{
		Type:        achievementType,
		UnlockedAt:  s.world.Clock.Now(),
		Description: description,
	}
	achComp.Achievements = append(achComp.Achievements, achievement)
}

// GetAchievements returns all unlocked achievements for an entity
func (s *AchievementSystem) GetAchievements(entityID uint64) []Achievement {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	achCompRaw, ok := entity.GetComponent("achievement")
	if !ok {
		return nil
	}

	achComp := achCompRaw.(*AchievementComponent)
	return achComp.Achievements
}

// GetAchievementCount returns the number of unlocked achievements
func (s *AchievementSystem) GetAchievementCount(entityID uint64) int {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return 0
	}

	achCompRaw, ok := entity.GetComponent("achievement")
	if !ok {
		return 0
	}

	achComp := achCompRaw.(*AchievementComponent)
	return len(achComp.Achievements)
}
