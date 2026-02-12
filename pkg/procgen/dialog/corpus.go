package dialog

// Corpus represents a genre-specific collection of training sentences.
type Corpus struct {
	GenreID   string
	Sentences []string
}

// GetCorpus returns training sentences for a specific genre.
//
// Supported genres:
//   - "fantasy": Medieval fantasy with magic, dungeons, dragons
//   - "scifi": Futuristic technology and space themes
//   - "horror": Dark, ominous, scary atmosphere
//   - "cyberpunk": Urban future with hacking and neon
//   - "postapocalyptic": Survival in wasteland
//
// Returns nil if genreID is not recognized.
func GetCorpus(genreID string) *Corpus {
	switch genreID {
	case "fantasy":
		return GetFantasyCorpus()
	case "scifi":
		return GetSciFiCorpus()
	case "horror":
		return GetHorrorCorpus()
	case "cyberpunk":
		return GetCyberpunkCorpus()
	case "postapocalyptic":
		return GetPostApocalypticCorpus()
	default:
		return nil
	}
}

// GetFantasyCorpus returns medieval fantasy training data.
func GetFantasyCorpus() *Corpus {
	return &Corpus{
		GenreID: "fantasy",
		Sentences: []string{
			// Greetings and common phrases
			"Greetings, traveler.",
			"Hail, brave adventurer!",
			"Well met, stranger.",
			"May the gods watch over you.",
			"What brings you to these parts?",
			"Have you come seeking glory?",
			"The road ahead is perilous.",
			"Dark times have fallen upon our land.",
			"I have seen many warriors pass this way.",
			"Few return from the depths below.",

			// Merchant dialog
			"I have fine wares if you have coin.",
			"These goods came from distant lands.",
			"My prices are fair, I assure you.",
			"A warrior needs sturdy equipment.",
			"This blade has seen many battles.",
			"The armor will serve you well.",
			"I can offer you a special price today.",
			"These potions are freshly brewed.",
			"Every adventurer needs supplies.",
			"Will that be all, my friend?",

			// Quest and lore
			"The ancient dungeon lies beyond the dark forest.",
			"Beware the creatures that dwell within.",
			"Many have sought the legendary artifact.",
			"The dragon has terrorized our village for years.",
			"An evil sorcerer cursed these lands long ago.",
			"The prophecy speaks of a chosen hero.",
			"The king offers great rewards for brave souls.",
			"Rumors tell of treasure hidden in the ruins.",
			"The old tower holds many secrets.",
			"Dark magic corrupts all who touch it.",

			// Warnings and advice
			"Tread carefully in the shadows.",
			"Trust not the whispers in the dark.",
			"Steel your heart against fear.",
			"Magic can be a double-edged sword.",
			"The undead rise when darkness falls.",
			"Ancient guardians protect the sacred temple.",
			"Bring light to banish the shadows.",
			"Fire is the bane of trolls.",
			"Silver cuts through cursed flesh.",
			"Knowledge is power in these lands.",

			// Directions and locations
			"The dungeon entrance is to the north.",
			"Follow the river to the village.",
			"Cross the bridge and turn east.",
			"The castle stands atop the high hill.",
			"Deep caverns run beneath the mountains.",
			"The forest grows thick around the ruins.",
			"A hidden path leads through the valley.",
			"The shrine can be found by the lake.",
			"Ancient roads connect the old kingdoms.",
			"The watchtower overlooks the plains.",

			// Combat and danger
			"Monsters grow stronger in the depths.",
			"The guardian will not let you pass easily.",
			"Sharp blades and strong armor save lives.",
			"Magic shields protect against dark forces.",
			"Healing potions restore your strength.",
			"Poison weakens even the mightiest warriors.",
			"Lightning strikes with terrible fury.",
			"Ice can freeze your enemies solid.",
			"The boss awaits in the final chamber.",
			"Victory brings glory and riches.",

			// NPCs and relationships
			"The blacksmith forges excellent weapons.",
			"The healer knows ancient restorative arts.",
			"The wizard studies forbidden tomes.",
			"The priest offers blessings to the faithful.",
			"The rogue trades in secrets and shadows.",
			"The knight upholds honor and justice.",
			"The ranger knows every path through the wilderness.",
			"The bard sings tales of legendary heroes.",
			"The alchemist brews powerful elixirs.",
			"The elder remembers the old ways.",

			// Emotional responses
			"I fear for my family's safety.",
			"Hope flickers like a candle in the wind.",
			"Joy fills my heart at your success!",
			"Sadness weighs heavy upon the land.",
			"Anger burns in the hearts of the oppressed.",
			"Courage will see you through the darkness.",
			"Despair has claimed many souls.",
			"Pride comes before the fall.",
			"Gratitude for your heroic deeds!",
			"Wisdom guides the righteous path.",

			// Weather and environment
			"The storm rages outside.",
			"Morning mist clings to the ground.",
			"Twilight shadows lengthen across the land.",
			"Stars shine bright in the clear night sky.",
			"Rain falls like tears from heaven.",
			"Snow blankets the mountain peaks.",
			"The wind howls through the ruins.",
			"Sunlight breaks through the clouds.",
			"Fog obscures the treacherous path.",
			"Thunder echoes in the distance.",

			// Magic and the supernatural
			"Arcane energies pulse through the ley lines.",
			"The veil between worlds grows thin.",
			"Spirits wander the ancient battlefields.",
			"Enchantments ward against evil.",
			"Curses linger for generations.",
			"Rituals must be performed with precision.",
			"The elements answer to those with power.",
			"Divination reveals glimpses of the future.",
			"Necromancy defiles the natural order.",
			"Holy light banishes the undead.",
		},
	}
}

