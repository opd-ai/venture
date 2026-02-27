package branching

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// Generator implements the procgen.Generator interface for story arcs
type Generator struct {
	logger *logrus.Entry
}

// NewGenerator creates a new story arc generator
func NewGenerator() *Generator {
	return &Generator{
		logger: logrus.WithField("system_name", "branching_generator"),
	}
}

// SetLogger sets a custom logger for the generator
func (g *Generator) SetLogger(logger *logrus.Entry) {
	if logger != nil {
		g.logger = logger
	}
}

// Generate creates a procedural story arc
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if err := validateGenerationParams(params); err != nil {
		g.logger.WithFields(logrus.Fields{
			"seed":  seed,
			"depth": params.Depth,
			"error": err.Error(),
		}).Debug("generation parameter validation failed")
		return nil, err
	}

	rng := rand.New(rand.NewSource(seed))
	arc := createStoryArc(seed, rng, params)
	nodeCount := calculateNodeCount(params.Depth)

	startNode := g.generateStartNode(rng, arc.GenreID)
	arc.Nodes[startNode.ID] = startNode
	arc.StartNodeID = startNode.ID

	endingNodes := g.buildStoryGraph(arc, rng, startNode.ID, nodeCount)
	g.ensureStoryEnding(arc, rng, endingNodes)

	g.logger.WithFields(logrus.Fields{
		"seed":       seed,
		"arc_id":     arc.ID,
		"node_count": len(arc.Nodes),
		"endings":    len(arc.Endings),
		"genre_id":   arc.GenreID,
	}).Debug("story arc generated successfully")

	return arc, nil
}

// validateGenerationParams verifies depth is at least 1 to ensure minimum story complexity.
// Depth controls node count: nodeCount = 10 + depth*2 (capped at 20).
func validateGenerationParams(params procgen.GenerationParams) error {
	if params.Depth < 1 {
		return fmt.Errorf("depth must be at least 1, got %d", params.Depth)
	}
	return nil
}

// createStoryArc initializes a new StoryArc with genre-specific procedural content.
// All fields are deterministically generated from the provided RNG.
func createStoryArc(seed int64, rng *rand.Rand, params procgen.GenerationParams) *StoryArc {
	return &StoryArc{
		ID:          generateID(rng, "arc"),
		Title:       generateTitle(rng, params.GenreID),
		Description: generateDescription(rng, params.GenreID),
		GenreID:     params.GenreID,
		Nodes:       make(map[string]*StoryNode),
		Endings:     make(map[string]EndingType),
		Seed:        seed,
	}
}

// calculateNodeCount determines target node count from depth parameter.
// Formula: 10 + depth*2, capped at 20 nodes maximum to prevent overly complex graphs.
func calculateNodeCount(depth int) int {
	nodeCount := 10 + int(float64(depth)*2)
	if nodeCount > 20 {
		nodeCount = 20
	}
	return nodeCount
}

func (g *Generator) buildStoryGraph(arc *StoryArc, rng *rand.Rand, startNodeID string, nodeCount int) []string {
	currentLayer := []string{startNodeID}
	endingNodes := []string{}

	for len(arc.Nodes) < nodeCount {
		nextLayer := []string{}
		for _, nodeID := range currentLayer {
			if len(arc.Nodes) >= nodeCount {
				break
			}
			layer, endings := g.processNodeBranches(arc, rng, nodeID, nodeCount)
			nextLayer = append(nextLayer, layer...)
			endingNodes = append(endingNodes, endings...)
		}

		currentLayer = nextLayer
		if len(currentLayer) == 0 {
			break
		}
	}

	return endingNodes
}

func (g *Generator) processNodeBranches(arc *StoryArc, rng *rand.Rand, nodeID string, nodeCount int) ([]string, []string) {
	parentNode := arc.Nodes[nodeID]
	branchCount := determineBranchCount(rng, parentNode.Type)

	nextLayer := []string{}
	endingNodes := []string{}

	for i := 0; i < branchCount && len(arc.Nodes) < nodeCount; i++ {
		newNode, isEnding := g.createBranchNode(arc, rng, nodeCount)
		arc.Nodes[newNode.ID] = newNode
		g.connectToParent(parentNode, newNode, rng, arc.GenreID)

		if isEnding {
			endingNodes = append(endingNodes, newNode.ID)
		} else {
			nextLayer = append(nextLayer, newNode.ID)
		}
	}

	return nextLayer, endingNodes
}

