// Package engine provides menu system for game UI.
// This file implements MenuSystem which handles in-game menus including
// main menu, save/load menus, and menu navigation.
package engine

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

// MenuType represents the type of menu being displayed.
type MenuType int

const (
	MenuTypeNone MenuType = iota
	MenuTypeMain
	MenuTypeSave
	MenuTypeLoad
	MenuTypeConfirm
)

// MenuItem represents a single menu option.
type MenuItem struct {
	Label    string
	Action   func() error // Callback when item is selected
	Enabled  bool
	Metadata interface{} // Optional data (e.g., save metadata)
}

// MenuComponent stores menu state data.
type MenuComponent struct {
	Active         bool
	CurrentMenu    MenuType
	Items          []MenuItem
	SelectedIndex  int
	MenuStack      []MenuType // For nested menu navigation
	ErrorMessage   string
	ErrorTimeout   float64 // Seconds remaining to show error
	ConfirmMessage string  // Message for confirmation dialogs
	ConfirmAction  func() error
}

// Type returns the component type identifier.
func (m *MenuComponent) Type() string {
	return "menu"
}

// MenuSystem manages the game menu, including pause, save, and load functionality.
type EbitenMenuSystem struct {
	world        *World
	screenWidth  int
	screenHeight int
	saveManager  *saveload.SaveManager

	// Callbacks for save/load operations
	onSave func(name string) error
	onLoad func(name string) error

	// Menu component reference (stored on a dedicated menu entity)
	menuEntity *Entity

	// Structured logger for menu system
	logger *logrus.Entry
}

// NewEbitenMenuSystem creates a new menu system.
func NewEbitenMenuSystem(world *World, screenWidth, screenHeight int, saveDir string) (*EbitenMenuSystem, error) {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "menu")
		logEntry.WithFields(logrus.Fields{
			"screen_width":  screenWidth,
			"screen_height": screenHeight,
			"save_dir":      saveDir,
		}).Debug("Creating menu system")
	}

	saveManager, err := saveload.NewSaveManager(saveDir)
	if err != nil {
		if logEntry != nil {
			logEntry.WithFields(logrus.Fields{
				"save_dir": saveDir,
				"error":    err.Error(),
			}).Error("Failed to initialize save manager")
		}
		return nil, fmt.Errorf("failed to initialize save manager: %w", err)
	}

	if logEntry != nil {
		logEntry.Info("Menu system created successfully")
	}

	return &EbitenMenuSystem{
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		saveManager:  saveManager,
		logger:       logEntry,
	}, nil
}

// SetSaveCallback sets the callback for save operations.
func (ms *EbitenMenuSystem) SetSaveCallback(callback func(name string) error) {
	if ms.logger != nil {
		ms.logger.WithField("callback_set", callback != nil).Debug("SetSaveCallback called")
	}
	ms.onSave = callback
	if ms.logger != nil {
		ms.logger.Debug("Save callback configured")
	}
}

// SetLoadCallback sets the callback for load operations.
func (ms *EbitenMenuSystem) SetLoadCallback(callback func(name string) error) {
	if ms.logger != nil {
		ms.logger.WithField("callback_set", callback != nil).Debug("SetLoadCallback called")
	}
	ms.onLoad = callback
	if ms.logger != nil {
		ms.logger.Debug("Load callback configured")
	}
}

