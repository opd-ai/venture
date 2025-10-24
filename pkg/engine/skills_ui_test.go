package engine

type StubSkillsUI struct {
	UpdateCount int
	DrawCount   int
	active      bool
}

func NewStubSkillsUI() *StubSkillsUI {
	return &StubSkillsUI{}
}

func (s *StubSkillsUI) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

func (s *StubSkillsUI) Draw(screen interface{}) {
	s.DrawCount++
}

func (s *StubSkillsUI) IsActive() bool {
	return s.active
}

func (s *StubSkillsUI) SetActive(active bool) {
	s.active = active
}

var _ UISystem = (*StubSkillsUI)(nil)
