// Package engine provides the CarryOver system for New Game Plus.
// This file implements CarryOverSystem which manages the selection,
// validation, and transfer of items/progress between NG+ cycles.
//
// Phase 112: Carry-Over System
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// CarryOverSystem manages carry-over selection UI and transfer of items between NG+ cycles.
// It validates selections against limits and coordinates the transfer process.
type CarryOverSystem struct {
	world  *World
	logger *logrus.Entry

	// Callbacks for carry-over events
	onSelectionComplete func(entityID uint64)
	onTransferComplete  func(entityID uint64, summary CarryOverSummary)
	onEquipmentTransfer func(entityID uint64, items []*item.Item)
	onCurrencyTransfer  func(entityID uint64, currency map[string]int64)
	onSkillsTransfer    func(entityID uint64, skills []string)

	// scaleItemLevel determines if equipment should be level-scaled on transfer
	scaleItemLevel bool
}

// NewCarryOverSystem creates a new carry-over system.
func NewCarryOverSystem(world *World) *CarryOverSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "carryover")
		logEntry.Debug("Carry-over system created")
	}
	return &CarryOverSystem{
		world:          world,
		logger:         logEntry,
		scaleItemLevel: true, // Default: scale items to player level
	}
}

// Update processes entities with carry-over components.
// Primarily validates selections and checks for transfer readiness.
func (s *CarryOverSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		carryComp, ok := entity.GetComponent("carryover")
		if !ok {
			continue
		}

		carry, ok := carryComp.(*CarryOverComponent)
		if !ok {
			continue
		}

		// If confirmed and locked but not yet transferred, process transfer
		if carry.IsConfirmed() && carry.IsLocked() && !carry.IsTransferComplete() {
			s.processTransfer(entity, carry)
		}
	}
}

// PrepareForNGPlus initializes carry-over selection for a player entity.
// Called when the player completes the game and is ready to start NG+.
func (s *CarryOverSystem) PrepareForNGPlus(entity *Entity) error {
	// Get or create carry-over component
	carryComp, ok := entity.GetComponent("carryover")
	var carry *CarryOverComponent
	if !ok {
		carry = NewCarryOverComponent()
		entity.AddComponent(carry)
	} else {
		carry, ok = carryComp.(*CarryOverComponent)
		if !ok {
			carry = NewCarryOverComponent()
			entity.AddComponent(carry)
		}
	}

	// Reset for new selection
	carry.Reset()

	// Set limits from NG+ component
	if ngpComp, ok := entity.GetComponent("newgameplus"); ok {
		if ngp, ok := ngpComp.(*NewGamePlusComponent); ok {
			totalSkills := s.countPlayerSkills(entity)
			carry.SetLimitsFromNGPlus(ngp, totalSkills)
		}
	}

	// Populate cosmetics and achievements (these always carry over)
	s.populatePermanentUnlocks(entity, carry)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"equipment_limit": carry.GetEquipmentSlotLimit(),
			"skill_limit":     carry.GetSkillSlotLimit(),
			"currency_limit":  carry.GetCurrencyPercentLimit(),
		}).Info("Prepared entity for NG+ carry-over selection")
	}

	return nil
}

// countPlayerSkills returns the total number of skills the player has.
func (s *CarryOverSystem) countPlayerSkills(entity *Entity) int {
	// Check for skill book component
	if skillComp, ok := entity.GetComponent("skill_book"); ok {
		if skillBook, ok := skillComp.(*SkillBookComponent); ok {
			return len(skillBook.LearnedSkills)
		}
	}

	// Check for spell component (spells count as skills)
	if spellComp, ok := entity.GetComponent("spell"); ok {
		if spellList, ok := spellComp.(*SpellComponent); ok {
			return len(spellList.KnownSpells)
		}
	}

	return 0
}

