package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"

	"gopkg.in/yaml.v3"
)

type Resume struct {
	Header     Header       `yaml:"header"`
	Summary    string       `yaml:"summary"`
	Education  []Education  `yaml:"education"`
	Experience []Experience `yaml:"experience"`
	Projects   []Project    `yaml:"projects"`
}

type Header struct {
	Name  string   `yaml:"name"`
	Email string   `yaml:"email"`
	Links []string `yaml:"links"`
}

type Education struct {
	School string   `yaml:"school"`
	Degree string   `yaml:"degree"`
	Notes  []string `yaml:"notes"`
}

type Group struct {
	Label   string   `yaml:"label"`
	Bullets []string `yaml:"bullets"`
}

type Experience struct {
	Org         string   `yaml:"org"`
	Title       string   `yaml:"title"`
	Subtitle    string   `yaml:"subtitle"`
	Link        string   `yaml:"link"`
	Description string   `yaml:"description"`
	Groups      []Group  `yaml:"groups"`
	Bullets     []string `yaml:"bullets"`
}

type Project struct {
	Name     string   `yaml:"name"`
	Subtitle string   `yaml:"subtitle"`
	Link     string   `yaml:"link"`
	Role     string   `yaml:"role"`
	Bullets  []string `yaml:"bullets"`
}

//go:embed resume.html.tmpl
var tmplFS embed.FS

func render(r *Resume, w io.Writer) error {
	t, err := template.ParseFS(tmplFS, "resume.html.tmpl")
	if err != nil {
		return err
	}
	return t.Execute(w, r)
}

func parseResume(data []byte) (*Resume, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var r Resume
	if err := dec.Decode(&r); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("resume file is empty")
		}
		return nil, err
	}
	if err := validate(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func validate(r *Resume) error {
	if r.Header.Name == "" {
		return fmt.Errorf("header: missing name")
	}
	for i, e := range r.Education {
		if e.School == "" {
			return fmt.Errorf("education[%d]: missing school", i)
		}
	}
	for i, e := range r.Experience {
		if e.Org == "" {
			return fmt.Errorf("experience[%d]: missing org", i)
		}
	}
	for i, p := range r.Projects {
		if p.Name == "" {
			return fmt.Errorf("projects[%d]: missing name", i)
		}
	}
	return nil
}
