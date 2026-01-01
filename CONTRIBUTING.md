# Contributing to snitch

thanks for your interest in contributing to snitch! this document outlines how to get started.

## development setup

### prerequisites

- go 1.21 or later
- make (optional, but recommended)
- linux or macos

### building from source

```bash
git clone https://github.com/karol-broda/snitch.git
cd snitch

# build
make build
# or
go build -o snitch .

# run
./snitch
```

### running tests

```bash
make test
# or
go test ./...
```

### linting

```bash
make lint
# requires golangci-lint
```

## making changes

### branch naming

use descriptive branch names following the [conventional branch naming](https://conventional-branch.github.io/) pattern:

- `fix/description` for bug fixes
- `feat/description` for new features
- `docs/description` for documentation changes
- `refactor/description` for refactoring
- `chore/description` for maintenance tasks

### code style

- follow existing code patterns and conventions
- avoid deep nesting; refactor for readability
- use explicit checks rather than implicit boolean coercion
- keep functions focused on a single responsibility
- write meaningful variable names
- add comments only when they clarify non-obvious behavior

### commits

this project follows [conventional commits](https://www.conventionalcommits.org/). format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

types: `fix`, `feat`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

examples:
- `fix: resolve crash on empty input`
- `feat(tui): add vim-style navigation`
- `docs: update installation instructions`
- `fix!: change default config path` (breaking change)

## ai-assisted contributions

ai tools (copilot, chatgpt, claude, cursor, etc.) are welcome, but i require transparency.

### disclosure requirements

**you must disclose ai involvement** in your pull request. add one of the following to your PR description:

- `ai: none` — no ai assistance used
- `ai: assisted` — ai helped with portions (specify what)
- `ai: generated` — significant portions were ai-generated

for commits with substantial ai involvement, use a git trailer:

```
feat: add new filtering option

Co-authored-by: AI Assistant <ai@example.com>
```

### your responsibilities

- **you own the code** — you are accountable for all submitted code, regardless of how it was produced
- **you must understand it** — don't submit code you can't explain or debug
- **you must test it** — verify the code works as intended before submitting
- **you must review it** — check for correctness, security issues, and code style compliance

### what i check

ai-generated code often has patterns i look for:
- overly verbose or generic variable names
- unnecessary abstractions or over-engineering
- hallucinated apis or non-existent functions
- inconsistent style with the rest of the codebase

i may ask clarifying questions or request changes if code appears to be unreviewed ai output.

### why i require disclosure

- maintains trust and transparency in the project
- helps reviewers understand context and potential issues
- ensures contributors remain engaged with their submissions
- respects the collaborative nature of open source

## submitting changes

1. fork the repository
2. create a feature branch from `master`
3. make your changes
4. run tests: `make test`
5. run linter: `make lint`
6. push to your fork
7. open a pull request

### pull request guidelines

- fill out the PR template
- link any related issues
- keep PRs focused on a single change
- respond to review feedback promptly

## reporting bugs

use the [bug report template](https://github.com/karol-broda/snitch/issues/new?template=bug_report.yml) and include:

- snitch version (`snitch version`)
- operating system and version
- steps to reproduce
- expected vs actual behavior

## requesting features

use the [feature request template](https://github.com/karol-broda/snitch/issues/new?template=feature_request.yml) and describe:

- the problem you're trying to solve
- your proposed solution
- any alternatives you've considered

## getting help

- open a [discussion](https://github.com/karol-broda/snitch/discussions) for questions
- check existing [issues](https://github.com/karol-broda/snitch/issues) before opening new ones

## license

by contributing, you agree that your contributions will be licensed under the project's MIT license.

