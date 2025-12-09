package profiles

import "fmt"

// Profile represents a firewall configuration profile.
type Profile struct {
	Name        string
	Description string
	Active      bool
	Rules       []string // Rule names belonging to this profile
}

// Validate performs basic profile validation.
func Validate(p Profile) error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	if p.Description == "" {
		return fmt.Errorf("profile description is required")
	}
	return nil
}
