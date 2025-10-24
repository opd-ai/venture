package engine

type StubInventoryUI struct {
UpdateCount int
DrawCount   int
active      bool
}

func NewStubInventoryUI() *StubInventoryUI {
return &StubInventoryUI{}
}

func (s *StubInventoryUI) Update(entities []*Entity, deltaTime float64) {
s.UpdateCount++
}

func (s *StubInventoryUI) Draw(screen interface{}) {
s.DrawCount++
}

func (s *StubInventoryUI) IsActive() bool {
return s.active
}

func (s *StubInventoryUI) SetActive(active bool) {
s.active = active
}

var _ UISystem = (*StubInventoryUI)(nil)
