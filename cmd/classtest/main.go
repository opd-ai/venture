package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/class/advanced"
)

func main() {
	listClasses := flag.Bool("list-classes", false, "List all base classes")
	listPrestige := flag.Bool("list-prestige", false, "List all prestige classes")
	listSynergies := flag.Bool("list-synergies", false, "List all class synergies")
	demo := flag.Bool("demo", false, "Run full demonstration")

	flag.Parse()

	if *listClasses {
		printClasses()
		return
	}

	if *listPrestige {
		printPrestigeClasses()
		return
	}

	if *listSynergies {
		printSynergies()
		return
	}

	if *demo {
		runDemo()
		return
	}

	fmt.Println("Advanced Class Customization Test Tool")
	fmt.Println("Usage:")
	fmt.Println("  -list-classes     List all 15 base classes")
	fmt.Println("  -list-prestige    List all 20 prestige classes")
	fmt.Println("  -list-synergies   List all multi-class synergies")
	fmt.Println("  -demo             Run full demonstration")
}

func printClasses() {
	fmt.Println("=== Base Classes (15 total) ===")

	classes := advanced.GetAllClasses()
	for _, class := range classes {
		fmt.Printf("%s: %s\n", class.Name, class.Description)
		fmt.Printf("  Base Stats: HP+%d, Mana+%d, Stamina+%d\n",
			class.BaseStats.Health, class.BaseStats.Mana, class.BaseStats.Stamina)
		fmt.Printf("  STR+%d, DEX+%d, INT+%d, WIS+%d, CHA+%d\n",
			class.BaseStats.Strength, class.BaseStats.Dexterity,
			class.BaseStats.Intelligence, class.BaseStats.Wisdom, class.BaseStats.Charisma)
		fmt.Printf("  DEF+%d, MDEF+%d, Crit+%.1f%%, CritDmg+%.1f%%, Speed+%.1f%%\n\n",
			class.BaseStats.Defense, class.BaseStats.MagicDefense,
			class.BaseStats.CritChance*100, class.BaseStats.CritDamage*100,
			class.BaseStats.Speed*100)
	}
}

func printPrestigeClasses() {
	fmt.Println("=== Prestige Classes (20 total) ===")

	classes := advanced.GetAllPrestigeClasses()
	for _, class := range classes {
		fmt.Printf("%s (Level %d+): %s\n", class.Name, class.Requirements.MinLevel, class.Description)
		if len(class.Requirements.RequiredPrimary) > 0 {
			fmt.Printf("  Requires Primary: %v\n", class.Requirements.RequiredPrimary)
		}
		if len(class.Requirements.RequiredSecondary) > 0 {
			fmt.Printf("  Requires Secondary: %v\n", class.Requirements.RequiredSecondary)
		}
		if class.Requirements.MinPrimaryStat > 0 {
			fmt.Printf("  Min Primary Stat: %d\n", class.Requirements.MinPrimaryStat)
		}
		fmt.Printf("  Bonuses: HP+%d, Mana+%d, STR+%d, DEX+%d, INT+%d, WIS+%d\n",
			class.BaseStats.Health, class.BaseStats.Mana,
			class.BaseStats.Strength, class.BaseStats.Dexterity,
			class.BaseStats.Intelligence, class.BaseStats.Wisdom)
		fmt.Printf("  DEF+%d, MDEF+%d, Crit+%.1f%%, CritDmg+%.1f%%\n\n",
			class.BaseStats.Defense, class.BaseStats.MagicDefense,
			class.BaseStats.CritChance*100, class.BaseStats.CritDamage*100)
	}
}

func printSynergies() {
	fmt.Println("=== Multi-Class Synergies (15 combinations) ===")

	manager := advanced.NewManager()
	synergies := manager.GetAllSynergies()

	for _, synergy := range synergies {
		fmt.Printf("%s: %s + %s\n", synergy.Name, synergy.Primary, synergy.Secondary)
		fmt.Printf("  Bonuses: HP+%d, Mana+%d, Stamina+%d\n",
			synergy.Bonuses.Health, synergy.Bonuses.Mana, synergy.Bonuses.Stamina)
		fmt.Printf("  STR+%d, DEX+%d, INT+%d, WIS+%d, CHA+%d\n",
			synergy.Bonuses.Strength, synergy.Bonuses.Dexterity,
			synergy.Bonuses.Intelligence, synergy.Bonuses.Wisdom, synergy.Bonuses.Charisma)
		fmt.Printf("  Crit+%.1f%%, CritDmg+%.1f%%, Speed+%.1f%%\n\n",
			synergy.Bonuses.CritChance*100, synergy.Bonuses.CritDamage*100,
			synergy.Bonuses.Speed*100)
	}
}

