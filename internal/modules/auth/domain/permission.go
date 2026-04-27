package domain

// Permission represents a granular permission, typically named "resource:action".
type Permission struct {
	ID   uint
	Name string
}
