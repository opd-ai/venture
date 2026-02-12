package engine

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MoralChoiceSystem processes moral choices, applies their consequences,
// and manages redemption arcs for regaining lost reputation.
type MoralChoiceSystem struct {
	world  *World
	logger *logrus.Logger
}

// NewMoralChoiceSystem creates a new moral choice system.
func NewMoralChoiceSystem(world *World, logger *logrus.Logger) *MoralChoiceSystem {
	if logger == nil {
		logger = logrus.New()
	}
	return &MoralChoiceSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes pending choices, checks for expirations, and updates redemption arcs.
func (s *MoralChoiceSystem) Update(deltaTime float64) {
	// Process all entities with moral choice components
	entities := s.world.GetEntitiesWith("moral_choice")
	s.logger.Infof("MoralChoiceSystem.Update: processing %d entities", len(entities))

	for _, entity := range entities {
		comp, ok := entity.GetComponent("moral_choice")
		if !ok {
			s.logger.Warn("Entity has moral_choice in query but GetComponent returned !ok", "entity", entity.ID)
			continue
		}

		moralChoice, ok := comp.(*MoralChoiceComponent)
		if !ok {
			s.logger.Warn("Component is not MoralChoiceComponent", "entity", entity.ID)
			continue
		}

		s.logger.Infof("Processing entity %d with %d pending choices", entity.ID, len(moralChoice.PendingChoices))

		// Remove expired choices
		s.removeExpiredChoices(entity, moralChoice)

		// Update redemption arcs
		s.updateRedemptionArcs(entity, moralChoice, deltaTime)
	}
}

// removeExpiredChoices removes any pending choices that have passed their expiration time.
func (s *MoralChoiceSystem) removeExpiredChoices(entity *Entity, moralChoice *MoralChoiceComponent) {
	// Iterate in reverse to safely remove while iterating
	for i := len(moralChoice.PendingChoices) - 1; i >= 0; i-- {
		choice := moralChoice.PendingChoices[i]
		if choice.IsExpired() {
			s.logger.WithFields(map[string]interface{}{
				"entity":      entity.ID,
				"choice":      choice.ID,
				"description": choice.Description,
			}).Info("Moral choice expired")

			// Remove the expired choice
			moralChoice.PendingChoices = append(
				moralChoice.PendingChoices[:i],
				moralChoice.PendingChoices[i+1:]...,
			)
		}
	}
}

// updateRedemptionArcs checks redemption arc progress and removes completed/expired arcs.
func (s *MoralChoiceSystem) updateRedemptionArcs(entity *Entity, moralChoice *MoralChoiceComponent, deltaTime float64) {
	repComp, ok := entity.GetComponent("reputation")
	if repComp == nil {
		return
	}

	reputation, ok := repComp.(*ReputationComponent)
	if !ok {
		return
	}

	// Iterate in reverse to safely remove while iterating
	for i := len(moralChoice.ActiveRedemptions) - 1; i >= 0; i-- {
		arc := &moralChoice.ActiveRedemptions[i]

		// Update current reputation from reputation component
		arc.CurrentReputation = reputation.GetReputation(arc.FactionName)

		// Check if complete
		if arc.IsComplete() {
			s.logger.Info("Redemption arc completed",
				"entity", entity.ID,
				"faction", arc.FactionName,
				"finalReputation", arc.CurrentReputation)

			// Remove completed arc
			moralChoice.ActiveRedemptions = append(
				moralChoice.ActiveRedemptions[:i],
				moralChoice.ActiveRedemptions[i+1:]...,
			)
			continue
		}

		// Check if expired
		if arc.IsExpired() {
			s.logger.Warn("Redemption arc expired",
				"entity", entity.ID,
				"faction", arc.FactionName,
				"progress", arc.GetProgress())

			// Remove expired arc
			moralChoice.ActiveRedemptions = append(
				moralChoice.ActiveRedemptions[:i],
				moralChoice.ActiveRedemptions[i+1:]...,
			)
		}
	}
}

// MakeChoice processes a moral choice selection and applies its consequences.
// Returns an error if the choice doesn't exist or the option index is invalid.
func (s *MoralChoiceSystem) MakeChoice(entity *Entity, choiceID string, optionIndex int) error {
	comp, ok := entity.GetComponent("moral_choice")
	if comp == nil {
		return fmt.Errorf("entity %d has no moral choice component", entity.ID)
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		return fmt.Errorf("invalid moral choice component type")
	}

	// Find the pending choice
	choice := moralChoice.GetPendingChoice(choiceID)
	if choice == nil {
		return fmt.Errorf("choice %s not found", choiceID)
	}

	// Validate option index
	if optionIndex < 0 || optionIndex >= len(choice.Options) {
		return fmt.Errorf("invalid option index %d (choice has %d options)", optionIndex, len(choice.Options))
	}

	option := choice.Options[optionIndex]

	// Apply rewards
	if option.Rewards != nil {
		if err := s.applyRewards(entity, option.Rewards); err != nil {
			s.logger.Warn("Failed to apply some rewards", "error", err)
		}
	}

	// Apply consequences
	if option.Consequences != nil {
		if err := s.applyConsequences(entity, option.Consequences); err != nil {
			s.logger.Warn("Failed to apply some consequences", "error", err)
		}
	}

	// Record as a deed in reputation component (this applies alignment and reputation changes)
	s.recordChoiceAsDeed(entity, choice, option)

	// Record the choice in history
	completed := CompletedChoice{
		ChoiceID:          choiceID,
		Description:       choice.Description,
		SelectedOption:    optionIndex,
		OptionLabel:       option.Label,
		Timestamp:         time.Now(),
		AlignmentChange:   option.AlignmentImpact,
		ReputationChanges: option.ReputationImpact,
		QuestID:           choice.QuestID,
	}
	moralChoice.RecordChoice(completed) // Remove from pending choices
	moralChoice.RemovePendingChoice(choiceID)

	s.logger.Info("Moral choice made",
		"entity", entity.ID,
		"choice", choiceID,
		"option", option.Label,
		"alignmentImpact", option.AlignmentImpact,
		"reputationImpact", option.ReputationImpact)

	return nil
}

// applyRewards grants rewards to the entity.
func (s *MoralChoiceSystem) applyRewards(entity *Entity, rewards *ChoiceRewards) error {
	s.applyXPReward(entity, rewards.XP)
	s.applyGoldReward(entity, rewards.Gold)
	s.applyItemRewards(entity, rewards.Items)
	s.applyQuestUnlock(entity, rewards.UnlockQuest)
	return nil
}

// applyXPReward grants experience points to entity
func (s *MoralChoiceSystem) applyXPReward(entity *Entity, xp int) {
	if xp <= 0 {
		return
	}
	expComp, ok := entity.GetComponent("experience")
	if !ok || expComp == nil {
		return
	}
	if exp, ok := expComp.(*ExperienceComponent); ok {
		exp.AddXP(xp)
	}
}

// applyGoldReward adds gold to entity's inventory
func (s *MoralChoiceSystem) applyGoldReward(entity *Entity, gold int) {
	if gold <= 0 {
		return
	}
	invComp, ok := entity.GetComponent("inventory")
	if !ok || invComp == nil {
		return
	}
	if inv, ok := invComp.(*InventoryComponent); ok {
		inv.Gold += gold
	}
}

// applyItemRewards spawns reward items near entity
func (s *MoralChoiceSystem) applyItemRewards(entity *Entity, items []string) {
	if len(items) == 0 {
		return
	}
	posComp, ok := entity.GetComponent("position")
	if !ok || posComp == nil {
		return
	}
	if pos, ok := posComp.(*PositionComponent); ok {
		s.spawnRewardItems(entity, pos, items)
	}
}

// spawnRewardItems creates item entities at position
func (s *MoralChoiceSystem) spawnRewardItems(entity *Entity, pos *PositionComponent, items []string) {
	for i, itemID := range items {
		itemEntity := s.world.CreateEntity()
		itemEntity.AddComponent(&PositionComponent{
			X: pos.X + float64(i),
			Y: pos.Y,
		})
		s.logger.Info("Item reward granted",
			"entity", entity.ID,
			"item", itemID)
	}
}

// applyQuestUnlock logs quest unlock (handled by quest system integration)
func (s *MoralChoiceSystem) applyQuestUnlock(entity *Entity, questID string) {
	if questID != "" {
		s.logger.Info("Quest unlocked",
			"entity", entity.ID,
			"quest", questID)
	}
}

// applyConsequences applies negative outcomes to the entity.
func (s *MoralChoiceSystem) applyConsequences(entity *Entity, consequences *ChoiceConsequences) error {
	s.makeFactionsHostile(entity, consequences.HostileFactions)
	s.logLostQuests(entity, consequences.LoseQuests)
	s.removeItemsFromInventory(entity, consequences.LoseItems)
	s.logEnemySpawn(entity, consequences.SpawnEnemies)
	return nil
}

// makeFactionsHostile sets specified factions to hostile reputation level.
func (s *MoralChoiceSystem) makeFactionsHostile(entity *Entity, hostileFactions []string) {
	if len(hostileFactions) == 0 {
		return
	}

	repComp, ok := entity.GetComponent("reputation")
	if !ok || repComp == nil {
		return
	}

	rep, ok := repComp.(*ReputationComponent)
	if !ok {
		return
	}

	for _, faction := range hostileFactions {
		rep.SetReputation(faction, -50.0)
		s.logger.Warn("Faction now hostile",
			"entity", entity.ID,
			"faction", faction)
	}
}

// logLostQuests logs quests that are lost as a consequence.
func (s *MoralChoiceSystem) logLostQuests(entity *Entity, loseQuests []string) {
	for _, questID := range loseQuests {
		s.logger.Warn("Quest lost",
			"entity", entity.ID,
			"quest", questID)
	}
}

// removeItemsFromInventory removes specified items from entity's inventory.
func (s *MoralChoiceSystem) removeItemsFromInventory(entity *Entity, loseItems []string) {
	if len(loseItems) == 0 {
		return
	}

	invComp, ok := entity.GetComponent("inventory")
	if !ok || invComp == nil {
		return
	}

	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return
	}

	for _, itemID := range loseItems {
		s.removeItemByID(entity, inv, itemID)
	}
}