// Toggle opens or closes the main menu.
func (ms *EbitenMenuSystem) Toggle() {
	if ms.logger != nil {
		ms.logger.WithField("has_menu_entity", ms.menuEntity != nil).Debug("Toggle called")
	}
	if ms.menuEntity == nil {
		if ms.logger != nil {
			ms.logger.Debug("Creating menu entity")
		}
		ms.menuEntity = ms.world.CreateEntity()
		menu := &MenuComponent{
			Active:      true,
			CurrentMenu: MenuTypeMain,
		}
		ms.menuEntity.AddComponent(menu)
		ms.buildMainMenu(menu)
		ms.world.Update(0) // Process entity addition
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"entity_id": ms.menuEntity.ID,
				"menu_type": "main",
				"active":    true,
			}).Info("Menu opened")
		}
	} else {
		if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
			// Type assert with safety check
			if menuComp, ok := menu.(*MenuComponent); ok {
				previousState := menuComp.Active
				menuComp.Active = !menuComp.Active

				// Rebuild main menu when opening
				if menuComp.Active {
					menuComp.CurrentMenu = MenuTypeMain
					menuComp.MenuStack = nil
					ms.buildMainMenu(menuComp)
				}

				if ms.logger != nil {
					ms.logger.WithFields(logrus.Fields{
						"entity_id":      ms.menuEntity.ID,
						"previous_state": previousState,
						"current_state":  menuComp.Active,
					}).Debug("Menu toggled")
				}
			}
		}
	}
}

// IsActive returns true if the menu is currently displayed.
func (ms *EbitenMenuSystem) IsActive() bool {
	if ms.menuEntity == nil {
		if ms.logger != nil {
			ms.logger.Debug("IsActive called: no menu entity")
		}
		return false
	}
	if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
		// Type assert with safety check
		if menuComp, ok := menu.(*MenuComponent); ok {
			if ms.logger != nil {
				ms.logger.WithField("active", menuComp.Active).Debug("IsActive called")
			}
			return menuComp.Active
		}
	}
	if ms.logger != nil {
		ms.logger.Debug("IsActive called: component retrieval failed")
	}
	return false
}

// Update processes menu input and state.
func (ms *EbitenMenuSystem) Update(entities []*Entity, deltaTime float64) {
	menuComp := ms.validateMenuComponent()
	if menuComp == nil {
		return
	}

	if !menuComp.Active {
		if ms.logger != nil {
			ms.logger.Debug("Update skipped: menu inactive")
		}
		return
	}

	ms.logMenuUpdate(menuComp, deltaTime)
	ms.updateErrorTimeout(menuComp, deltaTime)
	ms.handleInput(menuComp)
}

// validateMenuComponent retrieves and validates the menu component.
func (ms *EbitenMenuSystem) validateMenuComponent() *MenuComponent {
	if ms.logger != nil {
		ms.logger.Debug("Validating menu component")
	}
	if ms.menuEntity == nil {
		if ms.logger != nil {
			ms.logger.Debug("Validation failed: no menu entity")
		}
		return nil
	}

	menu, ok := ms.menuEntity.GetComponent("menu")
	if !ok {
		if ms.logger != nil {
			ms.logger.WithField("entity_id", ms.menuEntity.ID).Warn("Menu entity missing menu component")
		}
		return nil
	}

	menuComp, ok := menu.(*MenuComponent)
	if !ok {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"entity_id":      ms.menuEntity.ID,
				"component_type": "menu",
			}).Warn("Failed to type assert menu component")
		}
		return nil
	}

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"entity_id":    ms.menuEntity.ID,
			"active":       menuComp.Active,
			"current_menu": menuComp.CurrentMenu,
		}).Debug("Menu component validated successfully")
	}
	return menuComp
}

// logMenuUpdate logs debug information about menu update.
func (ms *EbitenMenuSystem) logMenuUpdate(menuComp *MenuComponent, deltaTime float64) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"delta_time":     deltaTime,
			"current_menu":   menuComp.CurrentMenu,
			"selected_index": menuComp.SelectedIndex,
			"item_count":     len(menuComp.Items),
		}).Debug("Menu system update started")
	}
}