// determineBranchCount calculates how many child nodes to create based on parent node type.
// Choice nodes create 2-3 branches for meaningful player agency; other types create single continuation.
func determineBranchCount(rng *rand.Rand, nodeType NodeType) int {
	if nodeType == NodeTypeChoice {
		return 2 + rng.Intn(2)
	}
	return 1
}

func (g *Generator) createBranchNode(arc *StoryArc, rng *rand.Rand, nodeCount int) (*StoryNode, bool) {
	if len(arc.Nodes) >= nodeCount-3 {
		newNode := g.generateEndingNode(rng, arc.GenreID)
		endingType := EndingType(rng.Intn(6))
		arc.Endings[newNode.ID] = endingType
		return newNode, true
	}

	nodeTypeRoll := rng.Float64()
	if nodeTypeRoll < 0.5 {
		return g.generateChoiceNode(rng, arc.GenreID), false
	} else if nodeTypeRoll < 0.8 {
		return g.generateEventNode(rng, arc.GenreID), false
	}
	return g.generateConsequenceNode(rng, arc.GenreID), false
}

func (g *Generator) connectToParent(parentNode, newNode *StoryNode, rng *rand.Rand, genreID string) {
	if parentNode.Type == NodeTypeChoice {
		choice := Choice{
			ID:             generateID(rng, "choice"),
			Text:           generateChoiceText(rng, genreID),
			Requirements:   make(map[string]interface{}),
			AlignmentShift: generateAlignmentShift(rng),
			FactionChange:  generateFactionChange(rng, genreID),
			NextNodeID:     newNode.ID,
		}
		parentNode.Choices = append(parentNode.Choices, choice)
	} else {
		parentNode.NextNodeID = newNode.ID
	}
}

func (g *Generator) ensureStoryEnding(arc *StoryArc, rng *rand.Rand, endingNodes []string) {
	if len(endingNodes) > 0 {
		return
	}

	endingNode := g.generateEndingNode(rng, arc.GenreID)
	arc.Nodes[endingNode.ID] = endingNode
	arc.Endings[endingNode.ID] = EndingTypeNeutral

	for _, node := range arc.Nodes {
		if node.NextNodeID == "" && len(node.Choices) == 0 && node.Type != NodeTypeEnding {
			node.NextNodeID = endingNode.ID
		}
	}
}

// Validate checks if the generated story arc is valid
func (g *Generator) Validate(result interface{}) error {
	arc, ok := result.(*StoryArc)
	if !ok {
		err := fmt.Errorf("expected *StoryArc, got %T", result)
		g.logger.WithFields(logrus.Fields{
			"expected_type": "*StoryArc",
			"actual_type":   fmt.Sprintf("%T", result),
			"error":         err.Error(),
		}).Debug("validation failed: type mismatch")
		return err
	}

	if err := validateBasicArcProperties(arc); err != nil {
		g.logger.WithFields(logrus.Fields{
			"arc_id": arc.ID,
			"error":  err.Error(),
		}).Debug("validation failed: basic properties")
		return err
	}

	if err := validateArcNodeReferences(arc); err != nil {
		g.logger.WithFields(logrus.Fields{
			"arc_id": arc.ID,
			"error":  err.Error(),
		}).Debug("validation failed: node references")
		return err
	}

	if err := validateArcConnections(arc); err != nil {
		g.logger.WithFields(logrus.Fields{
			"arc_id": arc.ID,
			"error":  err.Error(),
		}).Debug("validation failed: arc connections")
		return err
	}

	g.logger.WithFields(logrus.Fields{
		"arc_id":     arc.ID,
		"node_count": len(arc.Nodes),
	}).Debug("story arc validated successfully")

	return nil
}

// validateBasicArcProperties checks basic arc properties like ID, nodes count, and endings.
func validateBasicArcProperties(arc *StoryArc) error {
	if arc.ID == "" {
		return fmt.Errorf("arc must have an ID")
	}

	if arc.StartNodeID == "" {
		return fmt.Errorf("arc must have a start node")
	}

	if len(arc.Nodes) < 10 {
		return fmt.Errorf("arc must have at least 10 nodes, got %d", len(arc.Nodes))
	}

	if len(arc.Endings) < 1 {
		return fmt.Errorf("arc must have at least 1 ending, got %d", len(arc.Endings))
	}

	return nil
}

