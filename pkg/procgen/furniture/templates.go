package furniture

import "sort"

// This file contains furniture template definitions including:
// - 30+ furniture templates across 8 categories
// - Template access functions (GetTemplate, GetAllSubTypes, GetSubTypesByCategory)
// - Material, size, and functional property specifications

// GetTemplate returns the template for a given furniture subtype
func GetTemplate(subType string) *Template {
	templates := getAllTemplates()
	if tmpl, ok := templates[subType]; ok {
		return &tmpl
	}
	return nil
}

// getAllTemplates returns all furniture templates organized by category
func getAllTemplates() map[string]Template {
	return map[string]Template{
		// Seating (5 types)
		"Chair": {
			Type: TypeSeating, SubType: "Chair", BaseName: "Chair",
			MinWidth: 1.0, MaxWidth: 1.0, MinHeight: 2.0, MaxHeight: 3.0, MinDepth: 1.0, MaxDepth: 1.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal, MaterialStone},
			Walkable:         false, Functional: true, DetailComplexity: 1.0,
		},
		"Bench": {
			Type: TypeSeating, SubType: "Bench", BaseName: "Bench",
			MinWidth: 2.0, MaxWidth: 4.0, MinHeight: 1.5, MaxHeight: 2.0, MinDepth: 1.0, MaxDepth: 1.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialStone},
			Walkable:         false, Functional: true, DetailComplexity: 0.8,
		},
		"Stool": {
			Type: TypeSeating, SubType: "Stool", BaseName: "Stool",
			MinWidth: 1.0, MaxWidth: 1.0, MinHeight: 1.5, MaxHeight: 2.0, MinDepth: 1.0, MaxDepth: 1.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal},
			Walkable:         false, Functional: true, DetailComplexity: 0.5,
		},
		"Throne": {
			Type: TypeSeating, SubType: "Throne", BaseName: "Throne",
			MinWidth: 2.0, MaxWidth: 3.0, MinHeight: 4.0, MaxHeight: 5.0, MinDepth: 2.0, MaxDepth: 2.5,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialMetal, MaterialCrystal},
			Walkable:         false, Functional: true, DetailComplexity: 2.5,
		},
		"Couch": {
			Type: TypeSeating, SubType: "Couch", BaseName: "Couch",
			MinWidth: 3.0, MaxWidth: 5.0, MinHeight: 2.0, MaxHeight: 2.5, MinDepth: 2.0, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialFabric},
			Walkable:         false, Functional: true, DetailComplexity: 1.5,
		},

		// Storage (6 types)
		"Chest": {
			Type: TypeStorage, SubType: "Chest", BaseName: "Chest",
			MinWidth: 2.0, MaxWidth: 3.0, MinHeight: 1.5, MaxHeight: 2.0, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal},
			Walkable:         false, Functional: true, BaseCapacity: 20, DetailComplexity: 1.2,
		},
		"Wardrobe": {
			Type: TypeStorage, SubType: "Wardrobe", BaseName: "Wardrobe",
			MinWidth: 3.0, MaxWidth: 4.0, MinHeight: 5.0, MaxHeight: 6.0, MinDepth: 2.0, MaxDepth: 2.5,
			AllowedMaterials: []MaterialType{MaterialWood},
			Walkable:         false, Functional: true, BaseCapacity: 40, DetailComplexity: 1.8,
		},
		"Shelf": {
			Type: TypeStorage, SubType: "Shelf", BaseName: "Shelf",
			MinWidth: 2.0, MaxWidth: 4.0, MinHeight: 3.0, MaxHeight: 4.0, MinDepth: 0.5, MaxDepth: 1.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal, MaterialStone},
			Walkable:         true, Functional: true, BaseCapacity: 10, DetailComplexity: 0.8,
		},
		"Barrel": {
			Type: TypeStorage, SubType: "Barrel", BaseName: "Barrel",
			MinWidth: 1.5, MaxWidth: 2.0, MinHeight: 2.0, MaxHeight: 2.5, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal},
			Walkable:         false, Functional: true, BaseCapacity: 15, DetailComplexity: 1.0,
		},
		"Cabinet": {
			Type: TypeStorage, SubType: "Cabinet", BaseName: "Cabinet",
			MinWidth: 2.0, MaxWidth: 3.0, MinHeight: 3.0, MaxHeight: 4.0, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood},
			Walkable:         false, Functional: true, BaseCapacity: 30, DetailComplexity: 1.5,
		},
		"Crate": {
			Type: TypeStorage, SubType: "Crate", BaseName: "Crate",
			MinWidth: 1.5, MaxWidth: 2.0, MinHeight: 1.5, MaxHeight: 2.0, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood},
			Walkable:         false, Functional: true, BaseCapacity: 12, DetailComplexity: 0.6,
		},

		// Crafting (5 types)
		"Anvil": {
			Type: TypeCrafting, SubType: "Anvil", BaseName: "Anvil",
			MinWidth: 2.0, MaxWidth: 2.5, MinHeight: 2.5, MaxHeight: 3.0, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialMetal, MaterialStone},
			Walkable:         false, Functional: true, DetailComplexity: 1.5,
		},
		"Workbench": {
			Type: TypeCrafting, SubType: "Workbench", BaseName: "Workbench",
			MinWidth: 3.0, MaxWidth: 4.0, MinHeight: 2.5, MaxHeight: 3.0, MinDepth: 2.0, MaxDepth: 2.5,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal},
			Walkable:         false, Functional: true, DetailComplexity: 1.3,
		},
		"Forge": {
			Type: TypeCrafting, SubType: "Forge", BaseName: "Forge",
			MinWidth: 3.0, MaxWidth: 4.0, MinHeight: 3.0, MaxHeight: 4.0, MinDepth: 2.5, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialMetal},
			Walkable:         false, Functional: true, BaseLightLevel: 0.6, DetailComplexity: 2.0,
		},
		"Alchemy Table": {
			Type: TypeCrafting, SubType: "Alchemy Table", BaseName: "Alchemy Table",
			MinWidth: 2.5, MaxWidth: 3.5, MinHeight: 2.5, MaxHeight: 3.0, MinDepth: 2.0, MaxDepth: 2.5,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialStone, MaterialCrystal},
			Walkable:         false, Functional: true, DetailComplexity: 1.8,
		},
		"Enchanting Table": {
			Type: TypeCrafting, SubType: "Enchanting Table", BaseName: "Enchanting Table",
			MinWidth: 2.5, MaxWidth: 3.0, MinHeight: 2.5, MaxHeight: 3.5, MinDepth: 2.5, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialCrystal, MaterialMetal},
			Walkable:         false, Functional: true, BaseLightLevel: 0.4, DetailComplexity: 2.5,
		},

		// Decoration (5 types)
		"Statue": {
			Type: TypeDecoration, SubType: "Statue", BaseName: "Statue",
			MinWidth: 1.5, MaxWidth: 3.0, MinHeight: 4.0, MaxHeight: 6.0, MinDepth: 1.5, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialMetal, MaterialCrystal},
			Walkable:         false, Functional: false, DetailComplexity: 2.0,
		},
		"Painting": {
			Type: TypeDecoration, SubType: "Painting", BaseName: "Painting",
			MinWidth: 2.0, MaxWidth: 4.0, MinHeight: 2.0, MaxHeight: 3.0, MinDepth: 0.2, MaxDepth: 0.5,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialFabric},
			Walkable:         true, Functional: false, DetailComplexity: 1.5,
		},
		"Vase": {
			Type: TypeDecoration, SubType: "Vase", BaseName: "Vase",
			MinWidth: 0.8, MaxWidth: 1.5, MinHeight: 1.5, MaxHeight: 3.0, MinDepth: 0.8, MaxDepth: 1.5,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialCrystal},
			Walkable:         false, Functional: false, DetailComplexity: 1.2,
		},
		"Tapestry": {
			Type: TypeDecoration, SubType: "Tapestry", BaseName: "Tapestry",
			MinWidth: 3.0, MaxWidth: 6.0, MinHeight: 4.0, MaxHeight: 6.0, MinDepth: 0.2, MaxDepth: 0.3,
			AllowedMaterials: []MaterialType{MaterialFabric},
			Walkable:         true, Functional: false, DetailComplexity: 1.8,
		},
		"Plant": {
			Type: TypeDecoration, SubType: "Plant", BaseName: "Plant",
			MinWidth: 1.0, MaxWidth: 2.0, MinHeight: 1.5, MaxHeight: 4.0, MinDepth: 1.0, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood}, // Pot material
			Walkable:         false, Functional: false, DetailComplexity: 1.0,
		},

		// Lighting (4 types)
		"Torch": {
			Type: TypeLighting, SubType: "Torch", BaseName: "Torch",
			MinWidth: 0.5, MaxWidth: 0.8, MinHeight: 2.0, MaxHeight: 3.0, MinDepth: 0.5, MaxDepth: 0.8,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialMetal},
			Walkable:         true, Functional: true, BaseLightLevel: 0.8, DetailComplexity: 0.7,
		},
		"Chandelier": {
			Type: TypeLighting, SubType: "Chandelier", BaseName: "Chandelier",
			MinWidth: 2.0, MaxWidth: 4.0, MinHeight: 3.0, MaxHeight: 5.0, MinDepth: 2.0, MaxDepth: 4.0,
			AllowedMaterials: []MaterialType{MaterialMetal, MaterialCrystal},
			Walkable:         true, Functional: true, BaseLightLevel: 1.0, DetailComplexity: 2.5,
		},
		"Lantern": {
			Type: TypeLighting, SubType: "Lantern", BaseName: "Lantern",
			MinWidth: 0.8, MaxWidth: 1.2, MinHeight: 1.5, MaxHeight: 2.5, MinDepth: 0.8, MaxDepth: 1.2,
			AllowedMaterials: []MaterialType{MaterialMetal},
			Walkable:         false, Functional: true, BaseLightLevel: 0.6, DetailComplexity: 1.0,
		},
		"Crystal Light": {
			Type: TypeLighting, SubType: "Crystal Light", BaseName: "Crystal Light",
			MinWidth: 1.0, MaxWidth: 2.0, MinHeight: 2.0, MaxHeight: 4.0, MinDepth: 1.0, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialCrystal},
			Walkable:         false, Functional: true, BaseLightLevel: 0.9, DetailComplexity: 1.8,
		},

		// Bedding (3 types)
		"Bed": {
			Type: TypeBedding, SubType: "Bed", BaseName: "Bed",
			MinWidth: 3.0, MaxWidth: 4.0, MinHeight: 1.5, MaxHeight: 2.5, MinDepth: 5.0, MaxDepth: 6.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialFabric},
			Walkable:         false, Functional: true, DetailComplexity: 1.5,
		},
		"Hammock": {
			Type: TypeBedding, SubType: "Hammock", BaseName: "Hammock",
			MinWidth: 2.0, MaxWidth: 3.0, MinHeight: 1.0, MaxHeight: 1.5, MinDepth: 4.0, MaxDepth: 5.0,
			AllowedMaterials: []MaterialType{MaterialFabric},
			Walkable:         true, Functional: true, DetailComplexity: 0.8,
		},
		"Bedroll": {
			Type: TypeBedding, SubType: "Bedroll", BaseName: "Bedroll",
			MinWidth: 2.0, MaxWidth: 2.5, MinHeight: 0.5, MaxHeight: 1.0, MinDepth: 4.0, MaxDepth: 5.0,
			AllowedMaterials: []MaterialType{MaterialFabric},
			Walkable:         true, Functional: true, DetailComplexity: 0.5,
		},

		// Tables (4 types)
		"Table": {
			Type: TypeTable, SubType: "Table", BaseName: "Table",
			MinWidth: 2.0, MaxWidth: 4.0, MinHeight: 2.0, MaxHeight: 2.5, MinDepth: 2.0, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialStone, MaterialMetal},
			Walkable:         false, Functional: true, DetailComplexity: 1.0,
		},
		"Desk": {
			Type: TypeTable, SubType: "Desk", BaseName: "Desk",
			MinWidth: 3.0, MaxWidth: 4.0, MinHeight: 2.0, MaxHeight: 2.5, MinDepth: 2.0, MaxDepth: 2.5,
			AllowedMaterials: []MaterialType{MaterialWood},
			Walkable:         false, Functional: true, DetailComplexity: 1.3,
		},
		"Counter": {
			Type: TypeTable, SubType: "Counter", BaseName: "Counter",
			MinWidth: 3.0, MaxWidth: 6.0, MinHeight: 2.5, MaxHeight: 3.0, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialWood, MaterialStone},
			Walkable:         false, Functional: true, DetailComplexity: 1.2,
		},
		"Altar": {
			Type: TypeTable, SubType: "Altar", BaseName: "Altar",
			MinWidth: 2.5, MaxWidth: 4.0, MinHeight: 2.5, MaxHeight: 3.5, MinDepth: 2.0, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialCrystal},
			Walkable:         false, Functional: true, BaseLightLevel: 0.3, DetailComplexity: 2.0,
		},

		// Utility (4 types)
		"Fireplace": {
			Type: TypeUtility, SubType: "Fireplace", BaseName: "Fireplace",
			MinWidth: 3.0, MaxWidth: 5.0, MinHeight: 4.0, MaxHeight: 6.0, MinDepth: 2.0, MaxDepth: 3.0,
			AllowedMaterials: []MaterialType{MaterialStone},
			Walkable:         false, Functional: true, BaseLightLevel: 0.9, DetailComplexity: 2.0,
		},
		"Mirror": {
			Type: TypeUtility, SubType: "Mirror", BaseName: "Mirror",
			MinWidth: 1.5, MaxWidth: 3.0, MinHeight: 3.0, MaxHeight: 5.0, MinDepth: 0.3, MaxDepth: 0.5,
			AllowedMaterials: []MaterialType{MaterialMetal, MaterialCrystal},
			Walkable:         true, Functional: false, DetailComplexity: 1.5,
		},
		"Fountain": {
			Type: TypeUtility, SubType: "Fountain", BaseName: "Fountain",
			MinWidth: 3.0, MaxWidth: 5.0, MinHeight: 2.0, MaxHeight: 4.0, MinDepth: 3.0, MaxDepth: 5.0,
			AllowedMaterials: []MaterialType{MaterialStone, MaterialCrystal},
			Walkable:         false, Functional: true, DetailComplexity: 2.5,
		},
		"Brazier": {
			Type: TypeUtility, SubType: "Brazier", BaseName: "Brazier",
			MinWidth: 1.5, MaxWidth: 2.0, MinHeight: 2.5, MaxHeight: 3.5, MinDepth: 1.5, MaxDepth: 2.0,
			AllowedMaterials: []MaterialType{MaterialMetal, MaterialStone},
			Walkable:         false, Functional: true, BaseLightLevel: 0.7, DetailComplexity: 1.2,
		},
	}
}

// GetAllSubTypes returns a list of all available furniture subtypes in deterministic order
func GetAllSubTypes() []string {
	templates := getAllTemplates()
	subtypes := make([]string, 0, len(templates))
	for subtype := range templates {
		subtypes = append(subtypes, subtype)
	}
	// Sort to ensure deterministic order (map iteration is non-deterministic in Go)
	sort.Strings(subtypes)
	return subtypes
}

// GetSubTypesByCategory returns all subtypes for a given furniture type in deterministic order
func GetSubTypesByCategory(furnitureType FurnitureType) []string {
	templates := getAllTemplates()
	subtypes := make([]string, 0)
	for subtype, tmpl := range templates {
		if tmpl.Type == furnitureType {
			subtypes = append(subtypes, subtype)
		}
	}
	// Sort to ensure deterministic order (map iteration is non-deterministic in Go)
	sort.Strings(subtypes)
	return subtypes
}
