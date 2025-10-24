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
}

// NewEbitenMenuSystem creates a new menu system.
func NewEbitenMenuSystem(world *World, screenWidth, screenHeight int, saveDir string) (*EbitenMenuSystem, error) {
	saveManager, err := saveload.NewSaveManager(saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize save manager: %w", err)
	}

	return &EbitenMenuSystem{
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		saveManager:  saveManager,
	}, nil
}

// SetSaveCallback sets the callback for save operations.
func (ms *EbitenMenuSystem) SetSaveCallback(callback func(name string) error) {
	ms.onSave = callback
}

// SetLoadCallback sets the callback for load operations.
func (ms *EbitenMenuSystem) SetLoadCallback(callback func(name string) error) {
	ms.onLoad = callback
}

// Toggle opens or closes the main menu.
func (ms *EbitenMenuSystem) Toggle() {
	if ms.menuEntity == nil {
		ms.menuEntity = ms.world.CreateEntity()
		menu := &MenuComponent{
			Active:      true,
			CurrentMenu: MenuTypeMain,
		}
		ms.menuEntity.AddComponent(menu)
		ms.buildMainMenu(menu)
		ms.world.Update(0) // Process entity addition
	} else {
		if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
			menuComp := menu.(*MenuComponent)
			menuComp.Active = !menuComp.Active

			// Rebuild main menu when opening
			if menuComp.Active {
				menuComp.CurrentMenu = MenuTypeMain
				menuComp.MenuStack = nil
				ms.buildMainMenu(menuComp)
			}
		}
	}
}

// IsActive returns true if the menu is currently displayed.
func (ms *EbitenMenuSystem) IsActive() bool {
	if ms.menuEntity == nil {
		return false
	}
	if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
		return menu.(*MenuComponent).Active
	}
	return false
}

// Update processes menu input and state.
func (ms *EbitenMenuSystem) Update(entities []*Entity, deltaTime float64) {
	if ms.menuEntity == nil {
		return
	}

	menu, ok := ms.menuEntity.GetComponent("menu")
	if !ok || !menu.(*MenuComponent).Active {
		return
	}

	menuComp := menu.(*MenuComponent)

	// Update error message timeout
	if menuComp.ErrorTimeout > 0 {
		menuComp.ErrorTimeout -= deltaTime
		if menuComp.ErrorTimeout <= 0 {
			menuComp.ErrorMessage = ""
		}
	}

	// Handle input
	ms.handleInput(menuComp)
}

// handleInput processes keyboard and mouse input for menu navigation.
func (ms *EbitenMenuSystem) handleInput(menu *MenuComponent) {
	// Calculate menu bounds for mouse detection
	menuWidth := 400
	menuHeight := 300
	menuX := (ms.screenWidth - menuWidth) / 2
	menuY := (ms.screenHeight - menuHeight) / 2

	// Mouse input handling
	mouseX, mouseY := ebiten.CursorPosition()
	mouseClicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// Calculate item bounds and handle mouse hover/click
	itemY := menuY + 70
	for i := range menu.Items {
		itemBounds := struct {
			x, y, width, height int
		}{
			x:      menuX + 10,
			y:      itemY,
			width:  menuWidth - 20,
			height: 20,
		}

		// Check if mouse is over this item
		if mouseX >= itemBounds.x && mouseX < itemBounds.x+itemBounds.width &&
			mouseY >= itemBounds.y && mouseY < itemBounds.y+itemBounds.height {
			// Mouse is over this item - highlight it
			menu.SelectedIndex = i

			// Handle click
			if mouseClicked {
				item := menu.Items[i]
				if item.Enabled && item.Action != nil {
					if err := item.Action(); err != nil {
						menu.ErrorMessage = err.Error()
						menu.ErrorTimeout = 3.0
					}
				}
			}
		}

		itemY += 25
	}

	// Keyboard input - Navigate up
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		menu.SelectedIndex--
		if menu.SelectedIndex < 0 {
			menu.SelectedIndex = len(menu.Items) - 1
		}
	}

	// Navigate down
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		menu.SelectedIndex++
		if menu.SelectedIndex >= len(menu.Items) {
			menu.SelectedIndex = 0
		}
	}

	// Select item with keyboard
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if menu.SelectedIndex >= 0 && menu.SelectedIndex < len(menu.Items) {
			item := menu.Items[menu.SelectedIndex]
			if item.Enabled && item.Action != nil {
				if err := item.Action(); err != nil {
					menu.ErrorMessage = err.Error()
					menu.ErrorTimeout = 3.0 // Show error for 3 seconds
				}
			}
		}
	}

	// Back/Cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if len(menu.MenuStack) > 0 {
			// Pop back to previous menu
			menu.CurrentMenu = menu.MenuStack[len(menu.MenuStack)-1]
			menu.MenuStack = menu.MenuStack[:len(menu.MenuStack)-1]
			ms.rebuildMenu(menu)
		} else {
			// Close menu
			menu.Active = false
		}
	}
}

