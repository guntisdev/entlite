# Docs system

Plan for entlite documentation: markdown in the repo, html on entlite.dev.

```
pkg/entlite/** + internal/generator/** + examples/**
        │  make docs
        ▼
docs/**.md      committed, readable on github, drift checked in CI
        │  make site
        ▼
dist/**.html    gitignored, built in CI, deployed to cloudflare pages
```

## ~~1. Doc comments, do this first~~
- [x] ~~one line comment on every exported method in `pkg/entlite/field/field.go`~~
- [x] ~~fill the gaps in `query.go`, `index.go`, `filter.go`, `schema.go`~~
- [x] ~~they feed pkg.go.dev and the generated reference, everything below depends on them~~

## ~~2. Generator skeleton~~
- [x] ~~`internal/docs/` package, the steps are a pipeline in go, not targets in the makefile~~
- [x] ~~steps: reference, examples, readme injection, site~~
- [x] ~~one thin `package main` in `internal/docs/cmd`, flags `-out`, `-html`, `-check`~~
- [x] ~~under internal/ not cmd/, so it stays out of the published CLI surface~~
- [x] ~~`make docs` and `make site` are one line entry points, they hold no logic~~
- [x] ~~not `go generate ./...` at the root, it would walk into every example `ent/generate.go`, which needs sqlc, buf and network~~

## 3. Reference pages, generated
- [ ] ast walk over `pkg/entlite` for method tables per builder interface
- [ ] `docs/reference/fields.md` with a field type x option matrix built from the interface methods
- [ ] `queries.md`, `filters.md`, `indexes.md`, `contracts.md`, `entity.md`
- [ ] `type-mapping.md`: go type -> sqlite / postgres / mysql -> proto -> ts
- [ ] export a thin `SQLTypeFor(dialect, fieldType)` in `internal/generator/sqlc` so the table is called, not copied
- [ ] `cli.md` from the `cmd/entlite` flags

## 4. Readme injection
- [ ] readmes stay hand written, only the region between markers is rewritten
- [ ] same trick the repo already uses with `<!-- teaches:start -->`
- [ ] root readme examples table generated from each example teaches block
- [ ] runs before the example pages, so they read the final readme not a stale table
- [ ] never regenerate a whole readme, the prose and the teaches bullets are yours

## 5. Example pages, generated
- [ ] one `docs/examples/NN-name.md` per example
- [ ] pull the readme body, the teaches block is already machine readable
- [ ] embed `ent/schema/*.go` as input, `schema.proto`, `schema.sql`, `queries.sql` as output
- [ ] read the files at build time, never copy `examples/` into `docs/`

## 6. Drift check
- [ ] `internal/docs/docs_test.go` runs the pipeline with `-check` and diffs against the tree
- [ ] picked up by `make test`, so `make all` and CI fail on stale docs
- [ ] fix the readme, it advertises `make teaches` and there is no such target

## 7. Guide pages, hand written
- [ ] `docs/guide/01-getting-started.md` from the readme get started section
- [ ] `02-project-layout.md`, what each folder in `ent/` is
- [ ] `03-pipeline.md`, dsl -> contract -> sqlc / buf -> gen
- [ ] `04-writing-a-schema.md`, snippets pulled from examples/01
- [ ] `docs/README.md` as the index

## 8. Html layer
- [ ] goldmark + html/template in `internal/docs`, `make site` writes `dist/`
- [ ] link rewriting: `*.md` links become clean site urls, repo source links become github blob urls
- [ ] extensionless urls like `/reference/fields/`, pick them once and never change them
- [ ] landing page at `/` with the pitch and a schema in / generated out sample
- [ ] client side search from a json index, small vanilla js, no build step
- [ ] version from `git describe --tags` stamped into the header

## 9. Seo and analytics
- [ ] one h1 per page, meta description from the first paragraph
- [ ] canonical link and opengraph tags
- [ ] generated `sitemap.xml` and `robots.txt`
- [ ] cloudflare web analytics script tag, no cookies so no consent banner

## 10. Deploy
- [ ] cloudflare pages project, point entlite.dev at it
- [ ] github action runs `make site` then `cloudflare/wrangler-action`
- [ ] deploy on push to main, same go version as the rest of CI
- [ ] add `dist/` to `.gitignore`

## 11. Versions, only when v2 comes
- [ ] entlite.dev serves latest at the root, no version in the path
- [ ] when v2 ships, rebuild v1 from its tag with a git worktree and publish it under `/v1/`
- [ ] archived pages get a banner, `noindex` and a canonical to the live page
- [ ] `CHANGELOG.md` at the root, rendered at `/changelog/`, covers everything below major
- [ ] pkg.go.dev already has per tag api docs, so no version switcher is needed