// GetSciFiCorpus returns science fiction training data.
func GetSciFiCorpus() *Corpus {
	return &Corpus{
		GenreID: "scifi",
		Sentences: []string{
			// Greetings and common phrases
			"Greetings, citizen.",
			"Welcome to the station.",
			"Identify yourself, stranger.",
			"What is your business here?",
			"Systems check complete.",
			"All sectors operational.",
			"Initiate communication protocol.",
			"Neural link established.",
			"Data transfer in progress.",
			"Command acknowledged.",

			// Technology and equipment
			"Your cybernetic implants require maintenance.",
			"Weapon systems armed and ready.",
			"Energy shields at maximum capacity.",
			"The plasma rifle is fully charged.",
			"Nanobots can repair most damage.",
			"Quantum processors enhance cognitive function.",
			"Holographic displays show real-time data.",
			"The exosuit amplifies strength tenfold.",
			"Cloaking devices render you invisible.",
			"Teleportation pads connect distant locations.",

			// Corporate and political
			"The corporation controls all major sectors.",
			"Compliance with regulations is mandatory.",
			"Security forces patrol the lower levels.",
			"Profit margins determine resource allocation.",
			"The board of directors makes final decisions.",
			"Rebels hide in the underground networks.",
			"Government surveillance monitors all citizens.",
			"Contracts bind workers to their employers.",
			"Corporate espionage is common practice.",
			"The executive class enjoys special privileges.",

			// Space and exploration
			"The ship departs for deep space tomorrow.",
			"Alien artifacts were discovered on the outer rim.",
			"Warp drives enable faster-than-light travel.",
			"Space stations orbit every major planet.",
			"Asteroid mining provides valuable resources.",
			"The void between stars is cold and empty.",
			"First contact protocols must be followed.",
			"Unknown signals emanate from the nebula.",
			"Colony ships carry thousands of passengers.",
			"The gateway leads to another galaxy.",

			// AI and robotics
			"The artificial intelligence exhibits strange behavior.",
			"Androids perform most manual labor.",
			"Machine learning algorithms optimize everything.",
			"Neural networks process vast amounts of data.",
			"Robots maintain the infrastructure.",
			"The AI has become self-aware.",
			"Automated systems control life support.",
			"Drones patrol the perimeter constantly.",
			"Synthetic beings demand equal rights.",
			"The mainframe contains all knowledge.",

			// Hacking and cyberspace
			"Firewalls protect sensitive data.",
			"Hackers infiltrate secure networks daily.",
			"Ice programs defend against intrusion.",
			"The matrix is a digital battleground.",
			"Virtual reality simulations are indistinguishable from reality.",
			"Code injection exploits system vulnerabilities.",
			"Encryption keeps messages private.",
			"Backdoors provide unauthorized access.",
			"Cyber attacks cripple entire corporations.",
			"The grid connects all networked devices.",

			// Dangers and threats
			"Hostile aliens have been sighted nearby.",
			"Radiation levels exceed safety limits.",
			"The reactor is experiencing critical failure.",
			"Rogue AIs pose significant threats.",
			"Biological weapons are strictly forbidden.",
			"Mutants roam the contaminated zones.",
			"Pirates raid merchant vessels regularly.",
			"The quarantine must not be breached.",
			"Nanoplagues spread through populations rapidly.",
			"Experimental weapons have unpredictable effects.",

			// Missions and objectives
			"Your mission is to retrieve the data core.",
			"Eliminate all hostile targets in the sector.",
			"Establish contact with the research facility.",
			"Defend the outpost against incoming forces.",
			"Investigate the distress signal's origin.",
			"Secure the perimeter before proceeding.",
			"Hack into the mainframe and download files.",
			"Rescue the captured personnel immediately.",
			"Disable the security systems quietly.",
			"Complete the objective within the time limit.",

			// NPCs and factions
			"The scientist works on classified projects.",
			"Engineers maintain critical systems.",
			"Soldiers guard strategic locations.",
			"Pilots navigate through dangerous space.",
			"Medics treat injuries with advanced technology.",
			"Technicians repair damaged equipment.",
			"Smugglers trade in illegal goods.",
			"Mercenaries fight for the highest bidder.",
			"Hackers sell information to various clients.",
			"Diplomats negotiate between warring factions.",

			// Status and conditions
			"Life support systems functioning normally.",
			"Oxygen levels within acceptable range.",
			"Temperature regulation optimal.",
			"Gravity generators stable.",
			"Power reserves at seventy percent.",
			"Containment fields holding.",
			"Structural integrity compromised.",
			"Hull breach detected on deck three.",
			"Emergency protocols activated.",
			"All personnel report to stations.",
		},
	}
}