// validateArcNodeReferences verifies start node and ending nodes exist.
func validateArcNodeReferences(arc *StoryArc) error {
	if _, exists := arc.Nodes[arc.StartNodeID]; !exists {
		return fmt.Errorf("start node %s not found in nodes", arc.StartNodeID)
	}

	for endingID := range arc.Endings {
		if _, exists := arc.Nodes[endingID]; !exists {
			return fmt.Errorf("ending node %s not found in nodes", endingID)
		}
	}

	return nil
}

// validateArcConnections verifies all node connections are valid.
func validateArcConnections(arc *StoryArc) error {
	for _, node := range arc.Nodes {
		if node.NextNodeID != "" {
			if _, exists := arc.Nodes[node.NextNodeID]; !exists {
				return fmt.Errorf("node %s references non-existent next node %s", node.ID, node.NextNodeID)
			}
		}

		for _, choice := range node.Choices {
			if _, exists := arc.Nodes[choice.NextNodeID]; !exists {
				return fmt.Errorf("choice %s in node %s references non-existent node %s", choice.ID, node.ID, choice.NextNodeID)
			}
		}
	}

	return nil
}

// Helper functions for generation

func (g *Generator) generateStartNode(rng *rand.Rand, genreID string) *StoryNode {
	return &StoryNode{
		ID:           generateID(rng, "start"),
		Type:         NodeTypeStart,
		Title:        "The Journey Begins",
		Description:  generateStartDescription(rng, genreID),
		Requirements: make(map[string]interface{}),
		Effects:      make(map[string]interface{}),
	}
}

func (g *Generator) generateChoiceNode(rng *rand.Rand, genreID string) *StoryNode {
	return &StoryNode{
		ID:           generateID(rng, "choice"),
		Type:         NodeTypeChoice,
		Title:        generateNodeTitle(rng, genreID, NodeTypeChoice),
		Description:  generateNodeDescription(rng, genreID, NodeTypeChoice),
		Choices:      []Choice{},
		Requirements: make(map[string]interface{}),
		Effects:      make(map[string]interface{}),
	}
}

func (g *Generator) generateEventNode(rng *rand.Rand, genreID string) *StoryNode {
	return &StoryNode{
		ID:           generateID(rng, "event"),
		Type:         NodeTypeEvent,
		Title:        generateNodeTitle(rng, genreID, NodeTypeEvent),
		Description:  generateNodeDescription(rng, genreID, NodeTypeEvent),
		Requirements: make(map[string]interface{}),
		Effects:      make(map[string]interface{}),
	}
}

func (g *Generator) generateConsequenceNode(rng *rand.Rand, genreID string) *StoryNode {
	return &StoryNode{
		ID:           generateID(rng, "consequence"),
		Type:         NodeTypeConsequence,
		Title:        generateNodeTitle(rng, genreID, NodeTypeConsequence),
		Description:  generateNodeDescription(rng, genreID, NodeTypeConsequence),
		Requirements: make(map[string]interface{}),
		Effects:      make(map[string]interface{}),
	}
}

func (g *Generator) generateEndingNode(rng *rand.Rand, genreID string) *StoryNode {
	return &StoryNode{
		ID:           generateID(rng, "ending"),
		Type:         NodeTypeEnding,
		Title:        generateNodeTitle(rng, genreID, NodeTypeEnding),
		Description:  generateNodeDescription(rng, genreID, NodeTypeEnding),
		Requirements: make(map[string]interface{}),
		Effects:      make(map[string]interface{}),
	}
}

func generateID(rng *rand.Rand, prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, rng.Int63())
}

func generateTitle(rng *rand.Rand, genreID string) string {
	prefixes := map[string][]string{
		"fantasy":   {"The Legend of", "The Quest for", "The Tale of", "The Saga of"},
		"scifi":     {"The Mission to", "The Discovery of", "The Expedition to", "The Protocol"},
		"horror":    {"The Curse of", "The Terror of", "The Nightmare of", "The Haunting of"},
		"cyberpunk": {"The Hack of", "The Run on", "The Breach of", "The Code"},
		"postapoc":  {"The Survival of", "The Last", "The Ruins of", "The Wasteland"},
	}

	subjects := map[string][]string{
		"fantasy":   {"the Ancient Artifact", "the Lost Kingdom", "the Dragon's Lair", "the Forbidden Tower"},
		"scifi":     {"Nova Station", "the Quantum Gate", "Alpha Centauri", "the AI Consciousness"},
		"horror":    {"the Abandoned Asylum", "the Blood Moon", "the Forgotten Crypt", "the Dark Entity"},
		"cyberpunk": {"MegaCorp", "the Neural Net", "the Black Market", "the Ghost Protocol"},
		"postapoc":  {"the Vault", "the Wasteland", "the Last City", "the Survivors"},
	}

	prefix := prefixes[genreID]
	if prefix == nil {
		prefix = prefixes["fantasy"]
	}
	subject := subjects[genreID]
	if subject == nil {
		subject = subjects["fantasy"]
	}

	return prefix[rng.Intn(len(prefix))] + " " + subject[rng.Intn(len(subject))]
}

