package rules

import "testing"

func TestValidate_Success(t *testing.T) {
	r := Rule{
		Name:        "web",
		Application: "/usr/bin/app",
		Action:      "allow",
		Protocol:    "tcp",
		Direction:   "outbound",
		Ports:       []int{80, 443},
	}
	if err := Validate(r); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidate_RejectsMissingName(t *testing.T) {
	r := Rule{Application: "app", Action: "allow", Protocol: "tcp", Direction: "outbound", Ports: []int{80}}
	if err := Validate(r); err == nil {
		t.Fatalf("expected error for missing name")
	}
}

func TestValidate_RejectsMissingPortsForTCP(t *testing.T) {
	r := Rule{Name: "x", Application: "app", Action: "allow", Protocol: "tcp", Direction: "outbound"}
	if err := Validate(r); err == nil {
		t.Fatalf("expected error for missing ports")
	}
}

func TestValidate_RejectsInvalidPort(t *testing.T) {
	r := Rule{Name: "x", Application: "app", Action: "allow", Protocol: "tcp", Direction: "outbound", Ports: []int{70000}}
	if err := Validate(r); err == nil {
		t.Fatalf("expected error for invalid port")
	}
}

func TestValidate_ProtocolAnyAllowsNoPorts(t *testing.T) {
	r := Rule{Name: "x", Application: "app", Action: "allow", Protocol: "any", Direction: "outbound"}
	if err := Validate(r); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
