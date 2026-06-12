# YAML → HTML Resume Generator — Design

**Date:** 2026-06-12
**Status:** Approved

## Goal

Replace the hand-written `index.html` with a YAML source file (`resume.yaml`) that a
Go generator compiles into the same HTML document at Docker image build time. The
YAML becomes the only content source of truth; the rendered page keeps the current
visual design exactly.

## Decisions

- **Schema:** fixed resume schema (typed top-level keys: `header`, `summary`,
  `education`, `experience`, `projects`) — not a generic section system.
- **Tooling:** Go, `gopkg.in/yaml.v3`, `html/template`. No runtime dependencies.
- **Build:** Docker multi-stage. Stage 1 runs the generator; stage 2 is the existing
  nginx image. `index.html` is never committed.
- **Styling:** reproduce the current look (Verdana, two-column table, plain links)
  pixel-for-pixel. Restyling is out of scope.
- **Source layout:** single `resume.yaml` at the repo root.

## Repo layout

```
resume-website/
├── resume.yaml              # content source of truth
├── generator/
│   ├── main.go              # CLI: read YAML, validate, render template
│   ├── main_test.go         # golden-file + validation tests
│   └── resume.html.tmpl     # embedded via go:embed
├── go.mod
├── Dockerfile               # multi-stage: build HTML → nginx
├── nginx/nginx.conf         # unchanged
└── .github/workflows/build-push.yml  # unchanged
```

`index.html` is removed from the repo once the generated output is verified against it.

## YAML schema

```yaml
header:
  name: Eddison So            # required
  email: eddisonso@gmail.com
  links:
    - https://eddisonso2.com
    - https://eddisonso.com

summary: >
  Computer science student that specializes in ...

education:
  - school: Stevens Institute of Technology   # required per item
    degree: Bachelors of Science in Computer Science, 2024 Fall
    notes:
      - Minor in Mathematics

experience:
  - org: Stevens Institute of Technology      # required per item
    title: Research Assistant
    subtitle: Integrated Spatial Modeling and Remote Sensing Technologies Laboratory
    link: https://web.stevens.edu/ismart/
    description: >
      The goal of the project is ...
    groups:                   # optional labeled bullet groups
      - label: "Sensor Development:"
        bullets:
          - Added additional sensor function by ...
    bullets:                  # optional flat bullet list
      - Held laboratory sessions to ...

projects:
  - name: Reels               # required per item
    subtitle: HackNJIT Project
    link: https://github.com/team-reels/Reels
    role: Backend Developer
    bullets:
      - Collaborated with team members to ...
```

Field rules:

- Only the identifying field is required per item: `header.name`,
  `education[].school`, `experience[].org`, `projects[].name`.
- An experience item may carry `description`, flat `bullets`, labeled `groups`, or
  any mix; they render in that order (description, then groups, then bullets).
- All scalar text is plain text — auto-escaped by `html/template`. Links render as
  `<a href>` anchors whose text is the URL itself, matching the current page.

## Generator

- CLI: `go run ./generator -in resume.yaml -out index.html`, with those values as
  defaults so a bare `go run ./generator` works for local preview.
- Strict YAML decoding (`yaml.Decoder` with `KnownFields(true)`): unknown fields
  (e.g. a `bulets:` typo) fail the build with the YAML location in the error.
- Post-decode validation: missing required fields produce an error naming the
  section and item index (e.g. `experience[1]: missing org`). Empty input fails.
- Template (`resume.html.tmpl`, embedded with `go:embed`) reproduces the current
  page: same inline CSS, same `<table>` structure. The section label (`Summary`,
  `Education`, `Experience`, `Projects`) renders only in the first row of each
  section; continuation rows get an empty label cell. Sections with no items are
  omitted entirely.

## Build / deploy

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN go run ./generator -in resume.yaml -out /index.html

FROM nginx
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /index.html /etc/nginx/html/index.html
```

The existing GitHub Actions workflow and `buildx.sh` need no changes; they build the
Dockerfile as before. Invalid YAML fails the image build, so broken content cannot
deploy.

## Testing

- **Golden-file test:** a sample YAML fixture rendered and compared against a
  checked-in expected HTML file, so template changes show up as reviewable diffs.
- **Validation tests:** unknown field, missing required field, and empty file each
  return the expected error.
- **Migration check (one-time):** generate from a `resume.yaml` transcribed from the
  current content and diff against the existing `index.html`
  (whitespace-insensitive) to prove no content was lost before deleting it.

## Out of scope

- Visual redesign or CSS changes.
- Generic/pluggable section types.
- PDF output, multiple resume variants, i18n.
