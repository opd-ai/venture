package engine

type StubMapUI struct {
UpdateCount int
DrawCount   int
active      bool
}

func NewStubMapUI() *StubMapUI {
return &StubMapUI{}
}

func (s *StubMapUI) Update(entities []*Entity, deltaTime float64) {
s.UpdateCount++
}

func (s *StubMapUI) Draw(screen interface{}) {
s.DrawCount++
}

func (s *StubMapUI) IsActive() bool {
return s.active
}

func (s *StubMapUI) SetActive(active bool) {
s.active = active
}

var _ UISystem = (*StubMapUI)(nil)