// populatePermanentUnlocks adds cosmetics and achievements to carry-over.
func (s *CarryOverSystem) populatePermanentUnlocks(entity *Entity, carry *CarryOverComponent) {
	s.transferCosmetics(entity, carry)
	s.transferExtendedAchievements(entity, carry)
	s.transferBasicAchievements(entity, carry)
}

// transferCosmetics extracts and adds unlocked cosmetics to the carry-over component.
func (s *CarryOverSystem) transferCosmetics(entity *Entity, carry *CarryOverComponent) {
	cosmeticComp, ok := entity.GetComponent("cosmetic")
	if !ok {
		return
	}

	customization, ok := cosmeticComp.(*CosmeticComponent)
	if !ok {
		return
	}

	for _, cosmeticID := range customization.UnlockedCosmetics {
		carry.AddCosmetic(cosmeticID)
	}
}

// transferExtendedAchievements extracts and adds tiered achievements to the carry-over component.
func (s *CarryOverSystem) transferExtendedAchievements(entity *Entity, carry *CarryOverComponent) {
	achieveComp, ok := entity.GetComponent("extended_achievement")
	if !ok {
		return
	}

	extAchieve, ok := achieveComp.(*ExtendedAchievementComponent)
	if !ok {
		return
	}

	for achieveID, entry := range extAchieve.Achievements {
		if entry.CurrentTier > AchievementTierNone {
			carry.AddAchievement(achieveID)
		}
	}
}

// transferBasicAchievements extracts and adds basic achievements to the carry-over component.
func (s *CarryOverSystem) transferBasicAchievements(entity *Entity, carry *CarryOverComponent) {
	achieveComp, ok := entity.GetComponent("achievement")
	if !ok {
		return
	}

	achieve, ok := achieveComp.(*AchievementComponent)
	if !ok {
		return
	}

	for _, achieveID := range achieve.GetUnlockedIDs() {
		carry.AddAchievement(achieveID)
	}
}

// SelectEquipmentFromInventory allows player to select an item by index from inventory.
func (s *CarryOverSystem) SelectEquipmentFromInventory(entity *Entity, itemIndex int) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil || carry.IsLocked() {
		return false
	}

	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return false
	}

	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return false
	}

	if itemIndex < 0 || itemIndex >= len(inv.Items) {
		return false
	}

	itm := inv.Items[itemIndex]
	if itm == nil {
		return false
	}

	// Generate a unique ID for this item
	itemID := s.generateItemID(itm)
	return carry.SelectEquipment(itemID)
}

// generateItemID creates a unique identifier for an item.
func (s *CarryOverSystem) generateItemID(itm *item.Item) string {
	if itm == nil {
		return ""
	}
	// Use name + seed for uniqueness
	return itm.Name + "_" + formatSeed(itm.Seed)
}

// formatSeed converts a seed to a string representation.
func formatSeed(seed int64) string {
	if seed == 0 {
		return "0"
	}
	result := ""
	n := seed
	if n < 0 {
		n = -n
		result = "-"
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return result + digits
}

// DeselectEquipment removes an item from carry-over selection.
func (s *CarryOverSystem) DeselectEquipment(entity *Entity, itemID string) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return false
	}
	return carry.DeselectEquipment(itemID)
}

// SetCurrencyAmount sets the currency amount to carry over.
// The amount will be capped by the percentage limit during transfer.
func (s *CarryOverSystem) SetCurrencyAmount(entity *Entity, currencyType string, amount int64) {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return
	}
	carry.SetCurrencyCarryOver(currencyType, amount)
}

// SetGoldCarryOver is a convenience method for setting gold carry-over.
func (s *CarryOverSystem) SetGoldCarryOver(entity *Entity, amount int64) {
	s.SetCurrencyAmount(entity, "gold", amount)
}

// SelectSkill adds a skill to carry-over selection.
func (s *CarryOverSystem) SelectSkill(entity *Entity, skillID string) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil || carry.IsLocked() {
		return false
	}
	return carry.SelectSkill(skillID)
}