func generateDescription(rng *rand.Rand, genreID string) string {
	templates := map[string][]string{
		"fantasy": {
			"A tale of heroes and magic in a world on the brink of darkness.",
			"Ancient prophecies speak of a chosen one who will restore balance.",
			"The realm faces a threat unlike any it has seen before.",
		},
		"scifi": {
			"Humanity's future hangs in the balance as technology evolves beyond control.",
			"A scientific breakthrough leads to unforeseen consequences.",
			"The boundaries between human and machine blur in this distant future.",
		},
		"horror": {
			"Something dark has awakened, and it hungers for more than just fear.",
			"The line between reality and nightmare grows thin.",
			"An ancient evil stirs in the shadows, waiting for the right moment.",
		},
		"cyberpunk": {
			"In a world of corporate control, one hacker fights for freedom.",
			"The digital realm holds secrets that could change everything.",
			"Megacorporations rule the world, but the underground fights back.",
		},
		"postapoc": {
			"After the fall of civilization, survivors must rebuild from the ashes.",
			"The old world is gone, but humanity's struggle continues.",
			"In the wasteland, only the strong survive, but at what cost?",
		},
	}

	descriptions := templates[genreID]
	if descriptions == nil {
		descriptions = templates["fantasy"]
	}

	return descriptions[rng.Intn(len(descriptions))]
}

func generateStartDescription(rng *rand.Rand, genreID string) string {
	templates := map[string][]string{
		"fantasy":   {"Your journey begins in a quiet village...", "You stand at the gates of destiny..."},
		"scifi":     {"The ship's AI awakens you from cryo-sleep...", "Mission control confirms your arrival..."},
		"horror":    {"You awaken in an unfamiliar place...", "The nightmare begins as you open your eyes..."},
		"cyberpunk": {"You jack into the network...", "The neon-lit streets await..."},
		"postapoc":  {"You emerge from the vault...", "The wasteland stretches before you..."},
	}

	descriptions := templates[genreID]
	if descriptions == nil {
		descriptions = templates["fantasy"]
	}

	return descriptions[rng.Intn(len(descriptions))]
}

func generateNodeTitle(rng *rand.Rand, genreID string, nodeType NodeType) string {
	templates := map[NodeType][]string{
		NodeTypeChoice:      {"A Difficult Decision", "The Crossroads", "A Moral Dilemma", "Choose Your Path"},
		NodeTypeEvent:       {"An Unexpected Turn", "A Revelation", "A Twist of Fate", "The Discovery"},
		NodeTypeConsequence: {"The Aftermath", "Consequences Unfold", "The Price", "Echoes of Choice"},
		NodeTypeEnding:      {"The Final Chapter", "The End", "Destiny Fulfilled", "Journey's End"},
	}

	titles := templates[nodeType]
	return titles[rng.Intn(len(titles))]
}

