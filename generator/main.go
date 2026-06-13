package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"

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
	Notes       []string `yaml:"notes"`
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

func main() {
	in := flag.String("in", "resume.yaml", "input YAML file")
	out := flag.String("out", "index.html", "output HTML file")
	flag.Parse()

	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r, err := parseResume(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *in, err)
		os.Exit(1)
	}
	var buf bytes.Buffer
	if err := render(r, &buf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
