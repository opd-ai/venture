package book

import "fmt"

// loadRecipeGrammar loads grammar rules for recipe books.
func (g *Generator) loadRecipeGrammar(grammar *Grammar, genre string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("recipe_intro", []string{
			"This tome contains the secrets of #recipe_craft# passed down through generations of masters.",
			"Within these pages lie the instructions for creating #recipe_item# of exceptional quality.",
			"Learn the ancient art of #recipe_craft# as practiced by the greatest craftsmen of our age.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Required materials: #recipe_materials# Gather these components before beginning. The quality of materials directly affects the final result.",
			"You will need: #recipe_materials# All components must be fresh and properly prepared. Substitutions are not recommended.",
			"Essential ingredients: #recipe_materials# Handle each component with care, for they are imbued with latent power.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Begin by #recipe_action# Pay close attention to the temperature and timing. #recipe_warning#",
			"Next, #recipe_action# This step requires patience and precision. #recipe_tip#",
			"Carefully #recipe_action# The outcome depends on proper technique here. #recipe_detail#",
		})
		grammar.AddRule("recipe_notes", []string{
			"Master craftsmen recommend #recipe_advice# This will greatly improve your results. With practice, you will develop your own techniques.",
			"Note: #recipe_warning# Many apprentices make this mistake. Learn from their errors and yours will be fewer.",
			"Advanced practitioners suggest #recipe_tip# This knowledge comes from years of experience in the craft.",
		})
		grammar.AddRule("recipe_craft", []string{
			"potion brewing", "weapon smithing", "armor crafting", "enchanting", "alchemy",
		})
		grammar.AddRule("recipe_item", []string{
			"healing elixirs", "enchanted blades", "protective amulets", "powerful potions",
		})
		grammar.AddRule("recipe_materials", []string{
			"dragon scales, moonstone dust, and purified water",
			"iron ore, charcoal, and leather strips",
			"herbs from the sacred grove, crystal vials, and spring water",
			"silver wire, gemstones, and blessed oils",
		})
		grammar.AddRule("recipe_action", []string{
			"heating the mixture to precisely 300 degrees",
			"grinding the components into fine powder",
			"infusing the base with magical energy",
			"combining ingredients in the correct sequence",
		})
		grammar.AddRule("recipe_warning", []string{
			"Overheating will ruin the entire batch.",
			"Never skip this step or the result will be unstable.",
			"Rushing this process leads to failure.",
		})
		grammar.AddRule("recipe_tip", []string{
			"Adding a pinch of salt enhances the effect.",
			"Clockwise stirring produces better results.",
			"Morning dew works better than regular water.",
		})
		grammar.AddRule("recipe_detail", []string{
			"The mixture should turn golden when ready.",
			"You will hear a soft chiming sound at completion.",
			"The aroma indicates proper preparation.",
		})
		grammar.AddRule("recipe_advice", []string{
			"practicing the motions before working with materials",
			"keeping your workspace clean and organized",
			"starting with simpler recipes to build skill",
		})

	case "sci-fi":
		grammar.AddRule("recipe_intro", []string{
			"Technical specifications for #recipe_tech# manufacturing. Clearance Level: #recipe_clearance#",
			"Assembly protocol for #recipe_device# requires precise adherence to specifications.",
			"Fabrication guide for #recipe_product# Unauthorized reproduction is prohibited.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Component list: #recipe_components# All parts must meet ISO-9001 standards. Verify serial numbers before assembly.",
			"Required equipment: #recipe_tools# Ensure proper calibration before beginning. Tolerance: +/- 0.001mm.",
			"Materials needed: #recipe_materials# Source only from approved vendors. Substitutions void warranty.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Step #recipe_step#: #recipe_procedure# Verify output before proceeding. Error threshold: <0.1%.",
			"Procedure #recipe_step#: #recipe_technical# Monitor sensor readings continuously. Abort if parameters exceed safe limits.",
			"Phase #recipe_step#: #recipe_assembly# Quality check required at this stage. Document all measurements.",
		})
		grammar.AddRule("recipe_notes", []string{
			"WARNING: #recipe_safety# Failure to follow safety protocols may result in equipment damage or injury.",
			"Technical note: #recipe_optimization# Performance gains of 15-20% have been observed.",
			"Troubleshooting: #recipe_debug# Refer to error codes in Appendix C for detailed diagnostics.",
		})
		grammar.AddRule("recipe_tech", []string{
			"neural interface modules",
			"quantum processors",
			"antimatter containment units",
			"graviton emitters",
		})
		grammar.AddRule("recipe_clearance", []string{
			"Alpha-3", "Beta-7", "Gamma-2", "Delta-5",
		})
		grammar.AddRule("recipe_device", []string{
			"portable fusion reactor",
			"holographic display matrix",
			"emergency life support system",
			"subspace communication array",
		})
		grammar.AddRule("recipe_product", []string{
			"cybernetic enhancement modules",
			"portable force field generators",
			"molecular assemblers",
			"quantum entanglement communicators",
		})
		grammar.AddRule("recipe_components", []string{
			"Processor Unit (PN: QC-7742), Memory Banks (512 Petabyte), Power Coupling (Type-C)",
			"Optical Fibers (Grade-A), Superconducting Coils, Coolant System (Model RS-9)",
			"Nano-assemblers (Gen-4), Carbon Nanotube Framework, Energy Matrix (Class-7)",
		})
		grammar.AddRule("recipe_tools", []string{
			"molecular welder, precision laser cutter, electron microscope",
			"clean room facility (Class 1000), automated assembly station",
			"3D molecular printer, vacuum chamber, spectrometer",
		})
		grammar.AddRule("recipe_materials", []string{
			"titanium-steel alloy (99.9% purity), optical-grade silicon wafers, superconducting ceramic composites",
			"carbon nanotube sheets, graphene layers, rare earth elements",
			"artificial diamond substrates, metallic hydrogen cells, exotic matter samples",
		})
		grammar.AddRule("recipe_step", []string{
			"1", "2", "3", "4", "5",
		})
		grammar.AddRule("recipe_procedure", []string{
			"Initialize fabrication sequence",
			"Apply molecular bonding agent",
			"Activate quantum alignment protocol",
		})
		grammar.AddRule("recipe_technical", []string{
			"Maintain temperature at 4 Kelvin",
			"Apply electromagnetic field at 50 Tesla",
			"Inject nanobots for assembly",
		})
		grammar.AddRule("recipe_assembly", []string{
			"Connect primary power conduits",
			"Seal containment chamber",
			"Install control interface",
		})
		grammar.AddRule("recipe_safety", []string{
			"High voltage present - do not open while powered",
			"Radiation hazard - use proper shielding",
			"Extreme temperatures - allow cooldown period",
		})
		grammar.AddRule("recipe_optimization", []string{
			"Pre-cooling components improves bond strength",
			"Parallel processing reduces assembly time",
			"Using higher-grade materials extends lifespan",
		})
		grammar.AddRule("recipe_debug", []string{
			"Error code E-401 indicates calibration drift",
			"Intermittent signals suggest loose connections",
			"Power fluctuations may require capacitor replacement",
		})

	case "horror":
		grammar.AddRule("recipe_intro", []string{
			"These forbidden instructions detail the creation of #recipe_cursed# May God forgive what you are about to do.",
			"The ritual for crafting #recipe_unholy# should never be attempted. I write this only so others may understand what I have done.",
			"Dark knowledge: how to make #recipe_forbidden# The price is higher than you know.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"You will need: #recipe_dark_materials# Obtain these under cover of darkness. Some are best left unspoken.",
			"Required components: #recipe_profane# Each one damns you further. There is no turning back after this.",
			"Gather: #recipe_cursed_items# The sources are... unnatural. Do not ask where I found them.",
		})
		grammar.AddRule("recipe_steps", []string{
			"When the moon is dark, #recipe_ritual# #recipe_horror# I can still hear the sounds.",
			"At midnight, #recipe_dark_action# #recipe_dread# My hands shake even now, remembering.",
			"During the witching hour, #recipe_profane_act# #recipe_terror# Some things cannot be unseen.",
		})
		grammar.AddRule("recipe_notes", []string{
			"WARNING: #recipe_warning# I learned this too late. The voices never stop.",
			"Do not, under any circumstances, #recipe_prohibition# Three have died ignoring this. I will be the fourth.",
			"A note of caution: #recipe_madness# They watch through the cracks in reality now.",
		})
		grammar.AddRule("recipe_cursed", []string{
			"items that hunger",
			"weapons that remember",
			"armor that feeds",
			"potions that whisper",
		})
		grammar.AddRule("recipe_unholy", []string{
			"the Blade of Endless Night",
			"the Amulet of Screaming Souls",
			"the Elixir of Undeath",
			"the Crown of Madness",
		})
		grammar.AddRule("recipe_forbidden", []string{
			"things that should not exist",
			"doors to places we cannot comprehend",
			"vessels for entities beyond naming",
			"keys to realms of eternal suffering",
		})
		grammar.AddRule("recipe_dark_materials", []string{
			"grave dust from a murderer's tomb, blood freely given, a name that must not be spoken",
			"bones that walked again, tears of the innocent, iron forged in human suffering",
			"ashes of the unburied, water that never knew sunlight, hair from the condemned",
		})
		grammar.AddRule("recipe_profane", []string{
			"flesh that still moves, eyes that still see, a soul fragment",
			"something that died screaming, proof of betrayal, the last breath",
			"a binding written in pain, symbols that burn to look upon, essence of nightmare",
		})
		grammar.AddRule("recipe_cursed_items", []string{
			"items from murder scenes, relics of atrocity, fragments of broken minds",
			"stolen grave goods, cursed heirlooms, objects of obsession",
			"things touched by tragedy, remnants of ritual sacrifice, vessels of despair",
		})
		grammar.AddRule("recipe_ritual", []string{
			"speak the words that should remain unspoken",
			"perform the gestures learned from dreams",
			"call to things that wait in darkness",
		})
		grammar.AddRule("recipe_horror", []string{
			"Something answered.",
			"Reality cracked.",
			"They came through.",
		})
		grammar.AddRule("recipe_dark_action", []string{
			"combine the components in the old way",
			"trace the symbols in your own blood",
			"open yourself to the void",
		})
		grammar.AddRule("recipe_dread", []string{
			"I felt it watching me.",
			"Time moved strangely.",
			"Part of me stayed there.",
		})
		grammar.AddRule("recipe_profane_act", []string{
			"complete the final binding",
			"seal the pact in flesh",
			"offer the sacrifice",
		})
		grammar.AddRule("recipe_terror", []string{
			"I am changed forever.",
			"It knows my true name.",
			"I see them in mirrors now.",
		})
		grammar.AddRule("recipe_warning", []string{
			"The creation will turn on its maker eventually",
			"Using this brings them closer to our world",
			"Each use feeds something hungry in the dark",
		})
		grammar.AddRule("recipe_prohibition", []string{
			"use this during a full moon",
			"create more than one at a time",
			"speak its name aloud",
		})
		grammar.AddRule("recipe_madness", []string{
			"The process changes you in ways you cannot see",
			"Your dreams will never be your own again",
			"Something follows those who craft these things",
		})

	default:
		grammar.AddRule("recipe_intro", []string{
			"This guide explains how to craft useful items. Follow instructions carefully.",
			"Instructions for creating quality products. Precision is important.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Materials needed: basic components. Gather everything before starting.",
			"Required items: standard supplies. Quality materials produce better results.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Step: Combine materials properly. Follow the sequence exactly.",
			"Next: Process components carefully. Take your time with each stage.",
		})
		grammar.AddRule("recipe_notes", []string{
			"Tips: Practice makes perfect. Don't be discouraged by initial failures.",
			"Remember: Safety first. Work in appropriate conditions.",
		})
	}
}