func generateNodeDescription(rng *rand.Rand, genreID string, nodeType NodeType) string {
	templates := map[string]map[NodeType][]string{
		"fantasy": {
			NodeTypeChoice:      {"You must choose between honor and power.", "The path ahead splits in two."},
			NodeTypeEvent:       {"Ancient magic awakens around you.", "A prophecy comes to pass."},
			NodeTypeConsequence: {"Your choices have shaped the world.", "The consequences are clear."},
			NodeTypeEnding:      {"Your legend will be told for ages.", "The story concludes."},
		},
		"scifi": {
			NodeTypeChoice:      {"The data suggests two possible courses of action.", "Protocol demands a decision."},
			NodeTypeEvent:       {"Sensors detect an anomaly.", "A message arrives from deep space."},
			NodeTypeConsequence: {"The results are in.", "The experiment concludes."},
			NodeTypeEnding:      {"Mission parameters achieved.", "Final log entry recorded."},
		},
		"horror": {
			NodeTypeChoice:      {"Something lurks in the shadows, waiting for your decision.", "The whispers grow louder, demanding a choice."},
			NodeTypeEvent:       {"A chill runs down your spine as something stirs.", "The darkness seems to breathe around you."},
			NodeTypeConsequence: {"The nightmare takes shape from your actions.", "What you have done cannot be undone."},
			NodeTypeEnding:      {"The horror ends, but the scars remain.", "Silence falls at last."},
		},
		"cyberpunk": {
			NodeTypeChoice:      {"The neural interface flickers with options.", "Corps or streets—time to pick a side."},
			NodeTypeEvent:       {"A data spike floods your implants.", "The grid pulses with encrypted signals."},
			NodeTypeConsequence: {"The net remembers everything.", "Your digital footprint echoes across the grid."},
			NodeTypeEnding:      {"Connection terminated.", "You jack out one final time."},
		},
		"postapoc": {
			NodeTypeChoice:      {"Resources are scarce. Every choice counts.", "The wasteland offers no easy answers."},
			NodeTypeEvent:       {"Dust storms reveal something long buried.", "A caravan appears on the horizon."},
			NodeTypeConsequence: {"The fallout from your actions spreads.", "Survivors will remember this day."},
			NodeTypeEnding:      {"The wasteland claims another story.", "A new dawn breaks over the ruins."},
		},
	}

	genreTemplates := templates[genreID]
	if genreTemplates == nil {
		genreTemplates = templates["fantasy"]
	}

	descriptions := genreTemplates[nodeType]
	if len(descriptions) == 0 {
		return "The story continues..."
	}

	return descriptions[rng.Intn(len(descriptions))]
}

func generateChoiceText(rng *rand.Rand, genreID string) string {
	templates := map[string][]string{
		"fantasy":   {"Accept the quest", "Refuse and walk away", "Negotiate for better terms", "Seek counsel from others"},
		"scifi":     {"Proceed with mission", "Abort and return", "Investigate further", "Request backup"},
		"horror":    {"Enter the darkness", "Flee while you can", "Search for weapons", "Call for help"},
		"cyberpunk": {"Jack in", "Disconnect", "Upload the virus", "Trace the signal"},
		"postapoc":  {"Explore the ruins", "Return to shelter", "Scavenge for supplies", "Set up camp"},
	}

	choices := templates[genreID]
	if choices == nil {
		choices = templates["fantasy"]
	}

	return choices[rng.Intn(len(choices))]
}

func generateAlignmentShift(rng *rand.Rand) map[AlignmentAxis]float64 {
	shift := make(map[AlignmentAxis]float64)

	// 50% chance of affecting good/evil
	if rng.Float64() < 0.5 {
		shift[AlignmentGoodEvil] = (rng.Float64()*0.4 - 0.2) // -0.2 to +0.2
	}

	// 50% chance of affecting law/chaos
	if rng.Float64() < 0.5 {
		shift[AlignmentLawChaos] = (rng.Float64()*0.4 - 0.2)
	}

	// 30% chance of affecting honor/dishonor
	if rng.Float64() < 0.3 {
		shift[AlignmentHonorDishonor] = (rng.Float64()*0.4 - 0.2)
	}

	return shift
}

func generateFactionChange(rng *rand.Rand, genreID string) map[string]float64 {
	factions := map[string][]string{
		"fantasy":   {"The Order", "The Dark Guild", "The Merchant's League", "The Arcane Circle"},
		"scifi":     {"Federation", "Rebel Alliance", "Corporate Syndicate", "AI Coalition"},
		"horror":    {"The Survivors", "The Cult", "The Resistance", "The Lost"},
		"cyberpunk": {"MegaCorp", "The Underground", "Net Runners", "Street Gangs"},
		"postapoc":  {"Vault Dwellers", "Raiders", "Traders", "Mutants"},
	}

	factionList := factions[genreID]
	if factionList == nil {
		factionList = factions["fantasy"]
	}

	changes := make(map[string]float64)

	// 50% chance of affecting a faction
	if rng.Float64() < 0.5 {
		faction := factionList[rng.Intn(len(factionList))]
		changes[faction] = (rng.Float64()*0.4 - 0.2) // -0.2 to +0.2
	}

	return changes
}
