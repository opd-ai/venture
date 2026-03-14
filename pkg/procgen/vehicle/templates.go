// Package vehicle provides genre-specific vehicle templates.
// This file defines vehicle templates for each genre.
//
// Phase 21.1: Vehicle Foundation
package vehicle

// GetFantasyTemplates returns vehicle templates themed for fantasy genre settings,
// including mounts (horses, griffins), wagons, chariots, and magical flying vehicles.
func GetFantasyTemplates() []VehicleTemplate {
	return []VehicleTemplate{
		// Mounts
		{
			NamePrefix:          "Noble",
			NameSuffix:          "Steed",
			VehicleType:         TypeMount,
			BaseMaxSpeed:        200.0,
			BaseAcceleration:    100.0,
			BaseHandling:        3.0,
			BaseDurability:      100.0,
			BaseFuelCapacity:    100.0,
			BaseCapacity:        1,
			FuelType:            "Stamina",
			BaseColor:           0x8B4513, // Brown
			DescriptionTemplate: "A swift and loyal mount",
		},
		{
			NamePrefix:          "War",
			NameSuffix:          "Horse",
			VehicleType:         TypeMount,
			BaseMaxSpeed:        180.0,
			BaseAcceleration:    90.0,
			BaseHandling:        2.5,
			BaseDurability:      150.0,
			BaseFuelCapacity:    120.0,
			BaseCapacity:        1,
			FuelType:            "Stamina",
			BaseColor:           0x2F4F4F, // Dark slate gray
			DescriptionTemplate: "A heavily armored battle steed",
		},
		{
			NamePrefix:          "Ancient",
			NameSuffix:          "Dragon",
			VehicleType:         TypeGlider,
			BaseMaxSpeed:        300.0,
			BaseAcceleration:    120.0,
			BaseHandling:        2.8,
			BaseDurability:      200.0,
			BaseFuelCapacity:    80.0,
			BaseCapacity:        2,
			FuelType:            "Magic",
			BaseColor:           0xFF4500, // Red-orange
			DescriptionTemplate: "A majestic flying dragon",
		},
		// Carts
		{
			NamePrefix:          "Merchant",
			NameSuffix:          "Wagon",
			VehicleType:         TypeCart,
			BaseMaxSpeed:        120.0,
			BaseAcceleration:    50.0,
			BaseHandling:        1.5,
			BaseDurability:      150.0,
			BaseFuelCapacity:    200.0,
			BaseCapacity:        6,
			FuelType:            "Stamina",
			BaseColor:           0xD2691E, // Chocolate
			DescriptionTemplate: "A sturdy cargo wagon",
		},
		// Boats
		{
			NamePrefix:          "Swift",
			NameSuffix:          "Skiff",
			VehicleType:         TypeBoat,
			BaseMaxSpeed:        160.0,
			BaseAcceleration:    70.0,
			BaseHandling:        2.2,
			BaseDurability:      100.0,
			BaseFuelCapacity:    150.0,
			BaseCapacity:        2,
			FuelType:            "Energy",
			BaseColor:           0x4682B4, // Steel blue
			DescriptionTemplate: "A nimble sailing vessel",
		},
	}
}

// GetSciFiTemplates returns vehicle templates themed for sci-fi genre settings,
// including spacecraft, hover vehicles, mechs, and advanced transporters.
func GetSciFiTemplates() []VehicleTemplate {
	return []VehicleTemplate{
		// Mechs
		{
			NamePrefix:          "Titan",
			NameSuffix:          "Mk-7",
			VehicleType:         TypeMech,
			BaseMaxSpeed:        180.0,
			BaseAcceleration:    70.0,
			BaseHandling:        2.2,
			BaseDurability:      250.0,
			BaseFuelCapacity:    150.0,
			BaseCapacity:        1,
			FuelType:            "Energy",
			BaseColor:           0x708090, // Slate gray
			DescriptionTemplate: "An advanced combat mech",
		},
		// Hovercrafts
		{
			NamePrefix:          "Velocity",
			NameSuffix:          "X-9",
			VehicleType:         TypeCart,
			BaseMaxSpeed:        250.0,
			BaseAcceleration:    110.0,
			BaseHandling:        2.8,
			BaseDurability:      80.0,
			BaseFuelCapacity:    100.0,
			BaseCapacity:        2,
			FuelType:            "Energy",
			BaseColor:           0x00CED1, // Dark turquoise
			DescriptionTemplate: "A high-speed hovercraft",
		},
		// Gliders
		{
			NamePrefix:          "Aero",
			NameSuffix:          "Jet",
			VehicleType:         TypeGlider,
			BaseMaxSpeed:        300.0,
			BaseAcceleration:    130.0,
			BaseHandling:        3.0,
			BaseDurability:      60.0,
			BaseFuelCapacity:    70.0,
			BaseCapacity:        1,
			FuelType:            "Energy",
			BaseColor:           0xFFFFFF, // White
			DescriptionTemplate: "A personal jetpack",
		},
		// Boats
		{
			NamePrefix:          "Deep",
			NameSuffix:          "Sub",
			VehicleType:         TypeBoat,
			BaseMaxSpeed:        140.0,
			BaseAcceleration:    60.0,
			BaseHandling:        1.8,
			BaseDurability:      180.0,
			BaseFuelCapacity:    200.0,
			BaseCapacity:        3,
			FuelType:            "Energy",
			BaseColor:           0x191970, // Midnight blue
			DescriptionTemplate: "A pressurized submarine",
		},
		// Mounts (robotic)
		{
			NamePrefix:          "Cyber",
			NameSuffix:          "Stallion",
			VehicleType:         TypeMount,
			BaseMaxSpeed:        220.0,
			BaseAcceleration:    105.0,
			BaseHandling:        3.2,
			BaseDurability:      130.0,
			BaseFuelCapacity:    110.0,
			BaseCapacity:        1,
			FuelType:            "Energy",
			BaseColor:           0xC0C0C0, // Silver
			DescriptionTemplate: "A robotic mount with AI",
		},
	}
}