// updateErrorTimeout decrements error message timeout and clears message when expired.
func (ms *EbitenMenuSystem) updateErrorTimeout(menuComp *MenuComponent, deltaTime float64) {
	if menuComp.ErrorTimeout > 0 {
		previousTimeout := menuComp.ErrorTimeout
		menuComp.ErrorTimeout -= deltaTime
		if menuComp.ErrorTimeout <= 0 {
			if ms.logger != nil {
				ms.logger.WithFields(logrus.Fields{
					"previous_timeout": previousTimeout,
					"error_message":    menuComp.ErrorMessage,
				}).Debug("Error message timeout expired")
			}
			menuComp.ErrorMessage = ""
			if ms.logger != nil {
				ms.logger.Debug("Error message cleared")
			}
		}
	}
}

// handleInput processes keyboard and mouse input for menu navigation.
func (ms *EbitenMenuSystem) handleInput(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"current_menu":   menu.CurrentMenu,
			"selected_index": menu.SelectedIndex,
		}).Debug("Processing menu input")
	}
	menuX := (ms.screenWidth - 400) / 2
	menuY := (ms.screenHeight - 300) / 2

	ms.handleMouseInput(menu, menuX, menuY)
	ms.handleKeyboardNavigation(menu)
	ms.handleKeyboardSelection(menu)
	ms.handleBackCancel(menu)
}

// handleMouseInput processes mouse and touch input for menu items.
func (ms *EbitenMenuSystem) handleMouseInput(menu *MenuComponent, menuX, menuY int) {
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	mouseClicked := IsTouchOrMouseJustPressed()

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"mouse_x":       mouseX,
			"mouse_y":       mouseY,
			"mouse_clicked": mouseClicked,
		}).Debug("Handling mouse input")
	}

	itemY := menuY + 70
	for i := range menu.Items {
		itemX := menuX + 10
		itemWidth := 380
		itemHeight := 20

		if mouseX >= itemX && mouseX < itemX+itemWidth &&
			mouseY >= itemY && mouseY < itemY+itemHeight {
			if menu.SelectedIndex != i {
				if ms.logger != nil {
					ms.logger.WithFields(logrus.Fields{
						"previous_index": menu.SelectedIndex,
						"new_index":      i,
						"item_label":     menu.Items[i].Label,
						"input_method":   "mouse",
					}).Debug("Menu item selected")
				}
			}
			menu.SelectedIndex = i

			if mouseClicked {
				if ms.logger != nil {
					ms.logger.WithFields(logrus.Fields{
						"item_index": i,
						"item_label": menu.Items[i].Label,
						"enabled":    menu.Items[i].Enabled,
					}).Debug("Mouse click on menu item")
				}
				ms.executeMenuItem(menu, i)
			}
		}

		itemY += 25
	}
}

// handleKeyboardNavigation processes keyboard navigation (up/down arrows).
func (ms *EbitenMenuSystem) handleKeyboardNavigation(menu *MenuComponent) {
	previousIndex := menu.SelectedIndex
	var direction string

	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		direction = "up"
		menu.SelectedIndex--
		if menu.SelectedIndex < 0 {
			menu.SelectedIndex = len(menu.Items) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		direction = "down"
		menu.SelectedIndex++
		if menu.SelectedIndex >= len(menu.Items) {
			menu.SelectedIndex = 0
		}
	}

	if direction != "" && previousIndex != menu.SelectedIndex {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"direction":      direction,
				"previous_index": previousIndex,
				"new_index":      menu.SelectedIndex,
				"item_label":     menu.Items[menu.SelectedIndex].Label,
			}).Debug("Keyboard navigation")
		}
	}
}

// handleKeyboardSelection processes keyboard selection (Enter/Space).
func (ms *EbitenMenuSystem) handleKeyboardSelection(menu *MenuComponent) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if menu.SelectedIndex >= 0 && menu.SelectedIndex < len(menu.Items) {
			if ms.logger != nil {
				ms.logger.WithFields(logrus.Fields{
					"item_index": menu.SelectedIndex,
					"item_label": menu.Items[menu.SelectedIndex].Label,
					"enabled":    menu.Items[menu.SelectedIndex].Enabled,
				}).Debug("Keyboard selection")
			}
			ms.executeMenuItem(menu, menu.SelectedIndex)
		}
	}
}