// loadHistoryGrammar loads grammar rules for historical texts.
func (g *Generator) loadHistoryGrammar(grammar *Grammar, genre string, location string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("%s was built #history_when# by #history_builder# Its purpose was to #history_purpose#", location),
			fmt.Sprintf("In the annals of %s, #history_event# occurred. #history_consequence# The effects are felt even today.", location),
			fmt.Sprintf("The architecture of %s reveals #history_secret# Scholars have debated #history_debate# for centuries.", location),
			fmt.Sprintf("During the #history_era# %s #history_role# Many famous #history_figures# walked these halls.", location),
		})
		grammar.AddRule("history_when", []string{
			"in the First Age",
			"during the reign of King Aldric",
			"after the Dragon Wars",
			"in ancient times",
		})
		grammar.AddRule("history_builder", []string{
			"the archmages of the Crystal Tower",
			"dwarven master craftsmen",
			"elven architects of renown",
			"a forgotten civilization",
		})
		grammar.AddRule("history_purpose", []string{
			"protect against the darkness",
			"house a great magical artifact",
			"serve as a center of learning",
			"defend the realm from invasion",
		})
		grammar.AddRule("history_event", []string{
			"a great battle",
			"a historic treaty signing",
			"a magical catastrophe",
			"a royal coronation",
		})
		grammar.AddRule("history_consequence", []string{
			"The structure was damaged but stood firm.",
			"Power shifted throughout the realm.",
			"New alliances were forged.",
			"The place gained legendary status.",
		})
		grammar.AddRule("history_secret", []string{
			"hidden chambers and passages",
			"magical wards still active today",
			"symbols of an older civilization",
			"evidence of lost technologies",
		})
		grammar.AddRule("history_debate", []string{
			"its true original purpose",
			"the identity of its founders",
			"the source of its power",
			"the meaning of inscriptions",
		})
		grammar.AddRule("history_era", []string{
			"Golden Age",
			"Time of Troubles",
			"Age of Magic",
			"Dark Years",
		})
		grammar.AddRule("history_role", []string{
			"served as the capital",
			"was a center of magical learning",
			"housed the royal court",
			"protected the borders",
		})
		grammar.AddRule("history_figures", []string{
			"heroes and warriors",
			"mages and scholars",
			"kings and queens",
			"legendary craftsmen",
		})

	case "sci-fi":
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("Station %s was constructed in year %s. Initial purpose: %s Mission parameters have evolved significantly.", location, "#history_year#", "#history_mission#"),
			fmt.Sprintf("Historical records show %s experienced #history_incident# in %s. Response protocols: #history_response#", location, "#history_year#"),
			fmt.Sprintf("Technical analysis of %s reveals #history_tech# Modifications made over #history_timespan# improved functionality by #history_improvement#", location),
			fmt.Sprintf("During the #history_period# %s #history_significance# Personnel logs document #history_activity#", location),
		})
		grammar.AddRule("history_year", []string{
			"2156", "2187", "2203", "2245",
		})
		grammar.AddRule("history_mission", []string{
			"deep space research and observation",
			"colonial support and resupply",
			"military defense and surveillance",
			"scientific experimentation",
		})
		grammar.AddRule("history_incident", []string{
			"critical systems failure",
			"first contact with alien intelligence",
			"hull breach in sector 7",
			"AI malfunction event",
		})
		grammar.AddRule("history_response", []string{
			"Emergency evacuation successful, zero casualties.",
			"Diplomatic protocols initiated, contact ongoing.",
			"Repair crews deployed, systems restored within 72 hours.",
			"Manual override engaged, AI systems purged.",
		})
		grammar.AddRule("history_tech", []string{
			"advanced propulsion systems of unknown origin",
			"quantum computing cores predating current technology",
			"modular construction allowing expansion",
			"self-repairing hull materials",
		})
		grammar.AddRule("history_timespan", []string{
			"50 years",
			"three decades",
			"multiple generations",
			"the past century",
		})
		grammar.AddRule("history_improvement", []string{
			"200%",
			"450%",
			"a factor of 10",
			"several orders of magnitude",
		})
		grammar.AddRule("history_period", []string{
			"Colonial Expansion Era",
			"AI Integration Period",
			"Post-Contact Years",
			"Corporate Wars",
		})
		grammar.AddRule("history_significance", []string{
			"served as a strategic waypoint",
			"housed critical research programs",
			"provided emergency haven",
			"acted as a diplomatic neutral zone",
		})
		grammar.AddRule("history_activity", []string{
			"extensive scientific breakthroughs",
			"first contact negotiations",
			"classified military operations",
			"civilian life over generations",
		})

	default:
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("%s has a rich history. Built long ago, it has served many purposes over the years.", location),
			fmt.Sprintf("Historical records indicate %s was important for various reasons. Many significant events occurred here.", location),
			fmt.Sprintf("The story of %s is one of change and adaptation. Each era left its mark on this place.", location),
		})
	}
}