// DeselectSkill removes a skill from carry-over selection.
func (s *CarryOverSystem) DeselectSkill(entity *Entity, skillID string) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return false
	}
	return carry.DeselectSkill(skillID)
}

// ConfirmSelection locks in the player's carry-over choices.
func (s *CarryOverSystem) ConfirmSelection(entity *Entity) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return false
	}

	if !carry.ConfirmSelection() {
		return false
	}

	if s.logger != nil {
		summary := carry.GetSummary()
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"equipment":  summary.EquipmentCount,
			"skills":     summary.SkillCount,
			"currencies": summary.CurrencyTypes,
		}).Info("Carry-over selection confirmed")
	}

	if s.onSelectionComplete != nil {
		s.onSelectionComplete(entity.ID)
	}

	return true
}

// LockAndTransfer locks selections and initiates transfer.
// Call this when the player is ready to start NG+.
func (s *CarryOverSystem) LockAndTransfer(entity *Entity) bool {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return false
	}

	if !carry.IsConfirmed() {
		return false
	}

	carry.Lock()
	return true
}

// processTransfer executes the actual carry-over transfer.
func (s *CarryOverSystem) processTransfer(entity *Entity, carry *CarryOverComponent) {
	if carry.IsTransferComplete() {
		return
	}

	// Transfer equipment
	transferredItems := s.transferEquipment(entity, carry)

	// Transfer currency
	transferredCurrency := s.transferCurrency(entity, carry)

	// Transfer skills
	transferredSkills := s.transferSkills(entity, carry)

	// Mark complete
	carry.MarkTransferComplete()

	summary := carry.GetSummary()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"items":        len(transferredItems),
			"currencies":   len(transferredCurrency),
			"skills":       len(transferredSkills),
			"cosmetics":    summary.CosmeticCount,
			"achievements": summary.AchievementCount,
		}).Info("Carry-over transfer complete")
	}

	// Fire callbacks
	if s.onEquipmentTransfer != nil && len(transferredItems) > 0 {
		s.onEquipmentTransfer(entity.ID, transferredItems)
	}
	if s.onCurrencyTransfer != nil && len(transferredCurrency) > 0 {
		s.onCurrencyTransfer(entity.ID, transferredCurrency)
	}
	if s.onSkillsTransfer != nil && len(transferredSkills) > 0 {
		s.onSkillsTransfer(entity.ID, transferredSkills)
	}
	if s.onTransferComplete != nil {
		s.onTransferComplete(entity.ID, summary)
	}
}

// transferEquipment handles equipment carry-over.
func (s *CarryOverSystem) transferEquipment(entity *Entity, carry *CarryOverComponent) []*item.Item {
	selectedIDs := carry.GetSelectedEquipment()
	if len(selectedIDs) == 0 {
		return nil
	}

	// Get inventory
	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil
	}

	// Find matching items
	var transferredItems []*item.Item
	for _, itemID := range selectedIDs {
		for _, itm := range inv.Items {
			if s.generateItemID(itm) == itemID {
				// Apply level scaling if enabled
				if s.scaleItemLevel {
					s.scaleItemToLevel(entity, itm)
				}
				transferredItems = append(transferredItems, itm)
				break
			}
		}
	}

	return transferredItems
}

// scaleItemToLevel adjusts item stats based on player level.
func (s *CarryOverSystem) scaleItemToLevel(entity *Entity, itm *item.Item) {
	if itm == nil {
		return
	}

	// Get player level from ExperienceComponent
	playerLevel := 1
	if expComp, ok := entity.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			playerLevel = exp.Level
		}
	}

	// Scale item required level to player level (if lower)
	itemReqLevel := itm.Stats.RequiredLevel
	if itemReqLevel < playerLevel {
		// Calculate scaling factor (diminishing returns)
		scaleFactor := 1.0 + (float64(playerLevel-itemReqLevel) * 0.02)
		if scaleFactor > 1.5 {
			scaleFactor = 1.5 // Cap at 50% increase
		}

		// Scale stats
		itm.Stats.Damage = int(float64(itm.Stats.Damage) * scaleFactor)
		itm.Stats.Defense = int(float64(itm.Stats.Defense) * scaleFactor)
		itm.Stats.RequiredLevel = playerLevel
	}
}

