# `go test` — Command Reference

## Selecting tests with `-run` and `-skip`

`-run` and `-skip` take an **unanchored** regular expression, split on unbracketed `/` into one pattern per level (test, subtest, sub-subtest). `-run TestUser` also runs `TestUserDelete` and `TestUserList`; anchor with `^…$` for an exact match. Subtest names have spaces replaced by `_`, so `t.Run("empty input", …)` is selected with `/empty_input`.

```bash
go test -run '^TestParse$' ./...              # exactly TestParse (unanchored TestParse also runs TestParseURL)
go test -run TestParse ./...                  # every test whose name contains TestParse
go test -run '^TestParse$/^empty_input$' ./... # one subtest; spaces in the t.Run name become _
go test -run 'Test(Add|Sub)$' ./...           # several tests (regexp alternation)
go test -run '/(unit|integration)' ./...      # any top-level test, only subtests matching the second level
go test -skip '^TestSlow$' ./...              # everything except TestSlow (Go 1.20+)
go test -skip '/large' ./...                  # every test, minus subtests matching large
go test -run TestParse -skip '/legacy' ./...  # combine: select, then exclude
go test -list '.*' ./pkg                      # list test names without running them
```

A parent always runs when a subtest pattern could match under it, so `-run X/Y` also runs `X`'s own setup code.

## Flakiness and independence

```bash
go test -count=1 ./...                        # bypass the test cache (cached results hide flakes)
go test -count=50 -run '^TestFlaky$' ./pkg    # repeat one test to surface an intermittent failure
go test -race -count=20 -run '^TestFlaky$' ./pkg # repeat under the race detector
go test -shuffle=on ./...                     # randomize test order; prints the seed
go test -shuffle=1712345678 ./...             # replay a failing order with the printed seed
go test -failfast ./...                       # stop at the first failure
go test -timeout 30s ./...                    # fail (with goroutine dump) instead of hanging for 10m default
go test -p 1 ./...                            # run packages serially (shared external resource)
go test -parallel 4 ./...                     # cap t.Parallel() concurrency within a package
```

`-shuffle` exposes order dependence: a test that passes in file order but fails shuffled is reading state another test left behind.

## Output

```bash
go test -v ./...                              # per-test PASS/FAIL lines, including subtests
go test -json ./... | tee test.json           # machine-readable events for CI tooling
go test -v -run '^TestX$' ./pkg 2>&1 | grep -E '^(---|    ---)' # just the result lines
```

## Build variants

```bash
go test -tags=integration ./...               # include //go:build integration files
go test -short ./...                          # set testing.Short() (skips; does not exclude from compilation)
go test -race ./...                           # data race detector
go test -cover ./...                          # coverage summary (→ coverage.md for profiles and modes)
go test -bench=. -benchmem -run='^$' ./...    # benchmarks only (→ benchmarks.md)
go test -fuzz='^FuzzParse$' -fuzztime=30s ./pkg # fuzz one target (only one package per -fuzz run)
go test -artifacts -outputdir=out ./...       # keep t.ArtifactDir() files (Go 1.26+)
go test -c -o pkg.test ./pkg                  # compile the test binary without running it
go vet ./...                                  # go test already runs a vet subset; this runs the full set
```

## Fuzz corpus

```bash
go test -run='^FuzzParse$/<hash>' ./pkg       # replay one failing input from testdata/fuzz/FuzzParse/
go clean -fuzzcache                           # drop the generated corpus cache (not testdata/)
```

A fuzz failure writes the input to `testdata/fuzz/<FuzzName>/`; commit it so the input runs as a regular seed on every `go test`.