// GetHorrorTemplates returns vehicle templates themed for horror genre settings,
// including cursed carriages, hearse vehicles, and eerie transportation.
func GetHorrorTemplates() []VehicleTemplate {
	return []VehicleTemplate{
		{
			NamePrefix:          "Cursed",
			NameSuffix:          "Carriage",
			VehicleType:         TypeCart,
			BaseMaxSpeed:        100.0,
			BaseAcceleration:    45.0,
			BaseHandling:        1.2,
			BaseDurability:      120.0,
			BaseFuelCapacity:    150.0,
			BaseCapacity:        4,
			FuelType:            "Blood",
			BaseColor:           0x8B0000, // Dark red
			DescriptionTemplate: "A haunted vehicle that moves on its own",
		},
		{
			NamePrefix:          "Bone",
			NameSuffix:          "Steed",
			VehicleType:         TypeMount,
			BaseMaxSpeed:        180.0,
			BaseAcceleration:    95.0,
			BaseHandling:        2.7,
			BaseDurability:      80.0,
			BaseFuelCapacity:    90.0,
			BaseCapacity:        1,
			FuelType:            "Souls",
			BaseColor:           0xFFFFFF, // White (bone)
			DescriptionTemplate: "An undead skeletal mount",
		},
	}
}

// GetCyberpunkTemplates returns vehicle templates themed for cyberpunk genre settings,
// including motorcycles, armored cars, drones, and neon-lit street racers.
func GetCyberpunkTemplates() []VehicleTemplate {
	return []VehicleTemplate{
		{
			NamePrefix:          "Street",
			NameSuffix:          "Razor",
			VehicleType:         TypeCart,
			BaseMaxSpeed:        240.0,
			BaseAcceleration:    120.0,
			BaseHandling:        3.5,
			BaseDurability:      70.0,
			BaseFuelCapacity:    80.0,
			BaseCapacity:        1,
			FuelType:            "Fuel",
			BaseColor:           0xFF1493, // Deep pink (neon)
			DescriptionTemplate: "A neon-lit street racing bike",
		},
		{
			NamePrefix:          "Combat",
			NameSuffix:          "Frame",
			VehicleType:         TypeMech,
			BaseMaxSpeed:        170.0,
			BaseAcceleration:    65.0,
			BaseHandling:        2.0,
			BaseDurability:      220.0,
			BaseFuelCapacity:    140.0,
			BaseCapacity:        1,
			FuelType:            "Energy",
			BaseColor:           0x00FFFF, // Cyan
			DescriptionTemplate: "A corporate combat exosuit",
		},
	}
}

// GetPostApocTemplates returns vehicle templates themed for post-apocalyptic settings,
// including scavenged war rigs, dune buggies, and makeshift armored vehicles.
func GetPostApocTemplates() []VehicleTemplate {
	return []VehicleTemplate{
		{
			NamePrefix:          "Wasteland",
			NameSuffix:          "Cruiser",
			VehicleType:         TypeCart,
			BaseMaxSpeed:        150.0,
			BaseAcceleration:    70.0,
			BaseHandling:        1.8,
			BaseDurability:      180.0,
			BaseFuelCapacity:    120.0,
			BaseCapacity:        3,
			FuelType:            "Fuel",
			BaseColor:           0x8B4513, // Saddle brown (rust)
			DescriptionTemplate: "A makeshift armored vehicle",
		},
		{
			NamePrefix:          "Scrap",
			NameSuffix:          "Walker",
			VehicleType:         TypeMech,
			BaseMaxSpeed:        140.0,
			BaseAcceleration:    55.0,
			BaseHandling:        1.6,
			BaseDurability:      200.0,
			BaseFuelCapacity:    130.0,
			BaseCapacity:        1,
			FuelType:            "Fuel",
			BaseColor:           0x696969, // Dim gray
			DescriptionTemplate: "A jury-rigged mechanical suit",
		},
	}
}