func runDemo() {
	fmt.Println("=== Advanced Class System Demonstration ===")

	manager := advanced.NewManager()
	playerID := "demo_player"

	fmt.Println("Step 1: Create a Warrior character")
	if err := manager.SetPrimaryClass(playerID, advanced.ClassWarrior); err != nil {
		log.Fatal(err)
	}
	if err := manager.SetLevel(playerID, 1); err != nil {
		log.Fatal(err)
	}

	player, _ := manager.GetPlayerClass(playerID)
	fmt.Printf("  Level 1 Warrior created\n")
	fmt.Printf("  Talent points available: %d\n\n", player.TalentPoints.PointsTotal)

	fmt.Println("Step 2: Level up to 10 and allocate talents")
	if err := manager.SetLevel(playerID, 10); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		manager.AllocateTalent(playerID, "warrior_power_strike")
	}
	manager.AllocateTalent(playerID, "warrior_weapon_mastery")
	manager.AllocateTalent(playerID, "warrior_critical_hit")
	manager.AllocateTalent(playerID, "warrior_iron_skin")
	manager.AllocateTalent(playerID, "warrior_vitality")
	manager.AllocateTalent(playerID, "warrior_charge")

	player, _ = manager.GetPlayerClass(playerID)
	fmt.Printf("  Level 10 reached, spent %d/%d talent points\n",
		player.TalentPoints.PointsSpent, player.TalentPoints.PointsTotal)

	stats, _ := manager.CalculateTotalStats(playerID)
	fmt.Printf("  Current stats: HP+%d, STR+%d, DEF+%d, Crit+%.1f%%\n\n",
		stats.Health, stats.Strength, stats.Defense, stats.CritChance*100)

	fmt.Println("Step 3: Add secondary class (Multi-classing)")
	if err := manager.SetSecondaryClass(playerID, advanced.ClassMage); err != nil {
		log.Fatal(err)
	}

	stats, _ = manager.CalculateTotalStats(playerID)
	fmt.Printf("  Multiclassed as Warrior/Mage (Spellsword)\n")
	fmt.Printf("  New stats: HP+%d, Mana+%d, STR+%d, INT+%d\n\n",
		stats.Health, stats.Mana, stats.Strength, stats.Intelligence)

	fmt.Println("Step 4: Level up to 25 and unlock prestige class")
	if err := manager.SetLevel(playerID, 25); err != nil {
		log.Fatal(err)
	}
	if err := manager.SetPrestigeClass(playerID, advanced.PrestigeBladeMaster); err != nil {
		log.Fatal(err)
	}

	stats, _ = manager.CalculateTotalStats(playerID)
	fmt.Printf("  Prestige class unlocked: Blade Master\n")
	fmt.Printf("  Final stats: HP+%d, Mana+%d, STR+%d, INT+%d\n",
		stats.Health, stats.Mana, stats.Strength, stats.Intelligence)
	fmt.Printf("  DEF+%d, MDEF+%d, Crit+%.1f%%, CritDmg+%.1f%%\n\n",
		stats.Defense, stats.MagicDefense, stats.CritChance*100, stats.CritDamage*100)

	fmt.Println("Step 5: View talent tree")
	tree, _ := manager.GetTalentTree(advanced.ClassWarrior)
	fmt.Printf("  %s has %d talents:\n", tree.Name,
		len(tree.Offensive)+len(tree.Defensive)+len(tree.Utility))
	fmt.Printf("    Offensive: %d talents\n", len(tree.Offensive))
	fmt.Printf("    Defensive: %d talents\n", len(tree.Defensive))
	fmt.Printf("    Utility: %d talents\n\n", len(tree.Utility))

	fmt.Println("Sample Offensive Talents:")
	for i := 0; i < 3 && i < len(tree.Offensive); i++ {
		talent := tree.Offensive[i]
		fmt.Printf("  - %s (Max Rank %d): %s\n", talent.Name, talent.MaxRank, talent.Description)
	}
	fmt.Println()

	fmt.Println("Step 6: Respec demonstration")
	cost := manager.GetRespecCost(playerID)
	fmt.Printf("  Respec cost: %d gold\n", cost)

	if err := manager.RespecTalents(playerID, cost); err != nil {
		log.Fatal(err)
	}

	player, _ = manager.GetPlayerClass(playerID)
	fmt.Printf("  Talents reset! Points spent: %d (was 10)\n", player.TalentPoints.PointsSpent)
	fmt.Printf("  Available points: %d\n", player.TalentPoints.PointsTotal-player.TalentPoints.PointsSpent)

	newCost := manager.GetRespecCost(playerID)
	fmt.Printf("  Next respec cost: %d gold (+%d)\n\n", newCost, newCost-cost)

	fmt.Println("=== Demonstration Complete ===")
	fmt.Println("Summary:")
	fmt.Println("  ✓ 15 base classes available")
	fmt.Println("  ✓ 20 prestige classes unlocked at level 20+")
	fmt.Println("  ✓ Multi-classing with synergy bonuses")
	fmt.Println("  ✓ 30 talents per class (10 offensive, 10 defensive, 10 utility)")
	fmt.Println("  ✓ Talent respec system with escalating costs")
	fmt.Println("  ✓ Comprehensive stat calculation from all sources")
}