// handleBackCancel processes Escape key for back/cancel navigation.
func (ms *EbitenMenuSystem) handleBackCancel(menu *MenuComponent) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if len(menu.MenuStack) > 0 {
			previousMenu := menu.CurrentMenu
			menu.CurrentMenu = menu.MenuStack[len(menu.MenuStack)-1]
			menu.MenuStack = menu.MenuStack[:len(menu.MenuStack)-1]
			if ms.logger != nil {
				ms.logger.WithFields(logrus.Fields{
					"previous_menu": previousMenu,
					"current_menu":  menu.CurrentMenu,
					"stack_depth":   len(menu.MenuStack),
				}).Debug("Navigating back in menu stack")
			}
			ms.rebuildMenu(menu)
		} else {
			if ms.logger != nil {
				ms.logger.Debug("Closing menu via Escape key")
			}
			menu.Active = false
			if ms.logger != nil {
				ms.logger.WithField("active", menu.Active).Debug("Menu state changed to inactive")
			}
		}
	}
}

// executeMenuItem executes a menu item's action and handles errors.
func (ms *EbitenMenuSystem) executeMenuItem(menu *MenuComponent, index int) {
	item := menu.Items[index]

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"item_index": index,
			"item_label": item.Label,
			"enabled":    item.Enabled,
			"has_action": item.Action != nil,
		}).Debug("Executing menu item")
	}

	if !item.Enabled {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"item_index": index,
				"item_label": item.Label,
			}).Warn("Attempted to execute disabled menu item")
		}
		return
	}

	if item.Action == nil {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"item_index": index,
				"item_label": item.Label,
			}).Warn("Menu item has no action defined")
		}
		return
	}

	if err := item.Action(); err != nil {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"item_index": index,
				"item_label": item.Label,
				"error":      err.Error(),
			}).Error("Menu item action failed")
		}
		menu.ErrorMessage = err.Error()
		menu.ErrorTimeout = 3.0
	} else {
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"item_index": index,
				"item_label": item.Label,
			}).Debug("Menu item action executed successfully")
		}
	}
}

// buildMainMenu constructs the main pause menu.
func (ms *EbitenMenuSystem) buildMainMenu(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.Debug("Building main menu")
	}
	menu.Items = []MenuItem{
		{
			Label:   "Save Game",
			Enabled: true,
			Action: func() error {
				menu.MenuStack = append(menu.MenuStack, menu.CurrentMenu)
				menu.CurrentMenu = MenuTypeSave
				ms.buildSaveMenu(menu)
				return nil
			},
		},
		{
			Label:   "Load Game",
			Enabled: true,
			Action: func() error {
				menu.MenuStack = append(menu.MenuStack, menu.CurrentMenu)
				menu.CurrentMenu = MenuTypeLoad
				ms.buildLoadMenu(menu)
				return nil
			},
		},
		{
			Label:   "Resume Game",
			Enabled: true,
			Action: func() error {
				menu.Active = false
				return nil
			},
		},
		{
			Label:   "Exit to Desktop",
			Enabled: true,
			Action: func() error {
				// Confirm before exiting
				menu.ConfirmMessage = "Exit game? Unsaved progress will be lost."
				menu.ConfirmAction = func() error {
					// Note: Actual exit requires Game integration
					// For now, just close menu
					menu.Active = false
					return fmt.Errorf("exit not implemented (close window manually)")
				}
				menu.MenuStack = append(menu.MenuStack, menu.CurrentMenu)
				menu.CurrentMenu = MenuTypeConfirm
				ms.buildConfirmMenu(menu)
				return nil
			},
		},
	}
	menu.SelectedIndex = 0
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"item_count":     len(menu.Items),
			"selected_index": menu.SelectedIndex,
		}).Debug("Main menu built")
	}
}