// transferCurrency handles currency carry-over.
func (s *CarryOverSystem) transferCurrency(entity *Entity, carry *CarryOverComponent) map[string]int64 {
	currencyMap := carry.GetAllCurrencyCarryOver()
	if len(currencyMap) == 0 {
		return nil
	}

	// Get inventory for gold
	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil
	}

	transferred := make(map[string]int64)

	// Handle gold specifically
	if goldAmount, exists := currencyMap["gold"]; exists && goldAmount > 0 {
		finalGold := carry.CalculateFinalCurrencyAmount("gold", int64(inv.Gold))
		if finalGold > goldAmount {
			finalGold = goldAmount // Don't exceed requested amount
		}
		if finalGold > 0 {
			transferred["gold"] = finalGold
			// Note: Actual gold transfer happens in the callback
		}
	}

	// Other currencies would be handled via callbacks
	for currType, amount := range currencyMap {
		if currType == "gold" {
			continue // Already handled
		}
		if amount > 0 {
			transferred[currType] = amount
		}
	}

	return transferred
}

// transferSkills handles skill carry-over.
func (s *CarryOverSystem) transferSkills(entity *Entity, carry *CarryOverComponent) []string {
	selectedSkills := carry.GetSelectedSkills()
	if len(selectedSkills) == 0 {
		return nil
	}

	// Skills are preserved by ID; actual restoration happens in game initialization
	return selectedSkills
}

// getCarryOverComponent retrieves the carry-over component from an entity.
func (s *CarryOverSystem) getCarryOverComponent(entity *Entity) *CarryOverComponent {
	if entity == nil {
		return nil
	}

	comp, ok := entity.GetComponent("carryover")
	if !ok {
		return nil
	}

	carry, ok := comp.(*CarryOverComponent)
	if !ok {
		return nil
	}

	return carry
}

// GetAvailableEquipment returns a list of items that can be selected for carry-over.
func (s *CarryOverSystem) GetAvailableEquipment(entity *Entity) []EquipmentOption {
	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil
	}

	carry := s.getCarryOverComponent(entity)

	var options []EquipmentOption
	for i, itm := range inv.Items {
		if itm == nil {
			continue
		}
		// Only show equippable items
		if !itm.IsEquippable() {
			continue
		}

		itemID := s.generateItemID(itm)
		selected := false
		if carry != nil {
			selected = carry.IsEquipmentSelected(itemID)
		}

		options = append(options, EquipmentOption{
			Index:    i,
			ItemID:   itemID,
			Name:     itm.Name,
			Type:     itm.Type.String(),
			Rarity:   itm.Rarity.String(),
			Level:    itm.Stats.RequiredLevel,
			Selected: selected,
		})
	}

	return options
}

// EquipmentOption represents an item available for carry-over selection.
type EquipmentOption struct {
	Index    int    `json:"index"`
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Rarity   string `json:"rarity"`
	Level    int    `json:"level"`
	Selected bool   `json:"selected"`
}

// GetAvailableSkills returns a list of skills that can be selected for carry-over.
func (s *CarryOverSystem) GetAvailableSkills(entity *Entity) []SkillOption {
	carry := s.getCarryOverComponent(entity)

	var options []SkillOption

	// Check skill book
	if skillComp, ok := entity.GetComponent("skill_book"); ok {
		if skillBook, ok := skillComp.(*SkillBookComponent); ok {
			for skillID, skill := range skillBook.LearnedSkills {
				selected := false
				if carry != nil {
					selected = carry.IsSkillSelected(skillID)
				}
				options = append(options, SkillOption{
					SkillID:  skillID,
					Name:     skill.Name,
					Type:     skill.Type,
					Level:    skill.Level,
					Selected: selected,
				})
			}
		}
	}

	// Check spell component
	if spellComp, ok := entity.GetComponent("spell"); ok {
		if spellList, ok := spellComp.(*SpellComponent); ok {
			for spellID, spell := range spellList.KnownSpells {
				selected := false
				if carry != nil {
					selected = carry.IsSkillSelected(spellID)
				}
				options = append(options, SkillOption{
					SkillID:  spellID,
					Name:     spell.Name,
					Type:     "spell",
					Level:    spell.SpellLevel,
					Selected: selected,
				})
			}
		}
	}

	return options
}