// removeItemByID removes a single item from inventory by ID.
func (s *MoralChoiceSystem) removeItemByID(entity *Entity, inv *InventoryComponent, itemID string) {
	for i, item := range inv.Items {
		if item.ID == itemID {
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			s.logger.Warn("Item lost",
				"entity", entity.ID,
				"item", itemID)
			break
		}
	}
}

// logEnemySpawn logs enemy spawn events as a consequence.
func (s *MoralChoiceSystem) logEnemySpawn(entity *Entity, spawnCount int) {
	if spawnCount == 0 {
		return
	}

	posComp, ok := entity.GetComponent("position")
	if !ok || posComp == nil {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	s.logger.Warn("Enemies spawned",
		"entity", entity.ID,
		"count", spawnCount,
		"position", fmt.Sprintf("%.1f,%.1f", pos.X, pos.Y))
}

// recordChoiceAsDeed records the moral choice as a deed in the reputation component.
func (s *MoralChoiceSystem) recordChoiceAsDeed(entity *Entity, choice *MoralChoice, option ChoiceOption) {
	repComp, ok := entity.GetComponent("reputation")
	if repComp == nil {
		return
	}

	reputation, ok := repComp.(*ReputationComponent)
	if !ok {
		return
	}

	posComp, ok := entity.GetComponent("position")
	location := "Unknown"
	if posComp != nil {
		if pos, ok := posComp.(*PositionComponent); ok {
			location = fmt.Sprintf("%.0f,%.0f", pos.X, pos.Y)
		}
	}

	deed := Deed{
		Description:   fmt.Sprintf("%s: %s", choice.Description, option.Label),
		Timestamp:     time.Now(),
		FactionImpact: option.ReputationImpact,
		LawImpact:     option.AlignmentImpact.LawDelta,
		GoodImpact:    option.AlignmentImpact.GoodDelta,
		Location:      location,
	}

	reputation.RecordDeed(deed)
}

// StartRedemption initiates a redemption arc for regaining lost reputation with a faction.
// Returns an error if the entity already has an active redemption with that faction.
func (s *MoralChoiceSystem) StartRedemption(entity *Entity, factionName string, targetReputation float64, actions []RedemptionAction) error {
	comp, ok := entity.GetComponent("moral_choice")
	if comp == nil {
		// Create component if it doesn't exist
		comp = NewMoralChoiceComponent()
		entity.AddComponent(comp)
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		return fmt.Errorf("invalid moral choice component type")
	}

	// Check if redemption already active for this faction
	if moralChoice.GetRedemptionArc(factionName) != nil {
		return fmt.Errorf("redemption already active for faction %s", factionName)
	}

	// Get current reputation
	repComp, ok := entity.GetComponent("reputation")
	if repComp == nil {
		return fmt.Errorf("entity has no reputation component")
	}

	reputation, ok := repComp.(*ReputationComponent)
	if !ok {
		return fmt.Errorf("invalid reputation component type")
	}

	currentRep := reputation.GetReputation(factionName)

	arc := RedemptionArc{
		FactionName:        factionName,
		StartingReputation: currentRep,
		TargetReputation:   targetReputation,
		CurrentReputation:  currentRep,
		RequiredActions:    actions,
		CompletedActions:   0,
		StartTime:          time.Now(),
	}

	moralChoice.StartRedemption(arc)

	s.logger.Info("Redemption arc started",
		"entity", entity.ID,
		"faction", factionName,
		"currentRep", currentRep,
		"targetRep", targetReputation,
		"actions", len(actions))

	return nil
}

// UpdateRedemptionProgress updates progress on a specific redemption action.
// Returns an error if the redemption arc or action doesn't exist.
func (s *MoralChoiceSystem) UpdateRedemptionProgress(entity *Entity, factionName string, actionIndex, progressDelta int) error {
	moralChoice, err := s.getMoralChoiceComponent(entity)
	if err != nil {
		return err
	}

	arc, err := s.validateRedemptionArc(moralChoice, factionName, actionIndex)
	if err != nil {
		return err
	}

	action := &arc.RequiredActions[actionIndex]
	wasComplete := action.Progress >= action.Quantity
	action.Progress += progressDelta

	if !wasComplete && action.IsComplete() {
		s.handleActionCompletion(entity, arc, action, factionName)
	}

	return nil
}

// getMoralChoiceComponent retrieves and validates the moral choice component.
func (s *MoralChoiceSystem) getMoralChoiceComponent(entity *Entity) (*MoralChoiceComponent, error) {
	comp, ok := entity.GetComponent("moral_choice")
	if comp == nil {
		return nil, fmt.Errorf("entity has no moral choice component")
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		return nil, fmt.Errorf("invalid moral choice component type")
	}

	return moralChoice, nil
}

// validateRedemptionArc checks if the redemption arc and action index are valid.
func (s *MoralChoiceSystem) validateRedemptionArc(moralChoice *MoralChoiceComponent, factionName string, actionIndex int) (*RedemptionArc, error) {
	arc := moralChoice.GetRedemptionArc(factionName)
	if arc == nil {
		return nil, fmt.Errorf("no redemption arc found for faction %s", factionName)
	}

	if actionIndex < 0 || actionIndex >= len(arc.RequiredActions) {
		return nil, fmt.Errorf("invalid action index %d (arc has %d actions)", actionIndex, len(arc.RequiredActions))
	}

	return arc, nil
}

// handleActionCompletion processes completion of a redemption action.
func (s *MoralChoiceSystem) handleActionCompletion(entity *Entity, arc *RedemptionArc, action *RedemptionAction, factionName string) {
	arc.CompletedActions++
	s.applyReputationGain(entity, action, factionName)
}

// applyReputationGain applies reputation gain for completed redemption action.
func (s *MoralChoiceSystem) applyReputationGain(entity *Entity, action *RedemptionAction, factionName string) {
	repComp, ok := entity.GetComponent("reputation")
	if !ok || repComp == nil {
		return
	}

	rep, ok := repComp.(*ReputationComponent)
	if !ok {
		return
	}

	rep.AdjustReputation(factionName, action.ReputationGain)
	s.logger.Info("Redemption action completed",
		"entity", entity.ID,
		"faction", factionName,
		"action", action.Description,
		"reputationGain", action.ReputationGain)
}

// OfferFactionConflictChoice presents a choice where the player must pick sides in a faction conflict.
func (s *MoralChoiceSystem) OfferFactionConflictChoice(entity *Entity, faction1, faction2, context string) error {
	comp, ok := entity.GetComponent("moral_choice")
	if comp == nil {
		// Create component if it doesn't exist
		comp = NewMoralChoiceComponent()
		entity.AddComponent(comp)
	}

	moralChoice, ok := comp.(*MoralChoiceComponent)
	if !ok {
		return fmt.Errorf("invalid moral choice component type")
	}

	choiceID := fmt.Sprintf("faction_conflict_%s_vs_%s_%d", faction1, faction2, time.Now().Unix())

	choice := MoralChoice{
		ID:          choiceID,
		Description: fmt.Sprintf("Conflict between %s and %s", faction1, faction2),
		Context:     context,
		Options: []ChoiceOption{
			{
				Label:       fmt.Sprintf("Support %s", faction1),
				Description: fmt.Sprintf("Aid the %s in their conflict", faction1),
				ReputationImpact: map[string]float64{
					faction1: 20.0,
					faction2: -30.0,
				},
				AlignmentImpact: AlignmentDelta{
					LawDelta:  -0.05, // Taking sides is somewhat chaotic
					GoodDelta: 0.0,
				},
			},
			{
				Label:       fmt.Sprintf("Support %s", faction2),
				Description: fmt.Sprintf("Aid the %s in their conflict", faction2),
				ReputationImpact: map[string]float64{
					faction1: -30.0,
					faction2: 20.0,
				},
				AlignmentImpact: AlignmentDelta{
					LawDelta:  -0.05, // Taking sides is somewhat chaotic
					GoodDelta: 0.0,
				},
			},
			{
				Label:       "Stay neutral",
				Description: "Refuse to take sides in the conflict",
				ReputationImpact: map[string]float64{
					faction1: -5.0,
					faction2: -5.0,
				},
				AlignmentImpact: AlignmentDelta{
					LawDelta:  0.05, // Neutrality is lawful
					GoodDelta: 0.0,
				},
			},
		},
	}

	moralChoice.AddChoice(choice)

	s.logger.Info("Faction conflict choice offered",
		"entity", entity.ID,
		"faction1", faction1,
		"faction2", faction2)

	return nil
}