// buildSaveMenu constructs the save game menu with available save slots.
func (ms *EbitenMenuSystem) buildSaveMenu(menu *MenuComponent) {
	ms.logBuildingMenu()
	menu.Items = []MenuItem{
		ms.createSaveMenuItem("Quick Save (slot 1)", "quicksave", "Quick Save"),
		ms.createSaveMenuItem("Auto Save (slot 2)", "autosave", "Auto Save"),
		ms.createSaveMenuItem("Save Slot 3", "save3", "Slot 3"),
		ms.createBackMenuItem(menu),
	}
	menu.SelectedIndex = 0
	ms.logMenuBuilt(len(menu.Items))
}

// createSaveMenuItem creates a menu item for saving to a specific slot
func (ms *EbitenMenuSystem) createSaveMenuItem(label, saveName, displayName string) MenuItem {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"label":        label,
			"save_name":    saveName,
			"display_name": displayName,
			"enabled":      ms.onSave != nil,
		}).Debug("Creating save menu item")
	}
	return MenuItem{
		Label:   label,
		Enabled: ms.onSave != nil,
		Action: func() error {
			return ms.performSave(saveName, displayName)
		},
	}
}

// performSave executes the save operation with logging and error handling
func (ms *EbitenMenuSystem) performSave(saveName, displayName string) error {
	if ms.onSave == nil {
		if ms.logger != nil {
			ms.logger.Warn("performSave called but no save callback configured")
		}
		return nil
	}
	ms.logSaveStart(saveName)
	if err := ms.onSave(saveName); err != nil {
		ms.logSaveError(saveName, err)
		return fmt.Errorf("save failed: %w", err)
	}
	ms.logSaveSuccess(saveName, displayName)
	return nil
}

// createBackMenuItem creates the back navigation menu item
func (ms *EbitenMenuSystem) createBackMenuItem(menu *MenuComponent) MenuItem {
	if ms.logger != nil {
		ms.logger.Debug("Creating back menu item")
	}
	return MenuItem{
		Label:   "< Back",
		Enabled: true,
		Action: func() error {
			return ms.navigateBack(menu)
		},
	}
}

// navigateBack handles back navigation through menu stack
func (ms *EbitenMenuSystem) navigateBack(menu *MenuComponent) error {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"current_menu": menu.CurrentMenu,
			"stack_depth":  len(menu.MenuStack),
		}).Debug("Navigating back")
	}
	if len(menu.MenuStack) > 0 {
		previousMenu := menu.CurrentMenu
		menu.CurrentMenu = menu.MenuStack[len(menu.MenuStack)-1]
		menu.MenuStack = menu.MenuStack[:len(menu.MenuStack)-1]
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"previous_menu": previousMenu,
				"current_menu":  menu.CurrentMenu,
				"stack_depth":   len(menu.MenuStack),
			}).Debug("Menu navigation completed")
		}
		ms.rebuildMenu(menu)
	}
	return nil
}

// logBuildingMenu logs the start of menu building
func (ms *EbitenMenuSystem) logBuildingMenu() {
	if ms.logger != nil {
		ms.logger.WithField("callback_available", ms.onSave != nil).Debug("Building save menu")
	}
}

// logSaveStart logs the beginning of a save operation
func (ms *EbitenMenuSystem) logSaveStart(saveName string) {
	if ms.logger != nil {
		ms.logger.WithField("save_name", saveName).Info("Saving game")
	}
}

// logSaveError logs a save operation failure
func (ms *EbitenMenuSystem) logSaveError(saveName string, err error) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"save_name": saveName,
			"error":     err.Error(),
		}).Error("Save operation failed")
	}
}

// logSaveSuccess logs a successful save and displays user feedback
func (ms *EbitenMenuSystem) logSaveSuccess(saveName, displayName string) {
	if ms.logger != nil {
		ms.logger.WithField("save_name", saveName).Info("Game saved successfully")
	}
}

