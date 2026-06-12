# YAML → HTML Resume Generator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hand-written `index.html` with a `resume.yaml` source file that a Go generator compiles into the same HTML page at Docker image build time.

**Architecture:** A single Go program (`generator/`) decodes `resume.yaml` with strict field checking, validates required fields, and renders an embedded `html/template` that reproduces the current page layout. A multi-stage Dockerfile runs the generator in stage 1 and copies the output into the existing nginx image in stage 2. `index.html` is deleted from the repo once the generated output is verified against it.

**Tech Stack:** Go (stdlib `html/template`, `embed`), `gopkg.in/yaml.v3`, Docker multi-stage build, nginx (unchanged).

**Spec:** `docs/superpowers/specs/2026-06-12-yaml-resume-generator-design.md`

**Repo:** `/home/eddison/projects/resume_website` (all paths below are relative to this).

**Commit style note:** This repo's owner requires commit messages WITHOUT `Co-Authored-By` lines.

## File structure

| File | Responsibility |
|---|---|
| `go.mod`, `go.sum` | Go module (`github.com/EddisonSo/resume-website`), yaml.v3 dependency |
| `generator/main.go` | Types, strict YAML parsing, validation, template rendering, CLI entry |
| `generator/main_test.go` | Parse/validation unit tests + golden-file render test |
| `generator/resume.html.tmpl` | HTML template (embedded via `go:embed`), reproduces current look |
| `generator/testdata/sample.yaml` | Synthetic fixture exercising every schema feature |
| `generator/testdata/golden.html` | Expected render of `sample.yaml` (generated with `-update`, then reviewed) |
| `resume.yaml` | Real resume content, transcribed verbatim from current `index.html` |
| `Dockerfile` | Multi-stage: generator build → nginx |
| `index.html` | DELETED in Task 5 after verification |

---

### Task 1: Go module scaffold

**Files:**
- Create: `go.mod`, `go.sum`

- [ ] **Step 1: Initialize module and fetch yaml.v3**

```bash
cd /home/eddison/projects/resume_website
go mod init github.com/EddisonSo/resume-website
go get gopkg.in/yaml.v3
```

Expected: `go.mod` and `go.sum` created, `go get` reports `added gopkg.in/yaml.v3`.

- [ ] **Step 2: Commit**

```bash
git add go.mod go.sum
git commit -m "Initialize Go module with yaml.v3 dependency"
```

---

### Task 2: YAML parsing and validation (TDD)

**Files:**
- Create: `generator/main.go` (types + `parseResume` + `validate` only; CLI comes in Task 4)
- Create: `generator/main_test.go`

- [ ] **Step 1: Write the failing tests**

Create `generator/main_test.go`:

```go
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
	if len(r.Education) != 1 || r.Education[0].Notes[0] != "Minor in QA" {
		t.Errorf("Education = %+v", r.Education)
	}
	if len(r.Experience) != 2 {
		t.Fatalf("len(Experience) = %d, want 2", len(r.Experience))
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/eddison/projects/resume_website && go test ./generator/
```

Expected: FAIL — compile error, `undefined: parseResume` (and undefined types).

- [ ] **Step 3: Write the parsing implementation**

Create `generator/main.go`:

```go
package main

import (
	"bytes"
	"errors"
	"fmt"
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
```

Note: no `main()` yet — `go test` does not require one, and the CLI is Task 4.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./generator/ -v
```

Expected: PASS, all 8 tests.

- [ ] **Step 5: Commit**

```bash
git add generator/main.go generator/main_test.go
git commit -m "Add strict YAML parsing and validation for resume schema"
```

---

### Task 3: HTML template and golden-file render test

**Files:**
- Create: `generator/resume.html.tmpl`
- Create: `generator/testdata/sample.yaml`
- Create: `generator/testdata/golden.html` (generated via `-update`, then reviewed)
- Modify: `generator/main.go` (add `render`)
- Modify: `generator/main_test.go` (add golden test)

- [ ] **Step 1: Create the test fixture**

Create `generator/testdata/sample.yaml` (same content as the `validYAML` test constant — it exercises every schema feature: header links/email, summary, education notes, experience with description+groups, experience with flat bullets, project):

```yaml
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
```

- [ ] **Step 2: Write the failing golden test**

Append to `generator/main_test.go`:

```go
var update = flag.Bool("update", false, "update golden files")

func TestRenderGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, err := parseResume(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := render(r, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	if *update {
		if err := os.WriteFile("testdata/golden.html", buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile("testdata/golden.html")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != string(want) {
		t.Errorf("rendered HTML does not match testdata/golden.html; run 'go test ./generator/ -update' and review the diff")
	}
}
```

Add to the existing import block in `generator/main_test.go`:

```go
import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./generator/ -run TestRenderGolden
```

Expected: FAIL — compile error, `undefined: render`.

- [ ] **Step 4: Create the template**

Create `generator/resume.html.tmpl`. The CSS block and overall structure are copied verbatim from the current `index.html` (lines 1–41): same `<style>`, same nested `.header` divs, same two-column `<table>`. Section labels render only on the first row of each section (`{{if eq $i 0}}`); later rows get an empty label cell, matching the current page.

```html
<!DOCTYPE html>
<html>
<head>
<style>
    table {
        font-family: arial, sans-serif;
        border-collapse: collapse;
        width: 70%;
    }

    td, th {
        text-align: left;
        padding: 8px;
        font-weight: normal;
        vertical-align: top;
    }

    .header {
        font-size: 16px;
        margin: 0;
        margin-top: 50px;
        margin-bottom: 10px;
        padding: 0;
        text-align: center;
    }

    * {
        font-family: Verdana, sans-serif;
    }
</style>
</head>
<body>

    <div class="header">
        <div class="header">
            <b>{{.Header.Name}}</b><br>
            {{- range .Header.Links}}
            <a href="{{.}}">{{.}}</a><br>
            {{- end}}
            {{- if .Header.Email}}
            <a href="mailto:{{.Header.Email}}">{{.Header.Email}}</a>
            {{- end}}
        </div>
    </div>
    <table>
{{- if .Summary}}
        <tr>
            <th style="width:20%">
                <b>Summary</b>
            </th>
            <th style="width:80%">
                {{.Summary}}
            </th>
        </tr>
{{- end}}
{{- range $i, $e := .Education}}
        <tr>
            <td>{{if eq $i 0}}<b>Education</b>{{end}}</td>
            <td>
                <i><b>{{$e.School}}</b></i><br>
                {{- if $e.Degree}}
                {{$e.Degree}}<br>
                {{- end}}
                {{- range $e.Notes}}
                {{.}}<br>
                {{- end}}
            </td>
        </tr>
{{- end}}
{{- range $i, $e := .Experience}}
        <tr>
            <td>{{if eq $i 0}}<b>Experience</b>{{end}}</td>
            <td>
                <i><b>{{$e.Org}}</b></i><br>
                {{- if $e.Subtitle}}
                {{$e.Subtitle}}<br>
                {{- end}}
                {{- if $e.Link}}
                <a href="{{$e.Link}}">{{$e.Link}}</a><br>
                {{- end}}
                {{- if $e.Title}}
                {{$e.Title}}<br>
                {{- end}}
                {{- if $e.Description}}
                <br>{{$e.Description}}<br>
                {{- end}}
                {{- range $e.Groups}}
                <br>{{.Label}}<br>
                <ul>
                    {{- range .Bullets}}
                    <li>{{.}}
                    {{- end}}
                </ul>
                {{- end}}
                {{- if $e.Bullets}}
                <ul>
                    {{- range $e.Bullets}}
                    <li>{{.}}
                    {{- end}}
                </ul>
                {{- end}}
            </td>
        </tr>
{{- end}}
{{- range $i, $p := .Projects}}
        <tr>
            <td>{{if eq $i 0}}<b>Projects</b>{{end}}</td>
            <td>
                <i><b>{{$p.Name}}</b></i><br>
                {{- if $p.Subtitle}}
                {{$p.Subtitle}}<br>
                {{- end}}
                {{- if $p.Link}}
                <a href="{{$p.Link}}">{{$p.Link}}</a><br>
                {{- end}}
                {{- if $p.Role}}
                {{$p.Role}}<br>
                {{- end}}
                {{- if $p.Bullets}}
                <ul>
                    {{- range $p.Bullets}}
                    <li>{{.}}
                    {{- end}}
                </ul>
                {{- end}}
            </td>
        </tr>
{{- end}}
    </table>

</body>
</html>
```

- [ ] **Step 5: Add the render function**

In `generator/main.go`, add to the import block:

```go
	"embed"
	"html/template"
```

and add below the type definitions:

```go
//go:embed resume.html.tmpl
var tmplFS embed.FS

func render(r *Resume, w io.Writer) error {
	t, err := template.ParseFS(tmplFS, "resume.html.tmpl")
	if err != nil {
		return err
	}
	return t.Execute(w, r)
}
```

(`io` is already imported from Task 2.)

- [ ] **Step 6: Generate the golden file and verify tests pass**

```bash
go test ./generator/ -update -run TestRenderGolden
go test ./generator/ -v
```

Expected: first command PASS (writes `testdata/golden.html`), second PASS for all tests.

- [ ] **Step 7: Review the golden file by eye**

Open `generator/testdata/golden.html` and check: header block with name/link/email, Summary row, Education row with degree+note, first Experience row with subtitle/link/title/description/"Backend:" group with 2 bullets, second Experience row with empty label cell and a flat bullet list, Projects row with role and bullet. All text from `sample.yaml` must appear; links must be `<a href>` anchors.

- [ ] **Step 8: Commit**

```bash
git add generator/resume.html.tmpl generator/main.go generator/main_test.go generator/testdata/
git commit -m "Add HTML template and render with golden-file test"
```

---

### Task 4: CLI entry, real resume.yaml, migration check

**Files:**
- Modify: `generator/main.go` (add `main()`)
- Create: `resume.yaml`

- [ ] **Step 1: Add the CLI entry point**

In `generator/main.go`, add to the import block:

```go
	"flag"
	"os"
```

and add at the bottom of the file:

```go
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
```

- [ ] **Step 2: Verify it builds and tests still pass**

```bash
go vet ./generator/ && go test ./generator/
```

Expected: no vet errors, all tests PASS.

- [ ] **Step 3: Transcribe the real content**

Create `resume.yaml` at the repo root. Content is transcribed **verbatim** from the current `index.html` — including the summary's existing "by to improve" wording (flag it to the user at the end, but do not silently edit content):

```yaml
header:
  name: Eddison So
  email: eddisonso@gmail.com
  links:
    - https://eddisonso2.com
    - https://eddisonso.com

summary: >
  Computer science student that specializes in enterprise systems and backend
  engineering. I am currently seeking opportunities to further my knowledge in
  these fields by to improve my knowledge in engineering practices that
  increase reliability and performance of software.

education:
  - school: Stevens Institute of Technology
    degree: Bachelors of Science in Computer Science, 2024 Fall
    notes:
      - Minor in Mathematics

experience:
  - org: Stevens Institute of Technology
    subtitle: Integrated Spatial Modeling and Remote Sensing Technologies Laboratory
    link: https://web.stevens.edu/ismart/
    title: Research Assistant
    description: >
      The goal of the project is to develop a scalable network of sensors that
      provide a novel method of tracking flood in urban regions. My role in the
      project is to add additional sensor functionality through the use of a
      static rain-gauge and develop the backend infrastructure to support data
      persistence and processing.
    groups:
      - label: "Sensor Development:"
        bullets:
          - Added additional sensor function by implementing precipitation sensor by modifying the microcontroller configuration to allow for an additional serial device.
          - Assisted in the implementation of LoRaWAN functionality by rewriting sensor firmware for periodic polling with LoRaWAN network.
      - label: "Backend Development:"
        bullets:
          - Designed backend infrastructure to allow for horizontal scaling across stateless services (API service and mapping webservice).
          - Utilized the Flask micro-framework for API access to the persistence of sensor data.
          - Designed database schema to persist various domain model objects such as sensors and sensordata.
          - Utilized PostgreSQL for the atomic operations during the persistence of data to support a high number of concurrent read/write operations from sensors and users.
          - Utilized nginx to loadbalance across multiple Flask instances.
          - Deployed using Docker to allow for inter-container networking and consistent deployment using Docker Compose.
          - Utilized Folium and Plotly for data visualization on the mapping webserver.
  - org: Stevens Institute of Technology
    title: Course Assistant for CS382 (Computer Architecture and Organization)
    bullets:
      - Held laboratory sessions to guide students through low-level programming and computer organization.
      - Held office hours to assist students with course material.
  - org: Stevens Institute of Technology
    title: Course Assistant for CS392 (Systems Programming)
    bullets:
      - Held laboratory sessions to guide students through systems programming using Linux.
      - Held office hours to assist students with course material.

projects:
  - name: Reels
    subtitle: HackNJIT Project
    link: https://github.com/team-reels/Reels
    role: Backend Developer
    bullets:
      - Collaborated with team members to decide technical requirements.
      - Developed backend and database schema to support CRUD operations and data persistence.
      - Utilized Google Cloud for image storage and hosting.
      - Utilized Docker for consistent deployment.
```

- [ ] **Step 4: Generate and run the migration check**

The check compares the **rendered text content** of old and new pages (tag-stripped, whitespace-normalized), per the spec: the goal is proving no content was lost, not byte-identical markup.

```bash
cd /home/eddison/projects/resume_website
go run ./generator -in resume.yaml -out /tmp/generated.html
python3 - <<'EOF'
from html.parser import HTMLParser
import difflib, re

class T(HTMLParser):
    def __init__(self):
        super().__init__()
        self.parts = []
    def handle_data(self, d):
        self.parts.append(d)

def text(path):
    p = T()
    p.feed(open(path).read())
    return re.sub(r'\s+', ' ', ' '.join(p.parts)).strip()

a, b = text('index.html'), text('/tmp/generated.html')
if a == b:
    print('MATCH: text content identical')
else:
    print('DIFFER:')
    for l in difflib.unified_diff(a.split(' '), b.split(' '), lineterm=''):
        print(l)
EOF
```

Expected: `MATCH: text content identical`. If it differs, the diff shows exactly which words were lost or changed — fix `resume.yaml` (transcription error) or the template (structural omission) and re-run. Do not proceed until MATCH.

- [ ] **Step 5: Visual spot-check**

```bash
go run ./generator -in resume.yaml -out /tmp/generated.html
```

Open `/tmp/generated.html` in a browser (or `python3 -m http.server` in `/tmp`) alongside the current `index.html` and confirm the layout matches: centered header, two-column table, bullet groups under Experience.

- [ ] **Step 6: Commit**

```bash
git add generator/main.go resume.yaml
git commit -m "Add generator CLI and resume.yaml content source"
```

---

### Task 5: Multi-stage Dockerfile, delete index.html

**Files:**
- Modify: `Dockerfile` (full replacement)
- Delete: `index.html`

- [ ] **Step 1: Replace the Dockerfile**

Overwrite `Dockerfile` with:

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY generator/ generator/
COPY resume.yaml ./
RUN go run ./generator -in resume.yaml -out /index.html

FROM nginx
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /index.html /etc/nginx/html/index.html
```

- [ ] **Step 2: Delete the hand-written HTML**

```bash
git rm index.html
```

- [ ] **Step 3: Verify the image builds and serves the page**

```bash
cd /home/eddison/projects/resume_website
docker build -t resume-test .
docker run -d --rm --name resume-test -p 8080:80 resume-test
sleep 1
curl -s http://localhost:8080/ | grep -c "Eddison So"
docker rm -f resume-test
```

Expected: build succeeds; `grep -c` prints a number ≥ 1 (the name appears in the served page). If docker is unavailable in the environment, state that explicitly and verify with `go run ./generator` output instead — do not claim the docker build was verified.

- [ ] **Step 4: Verify a broken YAML fails the build**

```bash
sed -i 's/^summary:/sumary:/' resume.yaml
docker build -t resume-test-fail . ; echo "exit: $?"
git checkout resume.yaml
```

Expected: build FAILS with an error mentioning `sumary` (strict decoding), exit code non-zero. Restore `resume.yaml` afterward (the `git checkout` line).

- [ ] **Step 5: Run the full test suite one last time**

```bash
go test ./generator/ -v && go vet ./generator/
```

Expected: all PASS, no vet errors.

- [ ] **Step 6: Commit**

```bash
git add Dockerfile
git commit -m "Build index.html from resume.yaml in multi-stage Docker build"
```

(The `git rm index.html` from Step 2 is included in this commit automatically since it's already staged.)

---

## Final verification (after all tasks)

- `go test ./generator/ -v` — all tests pass.
- `docker build .` — image builds, page served by nginx contains all resume content.
- `index.html` no longer in the repo; `resume.yaml` is the single source of truth.
- `.github/workflows/build-push.yml`, `buildx.sh`, `nginx/nginx.conf` untouched.
- Report to the user: the original summary contains the phrase "by to improve my knowledge", transcribed verbatim — they may want to fix the wording in `resume.yaml` (a one-line edit, now trivially deployable).
