package main

import (
	"strings"
	"testing"
)

const validYAML = `
header:
  name: Jane Doe
  email: jane@example.com
  links:
    - https://example.com
summary: A short summary.
education:
  - school: Example University
    degree: BS in Testing, 2024
    notes:
      - Minor in QA
experience:
  - org: Example Corp
    title: Engineer
    subtitle: Platform Team
    link: https://example.com/platform
    description: Built things.
    groups:
      - label: "Backend:"
        bullets:
          - Did backend work.
          - Did more backend work.
  - org: Example Corp
    title: Intern
    bullets:
      - Did intern work.
projects:
  - name: Widget
    subtitle: Hackathon
    link: https://example.com/widget
    role: Developer
    bullets:
      - Made a widget.
`

func TestParseValid(t *testing.T) {
	r, err := parseResume([]byte(validYAML))
	if err != nil {
		t.Fatalf("parseResume: %v", err)
	}
	if r.Header.Name != "Jane Doe" {
		t.Errorf("Header.Name = %q, want %q", r.Header.Name, "Jane Doe")
	}
	if len(r.Header.Links) != 1 || r.Header.Links[0] != "https://example.com" {
		t.Errorf("Header.Links = %v", r.Header.Links)
	}
	if r.Header.Email != "jane@example.com" {
		t.Errorf("Header.Email = %q", r.Header.Email)
	}
	if r.Summary != "A short summary." {
		t.Errorf("Summary = %q", r.Summary)
	}
	if len(r.Education) != 1 || r.Education[0].Notes[0] != "Minor in QA" {
		t.Errorf("Education = %+v", r.Education)
	}
	if len(r.Experience) != 2 {
		t.Fatalf("len(Experience) = %d, want 2", len(r.Experience))
	}
	if r.Experience[0].Description != "Built things." {
		t.Errorf("Experience[0].Description = %q", r.Experience[0].Description)
	}
	if r.Experience[0].Groups[0].Label != "Backend:" {
		t.Errorf("Groups[0].Label = %q", r.Experience[0].Groups[0].Label)
	}
	if len(r.Experience[0].Groups[0].Bullets) != 2 {
		t.Errorf("Groups[0].Bullets = %v", r.Experience[0].Groups[0].Bullets)
	}
	if r.Experience[1].Bullets[0] != "Did intern work." {
		t.Errorf("Experience[1].Bullets = %v", r.Experience[1].Bullets)
	}
	if r.Projects[0].Role != "Developer" {
		t.Errorf("Projects[0].Role = %q", r.Projects[0].Role)
	}
}

func TestParseUnknownField(t *testing.T) {
	in := "header:\n  name: Jane\nexperience:\n  - org: Example Corp\n    bulets:\n      - typo\n"
	_, err := parseResume([]byte(in))
	if err == nil || !strings.Contains(err.Error(), "bulets") {
		t.Fatalf("want unknown-field error mentioning 'bulets', got: %v", err)
	}
}

func TestParseMissingName(t *testing.T) {
	_, err := parseResume([]byte("header:\n  email: jane@example.com\n"))
	if err == nil || !strings.Contains(err.Error(), "header: missing name") {
		t.Fatalf("want 'header: missing name' error, got: %v", err)
	}
}

func TestParseMissingOrg(t *testing.T) {
	in := "header:\n  name: Jane\nexperience:\n  - title: Engineer\n"
	_, err := parseResume([]byte(in))
	if err == nil || !strings.Contains(err.Error(), "experience[0]: missing org") {
		t.Fatalf("want 'experience[0]: missing org' error, got: %v", err)
	}
}

func TestParseMissingSchool(t *testing.T) {
	in := "header:\n  name: Jane\neducation:\n  - degree: BS\n"
	_, err := parseResume([]byte(in))
	if err == nil || !strings.Contains(err.Error(), "education[0]: missing school") {
		t.Fatalf("want 'education[0]: missing school' error, got: %v", err)
	}
}

func TestParseMissingProjectName(t *testing.T) {
	in := "header:\n  name: Jane\nprojects:\n  - role: Developer\n"
	_, err := parseResume([]byte(in))
	if err == nil || !strings.Contains(err.Error(), "projects[0]: missing name") {
		t.Fatalf("want 'projects[0]: missing name' error, got: %v", err)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := parseResume([]byte(""))
	if err == nil {
		t.Fatal("want error for empty input, got nil")
	}
}