// logMenuBuilt logs the completion of menu building
func (ms *EbitenMenuSystem) logMenuBuilt(itemCount int) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"item_count":     itemCount,
			"selected_index": 0,
		}).Debug("Save menu built")
	}
}

// buildLoadMenu constructs the load game menu with existing saves.
func (ms *EbitenMenuSystem) buildLoadMenu(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.WithField("callback_available", ms.onLoad != nil).Debug("Building load menu")
	}
	menu.Items = []MenuItem{}

	saves, err := ms.saveManager.ListSaves()
	if err != nil {
		ms.addErrorMenuItem(menu, err)
	} else {
		ms.addSaveMenuItems(menu, saves)
	}

	ms.addBackMenuItem(menu)
	menu.SelectedIndex = 0
	ms.logLoadMenuBuilt(len(menu.Items), menu.SelectedIndex)
}

// addErrorMenuItem adds error message to menu when save listing fails.
func (ms *EbitenMenuSystem) addErrorMenuItem(menu *MenuComponent, err error) {
	if ms.logger != nil {
		ms.logger.WithField("error", err.Error()).Error("Failed to list save files")
	}
	menu.Items = append(menu.Items, MenuItem{
		Label:   fmt.Sprintf("Error loading saves: %v", err),
		Enabled: false,
	})
	if ms.logger != nil {
		ms.logger.WithField("item_count", len(menu.Items)).Debug("Error menu item added")
	}
}

// addSaveMenuItems adds save file entries to load menu.
func (ms *EbitenMenuSystem) addSaveMenuItems(menu *MenuComponent, saves []*saveload.SaveMetadata) {
	if ms.logger != nil {
		ms.logger.WithField("save_count", len(saves)).Debug("Retrieved save file list")
	}

	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	if ms.logger != nil {
		ms.logger.WithField("save_count", len(saves)).Debug("Save files sorted by timestamp")
	}

	for _, save := range saves {
		menu.Items = append(menu.Items, ms.createLoadMenuItem(*save, menu))
	}

	if len(menu.Items) == 0 {
		ms.addNoSavesMenuItem(menu)
	}
}

// createLoadMenuItem creates menu item for individual save file.
func (ms *EbitenMenuSystem) createLoadMenuItem(save saveload.SaveMetadata, menu *MenuComponent) MenuItem {
	saveName := save.Name
	saveInfo := fmt.Sprintf("%s - Level %d (%s)", save.Name, save.PlayerLevel, save.GenreID)

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"save_name":    saveName,
			"player_level": save.PlayerLevel,
			"genre_id":     save.GenreID,
		}).Debug("Creating load menu item")
	}

	return MenuItem{
		Label:    saveInfo,
		Enabled:  ms.onLoad != nil,
		Metadata: saveName,
		Action:   ms.createLoadAction(saveName, menu),
	}
}

// createLoadAction creates load callback for save file.
func (ms *EbitenMenuSystem) createLoadAction(saveName string, menu *MenuComponent) func() error {
	return func() error {
		if ms.onLoad == nil {
			if ms.logger != nil {
				ms.logger.Warn("Load action called but no load callback configured")
			}
			return nil
		}

		if ms.logger != nil {
			ms.logger.WithField("save_name", saveName).Info("Loading game")
		}

		if err := ms.onLoad(saveName); err != nil {
			if ms.logger != nil {
				ms.logger.WithFields(logrus.Fields{
					"save_name": saveName,
					"error":     err.Error(),
				}).Error("Load operation failed")
			}
			return fmt.Errorf("load failed: %w", err)
		}

		if ms.logger != nil {
			ms.logger.WithField("save_name", saveName).Info("Game loaded successfully")
		}
		menu.ErrorMessage = "Game loaded!"
		menu.ErrorTimeout = 2.0
		menu.Active = false
		if ms.logger != nil {
			ms.logger.WithFields(logrus.Fields{
				"active":        menu.Active,
				"error_message": menu.ErrorMessage,
				"error_timeout": menu.ErrorTimeout,
			}).Debug("Menu state updated after successful load")
		}
		return nil
	}
}

