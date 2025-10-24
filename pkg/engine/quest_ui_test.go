package engine

type StubQuestUI struct {
	UpdateCount int
	DrawCount   int
	active      bool
}

func NewStubQuestUI() *StubQuestUI {
	return &StubQuestUI{}
}

func (s *StubQuestUI) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

func (s *StubQuestUI) Draw(screen interface{}) {
	s.DrawCount++
}

func (s *StubQuestUI) IsActive() bool {
	return s.active
}

func (s *StubQuestUI) SetActive(active bool) {
	s.active = active
}

var _ UISystem = (*StubQuestUI)(nil)
