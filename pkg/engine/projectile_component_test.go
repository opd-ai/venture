package engine

import (
	"testing"
)

func TestProjectileComponent_Type(t *testing.T) {
	proj := &ProjectileComponent{}
	if got := proj.Type(); got != "projectile" {
		t.Errorf("Type() = %v, want %v", got, "projectile")
	}
}

func TestProjectileComponent_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		age      float64
		lifetime float64
		want     bool
	}{
		{
			name:     "not expired",
			age:      1.0,
			lifetime: 2.0,
			want:     false,
		},
		{
			name:     "expired exactly",
			age:      2.0,
			lifetime: 2.0,
			want:     true,
		},
		{
			name:     "expired past",
			age:      3.0,
			lifetime: 2.0,
			want:     true,
		},
		{
			name:     "zero age",
			age:      0.0,
			lifetime: 1.0,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &ProjectileComponent{
				Age:      tt.age,
				LifeTime: tt.lifetime,
			}
			if got := proj.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectileComponent_CanPierce(t *testing.T) {
	tests := []struct {
		name   string
		pierce int
		want   bool
	}{
		{
			name:   "no pierce",
			pierce: 0,
			want:   false,
		},
		{
			name:   "can pierce once",
			pierce: 1,
			want:   true,
		},
		{
			name:   "can pierce multiple",
			pierce: 3,
			want:   true,
		},
		{
			name:   "infinite pierce",
			pierce: -1,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &ProjectileComponent{Pierce: tt.pierce}
			if got := proj.CanPierce(); got != tt.want {
				t.Errorf("CanPierce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectileComponent_DecrementPierce(t *testing.T) {
	tests := []struct {
		name          string
		pierce        int
		wantDestroy   bool
		wantPierceVal int
	}{
		{
			name:          "no pierce - destroy",
			pierce:        0,
			wantDestroy:   true,
			wantPierceVal: -1,
		},
		{
			name:          "one pierce - continue",
			pierce:        1,
			wantDestroy:   false,
			wantPierceVal: 0,
		},
		{
			name:          "multiple pierce - continue",
			pierce:        3,
			wantDestroy:   false,
			wantPierceVal: 2,
		},
		{
			name:          "infinite pierce - never destroy",
			pierce:        -1,
			wantDestroy:   false,
			wantPierceVal: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &ProjectileComponent{Pierce: tt.pierce}
			gotDestroy := proj.DecrementPierce()
			if gotDestroy != tt.wantDestroy {
				t.Errorf("DecrementPierce() destroy = %v, want %v", gotDestroy, tt.wantDestroy)
			}
			if proj.Pierce != tt.wantPierceVal {
				t.Errorf("Pierce after decrement = %v, want %v", proj.Pierce, tt.wantPierceVal)
			}
		})
	}
}

func TestProjectileComponent_CanBounce(t *testing.T) {
	tests := []struct {
		name   string
		bounce int
		want   bool
	}{
		{
			name:   "no bounce",
			bounce: 0,
			want:   false,
		},
		{
			name:   "can bounce once",
			bounce: 1,
			want:   true,
		},
		{
			name:   "can bounce multiple",
			bounce: 3,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &ProjectileComponent{Bounce: tt.bounce}
			if got := proj.CanBounce(); got != tt.want {
				t.Errorf("CanBounce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectileComponent_DecrementBounce(t *testing.T) {
	tests := []struct {
		name          string
		bounce        int
		wantDestroy   bool
		wantBounceVal int
	}{
		{
			name:          "no bounce - destroy",
			bounce:        0,
			wantDestroy:   true,
			wantBounceVal: -1,
		},
		{
			name:          "one bounce - continue",
			bounce:        1,
			wantDestroy:   false,
			wantBounceVal: 0,
		},
		{
			name:          "multiple bounce - continue",
			bounce:        3,
			wantDestroy:   false,
			wantBounceVal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &ProjectileComponent{Bounce: tt.bounce}
			gotDestroy := proj.DecrementBounce()
			if gotDestroy != tt.wantDestroy {
				t.Errorf("DecrementBounce() destroy = %v, want %v", gotDestroy, tt.wantDestroy)
			}
			if proj.Bounce != tt.wantBounceVal {
				t.Errorf("Bounce after decrement = %v, want %v", proj.Bounce, tt.wantBounceVal)
			}
		})
	}
}

func TestNewProjectileComponent(t *testing.T) {
	damage := 25.0
	speed := 400.0
	lifetime := 5.0
	projType := "arrow"
	ownerID := uint64(123)

	proj := NewProjectileComponent(damage, speed, lifetime, projType, ownerID)

	if proj.Damage != damage {
		t.Errorf("Damage = %v, want %v", proj.Damage, damage)
	}
	if proj.Speed != speed {
		t.Errorf("Speed = %v, want %v", proj.Speed, speed)
	}
	if proj.LifeTime != lifetime {
		t.Errorf("LifeTime = %v, want %v", proj.LifeTime, lifetime)
	}
	if proj.ProjectileType != projType {
		t.Errorf("ProjectileType = %v, want %v", proj.ProjectileType, projType)
	}
	if proj.OwnerID != ownerID {
		t.Errorf("OwnerID = %v, want %v", proj.OwnerID, ownerID)
	}
	if proj.Age != 0.0 {
		t.Errorf("Age = %v, want 0.0", proj.Age)
	}
	if proj.Pierce != 0 {
		t.Errorf("Pierce = %v, want 0", proj.Pierce)
	}
	if proj.Bounce != 0 {
		t.Errorf("Bounce = %v, want 0", proj.Bounce)
	}
	if proj.Explosive {
		t.Errorf("Explosive = true, want false")
	}
	if proj.HasHit {
		t.Errorf("HasHit = true, want false")
	}
}

func TestNewPiercingProjectile(t *testing.T) {
	damage := 30.0
	speed := 500.0
	lifetime := 3.0
	pierce := 2
	projType := "piercing_arrow"
	ownerID := uint64(456)

	proj := NewPiercingProjectile(damage, speed, lifetime, pierce, projType, ownerID)

	if proj.Pierce != pierce {
		t.Errorf("Pierce = %v, want %v", proj.Pierce, pierce)
	}
	if proj.Damage != damage {
		t.Errorf("Damage = %v, want %v", proj.Damage, damage)
	}
	if !proj.CanPierce() {
		t.Error("CanPierce() = false, want true")
	}
}

func TestNewBouncingProjectile(t *testing.T) {
	damage := 20.0
	speed := 300.0
	lifetime := 4.0
	bounce := 2
	projType := "rubber_bullet"
	ownerID := uint64(789)

	proj := NewBouncingProjectile(damage, speed, lifetime, bounce, projType, ownerID)

	if proj.Bounce != bounce {
		t.Errorf("Bounce = %v, want %v", proj.Bounce, bounce)
	}
	if proj.Damage != damage {
		t.Errorf("Damage = %v, want %v", proj.Damage, damage)
	}
	if !proj.CanBounce() {
		t.Error("CanBounce() = false, want true")
	}
}

func TestNewExplosiveProjectile(t *testing.T) {
	damage := 50.0
	speed := 200.0
	lifetime := 6.0
	explosionRadius := 100.0
	projType := "grenade"
	ownerID := uint64(999)

	proj := NewExplosiveProjectile(damage, speed, lifetime, explosionRadius, projType, ownerID)

	if !proj.Explosive {
		t.Error("Explosive = false, want true")
	}
	if proj.ExplosionRadius != explosionRadius {
		t.Errorf("ExplosionRadius = %v, want %v", proj.ExplosionRadius, explosionRadius)
	}
	if proj.Damage != damage {
		t.Errorf("Damage = %v, want %v", proj.Damage, damage)
	}
}

func TestProjectileComponent_LifecycleScenarios(t *testing.T) {
	t.Run("projectile ages and expires", func(t *testing.T) {
		proj := NewProjectileComponent(25.0, 400.0, 2.0, "arrow", 1)

		// Initially not expired
		if proj.IsExpired() {
			t.Error("New projectile should not be expired")
		}

		// Age to half lifetime
		proj.Age = 1.0
		if proj.IsExpired() {
			t.Error("Projectile at half lifetime should not be expired")
		}

		// Age to exactly lifetime
		proj.Age = 2.0
		if !proj.IsExpired() {
			t.Error("Projectile at exactly lifetime should be expired")
		}
	})

	t.Run("piercing projectile lifecycle", func(t *testing.T) {
		proj := NewPiercingProjectile(30.0, 500.0, 3.0, 2, "piercing_arrow", 1)

		// Can pierce initially
		if !proj.CanPierce() {
			t.Error("Piercing projectile should be able to pierce")
		}

		// First hit - should not destroy
		shouldDestroy := proj.DecrementPierce()
		if shouldDestroy {
			t.Error("First hit should not destroy projectile with 2 pierce")
		}
		if proj.Pierce != 1 {
			t.Errorf("Pierce should be 1 after first hit, got %d", proj.Pierce)
		}

		// Second hit - should not destroy
		shouldDestroy = proj.DecrementPierce()
		if shouldDestroy {
			t.Error("Second hit should not destroy projectile with 1 pierce remaining")
		}
		if proj.Pierce != 0 {
			t.Errorf("Pierce should be 0 after second hit, got %d", proj.Pierce)
		}

		// Third hit - should destroy
		shouldDestroy = proj.DecrementPierce()
		if !shouldDestroy {
			t.Error("Third hit should destroy projectile with 0 pierce")
		}
	})

	t.Run("bouncing projectile lifecycle", func(t *testing.T) {
		proj := NewBouncingProjectile(20.0, 300.0, 4.0, 1, "rubber_bullet", 1)

		// Can bounce initially
		if !proj.CanBounce() {
			t.Error("Bouncing projectile should be able to bounce")
		}

		// First bounce - should not destroy
		shouldDestroy := proj.DecrementBounce()
		if shouldDestroy {
			t.Error("First bounce should not destroy projectile with 1 bounce")
		}

		// Second bounce - should destroy
		shouldDestroy = proj.DecrementBounce()
		if !shouldDestroy {
			t.Error("Second bounce should destroy projectile with 0 bounces")
		}
	})

	t.Run("infinite pierce never decrements", func(t *testing.T) {
		proj := NewPiercingProjectile(40.0, 600.0, 5.0, -1, "magic_bolt", 1)

		// Multiple hits should never destroy
		for i := 0; i < 10; i++ {
			shouldDestroy := proj.DecrementPierce()
			if shouldDestroy {
				t.Errorf("Hit %d should not destroy infinite pierce projectile", i+1)
			}
			if proj.Pierce != -1 {
				t.Errorf("Infinite pierce should remain -1, got %d at hit %d", proj.Pierce, i+1)
			}
		}
	})
}