// GetHorrorCorpus returns horror genre training data.
func GetHorrorCorpus() *Corpus {
	return &Corpus{
		GenreID: "horror",
		Sentences: []string{
			// Ominous greetings
			"Turn back while you still can.",
			"You should not have come here.",
			"The darkness welcomes you.",
			"There is no escape now.",
			"They have been waiting for you.",
			"Do you hear the whispers?",
			"Something watches from the shadows.",
			"The air reeks of death and decay.",
			"Silence is never a good sign here.",
			"Your fate was sealed long ago.",

			// Warnings and fear
			"Fear will be your constant companion.",
			"Madness lurks in every corner.",
			"The walls bleed when night falls.",
			"Screams echo through the halls.",
			"Do not look into the mirrors.",
			"The basement door should stay locked.",
			"Never walk alone after dark.",
			"Trust nothing that you see here.",
			"The dead do not rest in this place.",
			"Sanity is a luxury you cannot afford.",

			// Creatures and monsters
			"Twisted abominations crawl in the darkness.",
			"The creature feeds on fear itself.",
			"Pale hands reach from beneath.",
			"Eyes watch from impossible angles.",
			"Rotting flesh shambles toward you.",
			"The beast knows your deepest terrors.",
			"Inhuman shrieks pierce the night.",
			"Claws scratch against the walls.",
			"Tentacles writhe in the shadows.",
			"The thing wears a human face.",

			// Curses and supernatural
			"An ancient curse plagues this land.",
			"Blood rituals summoned dark powers.",
			"The ritual must not be completed.",
			"Forbidden knowledge corrupts the mind.",
			"Unholy symbols mark the walls.",
			"The sacrifice was not enough.",
			"Spirits of the damned seek revenge.",
			"Reality warps in this accursed place.",
			"Nightmares become flesh and bone.",
			"The veil between worlds has torn.",

			// Locations and atmosphere
			"The mansion holds terrible secrets.",
			"Abandoned hospitals harbor dark things.",
			"Graveyards stir with unnatural life.",
			"The old asylum echoes with suffering.",
			"Crypts hide unspeakable horrors.",
			"The forest grows darker with each step.",
			"Fog conceals lurking dangers.",
			"Ruins mark where evil once dwelt.",
			"The chapel has been desecrated.",
			"Twisted trees claw at the sky.",

			// Gore and violence
			"Blood stains cover every surface.",
			"Flesh hangs from rusted hooks.",
			"Bones litter the ground like leaves.",
			"Internal organs arranged in patterns.",
			"The smell of carnage is overwhelming.",
			"Bodies twisted beyond recognition.",
			"Sharp instruments line the chamber.",
			"Fresh wounds never stop bleeding.",
			"Dismembered remains tell a story.",
			"The torture never truly ends.",

			// Psychological horror
			"You cannot trust your own senses.",
			"Reality fractures around you.",
			"Memories that are not yours surface.",
			"Time loses all meaning here.",
			"Your reflection moves independently.",
			"Voices call your name from nowhere.",
			"The same room loops endlessly.",
			"You have been here before.",
			"Everyone you meet seems familiar.",
			"Dreams bleed into waking life.",

			// Despair and hopelessness
			"Hope died here long ago.",
			"No one survives this place.",
			"Your struggles are meaningless.",
			"The darkness always wins.",
			"Death would be a mercy.",
			"All roads lead to the same end.",
			"You are already lost.",
			"The light fades with each moment.",
			"Despair consumes everything.",
			"There is no salvation here.",

			// Rituals and cults
			"The cult performs forbidden rites.",
			"Chanting rises from below.",
			"Candles burn with unnatural flames.",
			"The circle must not be broken.",
			"Ancient tomes describe dark ceremonies.",
			"The summoning has begun.",
			"Sacrifices appease hungry gods.",
			"The faithful serve terrible masters.",
			"Prophecies speak of your arrival.",
			"The ritual requires fresh blood.",

			// Sensory horror
			"Cold breath touches your neck.",
			"Rotten meat smell fills the air.",
			"Something wet drips from above.",
			"Fingernails scrape across stone.",
			"Wet footsteps follow behind you.",
			"The taste of copper fills your mouth.",
			"Skin crawls with invisible insects.",
			"A heartbeat pounds in the walls.",
			"Silence more frightening than screams.",
			"Vision swims with dark spots.",
		},
	}
}