// SkillOption represents a skill available for carry-over selection.
type SkillOption struct {
	SkillID  string `json:"skill_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Level    int    `json:"level"`
	Selected bool   `json:"selected"`
}

// GetCarryOverStatus returns the current carry-over selection status.
func (s *CarryOverSystem) GetCarryOverStatus(entity *Entity) *CarryOverSummary {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return nil
	}
	summary := carry.GetSummary()
	return &summary
}

// SetScaleItemLevel enables or disables item level scaling on transfer.
func (s *CarryOverSystem) SetScaleItemLevel(enabled bool) {
	s.scaleItemLevel = enabled
}

// SetOnSelectionComplete sets a callback for when selection is confirmed.
func (s *CarryOverSystem) SetOnSelectionComplete(callback func(entityID uint64)) {
	s.onSelectionComplete = callback
}

// SetOnTransferComplete sets a callback for when transfer is complete.
func (s *CarryOverSystem) SetOnTransferComplete(callback func(entityID uint64, summary CarryOverSummary)) {
	s.onTransferComplete = callback
}

// SetOnEquipmentTransfer sets a callback for equipment transfer.
func (s *CarryOverSystem) SetOnEquipmentTransfer(callback func(entityID uint64, items []*item.Item)) {
	s.onEquipmentTransfer = callback
}

// SetOnCurrencyTransfer sets a callback for currency transfer.
func (s *CarryOverSystem) SetOnCurrencyTransfer(callback func(entityID uint64, currency map[string]int64)) {
	s.onCurrencyTransfer = callback
}

// SetOnSkillsTransfer sets a callback for skills transfer.
func (s *CarryOverSystem) SetOnSkillsTransfer(callback func(entityID uint64, skills []string)) {
	s.onSkillsTransfer = callback
}

// CancelCarryOver unlocks and clears selections.
func (s *CarryOverSystem) CancelCarryOver(entity *Entity) {
	carry := s.getCarryOverComponent(entity)
	if carry == nil {
		return
	}

	carry.Unlock()
	carry.ClearSelections()

	if s.logger != nil {
		s.logger.WithField("entity_id", entity.ID).Debug("Carry-over cancelled")
	}
}

// ApplyCarryOverToNewCharacter restores carry-over items to a new character.
// This should be called during character creation in NG+.
func (s *CarryOverSystem) ApplyCarryOverToNewCharacter(entity *Entity, carryOverData *CarryOverComponent) error {
	if carryOverData == nil || !carryOverData.IsTransferComplete() {
		return nil
	}

	// Apply cosmetics
	if cosmeticComp, ok := entity.GetComponent("cosmetic"); ok {
		if customization, ok := cosmeticComp.(*CosmeticComponent); ok {
			for _, cosmeticID := range carryOverData.GetCosmetics() {
				customization.UnlockCosmetic(cosmeticID)
			}
		}
	}

	// Apply achievements
	if achieveComp, ok := entity.GetComponent("achievement"); ok {
		if achieve, ok := achieveComp.(*AchievementComponent); ok {
			for _, achieveID := range carryOverData.GetAchievements() {
				achieve.Unlock(achieveID)
			}
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"cosmetics":    len(carryOverData.GetCosmetics()),
			"achievements": len(carryOverData.GetAchievements()),
		}).Info("Applied carry-over to new character")
	}

	return nil
}
