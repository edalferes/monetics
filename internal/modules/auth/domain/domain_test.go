package domain

import "testing"

// --- password policy ---

func TestPasswordPolicy_Validate(t *testing.T) {
	policy := DefaultPasswordPolicy()
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid strong password", "Abcdefg1", nil},
		{"too short", "Ab1", ErrPasswordTooShort},
		{"missing upper", "abcdefg1", ErrPasswordMissingUpper},
		{"missing lower", "ABCDEFG1", ErrPasswordMissingLower},
		{"missing digit", "Abcdefgh", ErrPasswordMissingDigit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policy.Validate(tt.password); got != tt.wantErr {
				t.Fatalf("Validate(%q) = %v, want %v", tt.password, got, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_RequireSpecial(t *testing.T) {
	p := PasswordPolicy{MinLength: 6, RequireSpecial: true}
	if err := p.Validate("abcdef1"); err != ErrPasswordMissingSpecial {
		t.Fatalf("expected ErrPasswordMissingSpecial, got %v", err)
	}
	if err := p.Validate("abcdef!"); err != nil {
		t.Fatalf("expected valid password, got %v", err)
	}
}

// --- user / role / permission ---

func sampleUser() *User {
	return &User{
		ID:       1,
		Username: "alice",
		Roles: []Role{
			{
				Name: "admin",
				Permissions: []Permission{
					{Name: "users.read"},
					{Name: "users.write"},
				},
			},
			{
				Name: "viewer",
				Permissions: []Permission{
					{Name: "reports.read"},
				},
			},
		},
	}
}

func TestUser_HasRole(t *testing.T) {
	u := sampleUser()
	if !u.HasRole("admin") {
		t.Fatal("expected admin role")
	}
	if u.HasRole("missing") {
		t.Fatal("did not expect missing role")
	}
	empty := &User{}
	if empty.HasRole("admin") {
		t.Fatal("empty user should not have any role")
	}
}

func TestUser_HasPermission(t *testing.T) {
	u := sampleUser()
	if !u.HasPermission("users.write") {
		t.Fatal("expected users.write")
	}
	if !u.HasPermission("reports.read") {
		t.Fatal("expected reports.read from viewer role")
	}
	if u.HasPermission("billing.delete") {
		t.Fatal("did not expect billing.delete")
	}
}

func TestRole_HasPermission(t *testing.T) {
	r := &Role{
		Name: "admin",
		Permissions: []Permission{
			{Name: "users.read"},
		},
	}
	if !r.HasPermission("users.read") {
		t.Fatal("expected users.read")
	}
	if r.HasPermission("users.write") {
		t.Fatal("did not expect users.write")
	}
	if (&Role{}).HasPermission("any") {
		t.Fatal("empty role should not have permissions")
	}
}