// GetCyberpunkCorpus returns cyberpunk genre training data.
func GetCyberpunkCorpus() *Corpus {
	return &Corpus{
		GenreID: "cyberpunk",
		Sentences: []string{
			// Street slang and greetings
			"Hey choom, what's the word?",
			"Flatline or payline, your choice.",
			"The street has eyes and ears.",
			"Corporate dogs everywhere these days.",
			"Neon lights hide dark deeds.",
			"You look like you need chrome.",
			"Got any eddies to spare?",
			"The net is burning tonight.",
			"Fixers always take their cut.",
			"Trust is expensive in this city.",

			// Technology and cyberware
			"Neural implants boost your reflexes.",
			"Chrome makes you more than human.",
			"The latest software patch just dropped.",
			"Optical enhancements let you see more.",
			"Dermal plating stops most bullets.",
			"Brain dance experiences feel real.",
			"Synthetic organs replace failing meat.",
			"Cyber arms hit harder than natural.",
			"Smart weapons link to your nervous system.",
			"Interface plugs jack you into the grid.",

			// Corporate dystopia
			"Megacorps own everything worth having.",
			"Wage slaves toil in endless shifts.",
			"The suits make all the rules.",
			"Corporate security shoots first.",
			"Quarterly profits matter more than lives.",
			"Executive towers scrape the polluted sky.",
			"Marketing controls what you think.",
			"Product launches are cultural events.",
			"The middle class is a myth.",
			"Company loyalty is rewarded with scraps.",

			// Hacking and netrunning
			"Ice protects valuable data.",
			"Netrunners dive into cyberspace nightly.",
			"Black ice kills flatliners instantly.",
			"The grid pulses with digital life.",
			"Daemons guard secure systems.",
			"Backdoors sell for good eddies.",
			"Encrypted traffic flows through dark nodes.",
			"Viruses spread faster than thought.",
			"The deep net hides forbidden knowledge.",
			"Jack in and ride the data streams.",

			// Urban environment
			"Rain never stops in this city.",
			"Smog chokes the lower levels.",
			"Towers block out the sun.",
			"Alleys hide from authority.",
			"Markets sell anything for a price.",
			"Gangs control their territories fiercely.",
			"Neon signs advertise false promises.",
			"The streets are always crowded.",
			"Poverty and luxury exist side by side.",
			"The city never sleeps.",

			// Crime and violence
			"Guns are easier to get than food.",
			"Street violence is daily routine.",
			"Assassins work for corporate interests.",
			"Black markets thrive in the shadows.",
			"Kidnapping is a profitable business.",
			"Organ harvesting funds the underworld.",
			"Protection rackets squeeze locals.",
			"Drugs keep the masses compliant.",
			"Murder is just another transaction.",
			"The law protects the rich.",

			// Rebellion and underground
			"Runners move contraband through checkpoints.",
			"The underground resists corporate control.",
			"Punk rockers rage against the machine.",
			"Revolutionaries hide in plain sight.",
			"Anonymous networks spread the truth.",
			"Hacktivists target corrupt systems.",
			"Street art defies authority.",
			"Illegal modifications empower the weak.",
			"The resistance grows in the dark.",
			"Freedom fighters strike from shadows.",

			// NPCs and fixers
			"My fixer has what you need.",
			"The ripperdoc can install anything.",
			"Street samurai sell their skills.",
			"Netrunners know all the backdoors.",
			"Corpo agents watch everything.",
			"Solo mercenaries work alone.",
			"Nomads travel beyond the walls.",
			"Media personalities spin the news.",
			"Rockerboys start riots with songs.",
			"Techies keep the machines running.",

			// Transactions and deals
			"Everything has a price in eddies.",
			"Crypto transfers are untraceable.",
			"The job pays well if you survive.",
			"No refunds on illegal goods.",
			"Information is the most valuable commodity.",
			"The deal goes down at midnight.",
			"Contracts written in code and blood.",
			"Payment on delivery, no exceptions.",
			"The market determines value.",
			"Reputation buys trust on the street.",

			// Atmosphere and mood
			"Holographic ads assault your senses.",
			"Synthetic music pounds from clubs.",
			"The air tastes of ozone and exhaust.",
			"Augmented reality overlays everything.",
			"Flickering lights create shadows.",
			"The stench of humanity is overwhelming.",
			"Sirens wail in the distance constantly.",
			"Drones buzz overhead like insects.",
			"The future arrived but failed to deliver.",
			"Beauty and brutality walk hand in hand.",
		},
	}
}