// buildMainMenu constructs the main pause menu.
func (ms *EbitenMenuSystem) buildMainMenu(menu *MenuComponent) {
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
}

// buildSaveMenu constructs the save game menu with available save slots.
func (ms *EbitenMenuSystem) buildSaveMenu(menu *MenuComponent) {
	menu.Items = []MenuItem{
		{
			Label:   "Quick Save (slot 1)",
			Enabled: ms.onSave != nil,
			Action: func() error {
				if ms.onSave != nil {
					if err := ms.onSave("quicksave"); err != nil {
						return fmt.Errorf("save failed: %w", err)
					}
					menu.ErrorMessage = "Game saved to Quick Save!"
					menu.ErrorTimeout = 2.0
				}
				return nil
			},
		},
		{
			Label:   "Auto Save (slot 2)",
			Enabled: ms.onSave != nil,
			Action: func() error {
				if ms.onSave != nil {
					if err := ms.onSave("autosave"); err != nil {
						return fmt.Errorf("save failed: %w", err)
					}
					menu.ErrorMessage = "Game saved to Auto Save!"
					menu.ErrorTimeout = 2.0
				}
				return nil
			},
		},
		{
			Label:   "Save Slot 3",
			Enabled: ms.onSave != nil,
			Action: func() error {
				if ms.onSave != nil {
					if err := ms.onSave("save3"); err != nil {
						return fmt.Errorf("save failed: %w", err)
					}
					menu.ErrorMessage = "Game saved to Slot 3!"
					menu.ErrorTimeout = 2.0
				}
				return nil
			},
		},
		{
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
		},
	}
	menu.SelectedIndex = 0
}

// buildLoadMenu constructs the load game menu with existing saves.
func (ms *EbitenMenuSystem) buildLoadMenu(menu *MenuComponent) {
	menu.Items = []MenuItem{}

	// Get list of saves
	saves, err := ms.saveManager.ListSaves()
	if err != nil {
		menu.Items = append(menu.Items, MenuItem{
			Label:   fmt.Sprintf("Error loading saves: %v", err),
			Enabled: false,
		})
	} else {
		// Sort saves by timestamp (newest first)
		sort.Slice(saves, func(i, j int) bool {
			return saves[i].Timestamp.After(saves[j].Timestamp)
		})

		// Add save entries
		for _, save := range saves {
			saveName := save.Name
			saveInfo := fmt.Sprintf("%s - Level %d (%s)", save.Name, save.PlayerLevel, save.GenreID)

			menu.Items = append(menu.Items, MenuItem{
				Label:    saveInfo,
				Enabled:  ms.onLoad != nil,
				Metadata: saveName,
				Action: func() error {
					if ms.onLoad != nil {
						if err := ms.onLoad(saveName); err != nil {
							return fmt.Errorf("load failed: %w", err)
						}
						menu.ErrorMessage = "Game loaded!"
						menu.ErrorTimeout = 2.0
						menu.Active = false // Close menu after successful load
					}
					return nil
				},
			})
		}

		// If no saves found
		if len(menu.Items) == 0 {
			menu.Items = append(menu.Items, MenuItem{
				Label:   "No save files found",
				Enabled: false,
			})
		}
	}

	// Add back button
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

	menu.SelectedIndex = 0
}

// buildConfirmMenu constructs a confirmation dialog.
func (ms *EbitenMenuSystem) buildConfirmMenu(menu *MenuComponent) {
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
}

// rebuildMenu reconstructs the menu based on current menu type.
func (ms *EbitenMenuSystem) rebuildMenu(menu *MenuComponent) {
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
	}
}

// Draw renders the menu overlay.
// Implements UISystem interface.
func (ms *EbitenMenuSystem) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	if ms.menuEntity == nil {
		return
	}

	menu, ok := ms.menuEntity.GetComponent("menu")
	if !ok || !menu.(*MenuComponent).Active {
		return
	}

	menuComp := menu.(*MenuComponent)

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
	if active {
		if ms.menuEntity == nil {
			ms.Toggle()
		}
	} else {
		if ms.menuEntity != nil {
			ms.world.RemoveEntity(ms.menuEntity.ID)
			ms.menuEntity = nil
		}
	}
}

// Compile-time check that EbitenMenuSystem implements UISystem
var _ UISystem = (*EbitenMenuSystem)(nil)