// addNoSavesMenuItem adds placeholder when no saves exist.
func (ms *EbitenMenuSystem) addNoSavesMenuItem(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.Warn("No save files found")
	}
	menu.Items = append(menu.Items, MenuItem{
		Label:   "No save files found",
		Enabled: false,
	})
}

// logLoadMenuBuilt logs menu construction completion.
func (ms *EbitenMenuSystem) logLoadMenuBuilt(itemCount, selectedIndex int) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"item_count":     itemCount,
			"selected_index": selectedIndex,
		}).Debug("Load menu built")
	}
}

// addBackMenuItem adds back navigation button to menu.
func (ms *EbitenMenuSystem) addBackMenuItem(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.Debug("Adding back navigation item")
	}
	menu.Items = append(menu.Items, MenuItem{
		Label:   "< Back",
		Enabled: true,
		Action: func() error {
			if len(menu.MenuStack) > 0 {
				menu.CurrentMenu = menu.MenuStack[len(menu.MenuStack)-1]
				menu.MenuStack = menu.MenuStack[:len(menu.MenuStack)-1]
				ms.rebuildMenu(menu)
			}
			return nil
		},
	})
}

// buildConfirmMenu constructs a confirmation dialog.
func (ms *EbitenMenuSystem) buildConfirmMenu(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.WithField("confirm_message", menu.ConfirmMessage).Debug("Building confirm menu")
	}
	menu.Items = []MenuItem{
		{
			Label:   "Yes",
			Enabled: true,
			Action: func() error {
				if menu.ConfirmAction != nil {
					return menu.ConfirmAction()
				}
				return nil
			},
		},
		{
			Label:   "No",
			Enabled: true,
			Action: func() error {
				// Go back to previous menu
				if len(menu.MenuStack) > 0 {
					menu.CurrentMenu = menu.MenuStack[len(menu.MenuStack)-1]
					menu.MenuStack = menu.MenuStack[:len(menu.MenuStack)-1]
					ms.rebuildMenu(menu)
				}
				return nil
			},
		},
	}
	menu.SelectedIndex = 1 // Default to "No"
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"item_count":     len(menu.Items),
			"selected_index": menu.SelectedIndex,
		}).Debug("Confirm menu built")
	}
}

// rebuildMenu reconstructs the menu based on current menu type.
func (ms *EbitenMenuSystem) rebuildMenu(menu *MenuComponent) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"current_menu": menu.CurrentMenu,
		}).Debug("Rebuilding menu")
	}
	menu.SelectedIndex = 0
	switch menu.CurrentMenu {
	case MenuTypeMain:
		ms.buildMainMenu(menu)
	case MenuTypeSave:
		ms.buildSaveMenu(menu)
	case MenuTypeLoad:
		ms.buildLoadMenu(menu)
	case MenuTypeConfirm:
		ms.buildConfirmMenu(menu)
	default:
		if ms.logger != nil {
			ms.logger.WithField("menu_type", menu.CurrentMenu).Warn("Unknown menu type in rebuildMenu")
		}
	}
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"current_menu": menu.CurrentMenu,
			"item_count":   len(menu.Items),
		}).Debug("Menu rebuilt")
	}
}