// GetPostApocalypticCorpus returns post-apocalyptic genre training data.
func GetPostApocalypticCorpus() *Corpus {
	return &Corpus{
		GenreID: "postapocalyptic",
		Sentences: []string{
			// Survival greetings
			"Still breathing out there?",
			"Another survivor in the wasteland.",
			"Didn't think I'd see anyone today.",
			"The world ended but we're still here.",
			"Scavenging keeps us alive.",
			"Trust is hard to come by.",
			"Water is more valuable than gold.",
			"Food grows scarcer every day.",
			"The old world is gone forever.",
			"We adapt or we die.",

			// Environment and dangers
			"Radiation zones glow at night.",
			"Mutants roam the ruins freely.",
			"Toxic rain falls without warning.",
			"The wasteland stretches endlessly.",
			"Dust storms bury everything eventually.",
			"Ruins hide both treasure and death.",
			"The sun burns hotter than before.",
			"Nights grow colder each season.",
			"Contaminated water kills slowly.",
			"The earth remembers the bombs.",

			// Resources and scavenging
			"Canned goods are worth fighting for.",
			"Clean water is rarer than bullets.",
			"Scrap metal builds shelters.",
			"Medicine saves lives but is scarce.",
			"Fuel powers generators and vehicles.",
			"Ammunition determines survival odds.",
			"Seeds promise future harvests.",
			"Tools mean self-sufficiency.",
			"Weapons are necessary evils.",
			"Salvage from the old world sustains us.",

			// Factions and groups
			"Raiders take what they want.",
			"Settlers build new communities.",
			"Nomads wander between locations.",
			"Cults worship strange new gods.",
			"Traders connect isolated groups.",
			"Slavers prey on the weak.",
			"Militia defend their territory.",
			"Scientists search for solutions.",
			"Mutants form their own tribes.",
			"The strong protect the weak, sometimes.",

			// Mutations and creatures
			"Radiation twisted the wildlife.",
			"Two-headed beasts hunt at night.",
			"Giant insects nest in the ruins.",
			"Mutated plants spread toxic spores.",
			"Feral humans lost their minds.",
			"Creatures evolved to survive radiation.",
			"Genetic horrors stalk the wasteland.",
			"Swarms overwhelm lone travelers.",
			"Apex predators rule the food chain.",
			"Nature adapted in terrible ways.",

			// Old world remnants
			"Pre-war bunkers still hold supplies.",
			"Ancient technology sometimes works.",
			"Cities stand as broken monuments.",
			"Highways lead to nowhere now.",
			"Libraries preserve forgotten knowledge.",
			"Vehicles rust in endless rows.",
			"Military bases were hit hardest.",
			"Skyscrapers cast long shadows.",
			"The past haunts every ruin.",
			"Civilization's bones litter the earth.",

			// Survival strategies
			"Always have a backup plan.",
			"Never travel alone if possible.",
			"Check supplies before every journey.",
			"Mark safe paths for others.",
			"Share information, not resources.",
			"Fortify your shelter well.",
			"Learn to fix what breaks.",
			"Ration carefully or starve.",
			"Fight only when necessary.",
			"Hide what you value most.",

			// Hope and despair
			"Some days hope feels foolish.",
			"Maybe things will improve eventually.",
			"The children deserve better futures.",
			"Humanity might rebuild someday.",
			"Each sunrise is a small victory.",
			"Giving up means certain death.",
			"Stories of the old world inspire us.",
			"New growth appears in strange places.",
			"Communities form despite the odds.",
			"Life finds a way forward.",

			// Trade and economics
			"Barter systems replace currency.",
			"Your skills determine your value.",
			"Nothing is free in the wasteland.",
			"Trade routes connect settlements.",
			"Prices fluctuate with scarcity.",
			"Debt is paid with labor or goods.",
			"Markets emerge in safe zones.",
			"Reputation affects trade terms.",
			"Honesty is rare but valuable.",
			"Everyone wants what they lack.",

			// Weather and hazards
			"Acid rain eats through metal.",
			"Sandstorms reduce visibility to zero.",
			"Electrical storms light up the sky.",
			"Fallout drifts on the wind.",
			"Fog hides lurking dangers.",
			"Flash floods sweep through valleys.",
			"Freezing nights kill the unprepared.",
			"Heat waves dry up water sources.",
			"Ash clouds block the sun for days.",
			"The weather grows more hostile yearly.",
		},
	}
}

// GetAllCorpora returns all available genre corpora.
func GetAllCorpora() []*Corpus {
	return []*Corpus{
		GetFantasyCorpus(),
		GetSciFiCorpus(),
		GetHorrorCorpus(),
		GetCyberpunkCorpus(),
		GetPostApocalypticCorpus(),
	}
}

// GetAvailableGenres returns a list of supported genre IDs.
func GetAvailableGenres() []string {
	return []string{
		"fantasy",
		"scifi",
		"horror",
		"cyberpunk",
		"postapoc",
	}
}