// Draw renders the menu overlay.
// Implements UISystem interface.
func (ms *EbitenMenuSystem) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		if ms.logger != nil {
			ms.logger.Warn("Draw called with invalid screen type")
		}
		return
	}
	if ms.menuEntity == nil {
		return
	}

	menu, ok := ms.menuEntity.GetComponent("menu")
	if !ok {
		return
	}

	// Type assert with safety check
	menuComp, ok := menu.(*MenuComponent)
	if !ok || !menuComp.Active {
		return
	}

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(ms.screenWidth, ms.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	img.DrawImage(overlay, nil)
	img.DrawImage(overlay, nil)

	// Calculate menu position (centered)
	menuWidth := 400
	menuHeight := 300
	menuX := (ms.screenWidth - menuWidth) / 2
	menuY := (ms.screenHeight - menuHeight) / 2

	// Draw menu background
	menuBg := ebiten.NewImage(menuWidth, menuHeight)
	menuBg.Fill(color.RGBA{40, 40, 50, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(menuX), float64(menuY))
	img.DrawImage(menuBg, opts)

	// Draw menu title
	var title string
	switch menuComp.CurrentMenu {
	case MenuTypeMain:
		title = "GAME MENU"
	case MenuTypeSave:
		title = "SAVE GAME"
	case MenuTypeLoad:
		title = "LOAD GAME"
	case MenuTypeConfirm:
		title = "CONFIRM"
	}

	ebitenutil.DebugPrintAt(img, title, menuX+10, menuY+10)

	// Draw confirmation message if present
	if menuComp.CurrentMenu == MenuTypeConfirm && menuComp.ConfirmMessage != "" {
		ebitenutil.DebugPrintAt(img, menuComp.ConfirmMessage, menuX+10, menuY+40)
	}

	// Draw menu items
	itemY := menuY + 70
	for i, item := range menuComp.Items {
		// Highlight selected item
		isSelected := i == menuComp.SelectedIndex

		if isSelected {
			// Draw selection background
			selectionBg := ebiten.NewImage(menuWidth-20, 20)
			selectionBg.Fill(color.RGBA{80, 80, 100, 200})
			bgOpts := &ebiten.DrawImageOptions{}
			bgOpts.GeoM.Translate(float64(menuX+10), float64(itemY))
			img.DrawImage(selectionBg, bgOpts)

			// Draw selection indicator
			ebitenutil.DebugPrintAt(img, ">", menuX+10, itemY)
		}

		// Draw item label (offset for selection indicator)
		// Note: Disabled items should appear grayed out, but ebitenutil doesn't support color
		ebitenutil.DebugPrintAt(img, item.Label, menuX+30, itemY)

		itemY += 25
	}

	// Draw error message if present
	if menuComp.ErrorMessage != "" {
		errorY := menuY + menuHeight - 30
		ebitenutil.DebugPrintAt(img, menuComp.ErrorMessage, menuX+10, errorY)
	}

	// Draw controls hint
	controlsY := menuY + menuHeight - 10
	ebitenutil.DebugPrintAt(img, "WASD/Arrows: Navigate | Enter/Click: Select | ESC: Back", menuX+10, controlsY)
}

// SetActive opens or closes the menu.
// Implements UISystem interface.
func (ms *EbitenMenuSystem) SetActive(active bool) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"active":          active,
			"has_menu_entity": ms.menuEntity != nil,
		}).Debug("SetActive called")
	}

	if active {
		if ms.menuEntity == nil {
			if ms.logger != nil {
				ms.logger.Debug("Activating menu via SetActive")
			}
			ms.Toggle()
		}
	} else {
		if ms.menuEntity != nil {
			entityID := ms.menuEntity.ID
			ms.world.RemoveEntity(entityID)
			ms.menuEntity = nil
			if ms.logger != nil {
				ms.logger.WithField("entity_id", entityID).Info("Menu entity removed")
			}
		} else {
			if ms.logger != nil {
				ms.logger.Debug("SetActive(false) called but no menu entity exists")
			}
		}
	}
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"active":          active,
			"has_menu_entity": ms.menuEntity != nil,
		}).Debug("SetActive completed")
	}
}

// Compile-time check that EbitenMenuSystem implements UISystem
var _ UISystem = (*EbitenMenuSystem)(nil)
