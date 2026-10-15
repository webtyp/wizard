---
PLAN: "feat!: typed step sources, steps added at run time, translated labels"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 18053327749468295567
---

> Read AGENTS.md at the repo root before any change.

# Plan — wizard: typed sources, run-time steps, English/Spanish

Part of the "Despliegue guiado" wave (`https://github.com/webtyp/docs/blob/main/DEPLOY_WIZARD_MASTER_PLAN.md`,
decisions D22, D36). The deploy tab of `webtyp/app` builds one wizard step per missing item; after a
value is accepted it re-checks and must **append** the steps that are still missing. Today the
wizard cannot add steps after `New` (`webtyp/deploy` worked around it with 5 fixed "proxy" steps),
accepts `items ...any` with a runtime type assertion, and shows labels untranslated.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

## Design gate

1. **Prior art.** `charmbracelet/huh` (forms whose groups can be hidden/shown at run time with
   `WithHideFunc`), `AlecAivazis/survey` (questions as a typed slice), Inquirer.js (`when` per
   question). All type their questions and allow the flow to depend on earlier answers. Here a step
   returns the steps that follow it (`NextFn`), because the next steps are only known after a
   network check, not from a boolean on a pre-declared step.
2. **Names.** `StepSource` — "a thing that gives steps". `NextFn` — matches the existing
   `DefaultFn`, `OnInputFn`, `OnShowFn`.
3. **Ledger.** Concepts +1 (`NextFn`) / −0. Ways to provide steps: `items ...any` (anything with
   `GetSteps`, silently ignored otherwise) → `sources ...StepSource` (compile-time checked); −1 silent
   path.
4. **Where.** In `wizard`: it owns step order.
5. **Deletes.** The `any` parameter and its type assertion (`wizard.go:43-49`).

## Stage 1 — typed sources

`wizard.go`:

```go
// StepSource provides the steps a wizard runs, in order.
type StepSource interface {
	GetSteps() []*Step
}

func New(onComplete func(ctx *context.Context), sources ...StepSource) *Wizard
```

The field `items []any` becomes `sources []StepSource`; `OnTabActive` keeps forwarding to every
source that also implements `interface{ OnTabActive() }`.

Known callers that already satisfy `StepSource` and need no change: `devflow.GoNew`,
`devflow.CodeJob`, `app.SecretSteps`.

## Stage 2 — steps added at run time

`step.go`: add to `Step`

```go
	// NextFn, when set, runs after this step's input is accepted (OnInputFn
	// returned true with no error). The steps it returns are inserted right
	// after this step, before any step that was already queued.
	NextFn func(ctx *context.Context) []*Step
```

`tui.go` `Change`: after a successful, continuing input and before `currentStepIdx++`, if the
current step is a `*Step` with `NextFn != nil`, insert its result at `currentStepIdx+1`. The
orchestrator only knows `orchestratorStep`; add `Next(ctx *context.Context) []*Step` to that
interface and implement it on `*Step` (nil `NextFn` → nil).

## Stage 3 — translated text

- `step.go`: `LabelText lang.Text` (import `webtyp.com/lang`). `(*Step).Label()` returns
  `lang.Translate(s.LabelText).String()`.
- Every fixed message the wizard prints becomes a typed constant and is translated when printed:

  ```go
  const (
  	textError     lang.Text = "Error"
  	textCancelled lang.Text = "Wizard cancelled"
  )
  ```

  `Change` logs `lang.Translate(textError).String() + ": " + err.Error()`; `Cancel` logs
  `lang.Translate(textCancelled).String()`. The confirmation prefix `"✓ "` and the mask
  `"••••••••"` are symbols, not text: make them named constants (`symbolConfirmed`,
  `symbolMasked`), untranslated.
- New `lang.json` at the repo root with `"languages": ["es", "zh", "hi", "ar", "pt", "fr", "de", "ru"]`
  and one key per constant; the `es` column is mandatory (`"Error"` → `"Error"`,
  `"Wizard cancelled"` → `"Asistente cancelado"`); other columns may be empty strings.
- New `dictionary.go`:

  ```go
  //go:embed lang.json
  var dictionary []byte

  // Dictionary returns wizard's translations (its lang.json, embedded), for
  // lang.Load at the tool's startup.
  func Dictionary() []byte { return dictionary }
  ```

- Run `langc sync` (see `langc -h`) and keep `lang.json` in the format it writes.

Callers that pass a `string` variable (not a literal) to `LabelText` must convert:
`app/section-secrets.go:28` (`LabelText: key`) — that file is rewritten by the app plan of this
wave, so do **not** touch `webtyp/app` here.

## Stage 4 — tests in `tests/`

Move `wizard_test.go` and `sensitive_test.go` to `tests/` as package `tests` (external). Assertions
that read unexported fields (`w.currentStepIdx`, `wizard_test.go:106-119`) are rewritten against
exported behaviour (`Label()`, `WaitingForUser()`, the logged lines). Do not export anything for a
test.

New tests:

| Branch | Test |
|---|---|
| N1 | step 1 has `NextFn` returning steps A, B; after step 1 is accepted, the labels shown are A, B, then the originally queued step 2 |
| N2 | `OnInputFn` returns an error → `NextFn` is not called and no step is inserted |
| N3 | `OnInputFn` returns `false, nil` → `NextFn` is not called |
| N4 | `NextFn` returns nil → flow continues to the next queued step; `onComplete` runs once at the end |
| T1 | with `lang.OutLang("es")` and `lang.Load(wizard.Dictionary())`, `Cancel()` logs `Asistente cancelado` |
| T2 | a `LabelText` with no dictionary entry is shown unchanged |
| S1 | a `Sensitive` step inserted by `NextFn` is still masked in the log (`symbolMasked`) |

## Stage 5 — docs

`README.md`: document `StepSource`, `NextFn`, `LabelText lang.Text`, `Dictionary()`.

| Stage | Files |
|---|---|
| 1 | `wizard.go` |
| 2 | `step.go`, `tui.go`, `wizard.go` |
| 3 | `step.go`, `tui.go`, `lang.json` (new), `dictionary.go` (new), `go.mod`, `go.sum` |
| 4 | `tests/wizard_test.go`, `tests/sensitive_test.go`, `tests/next_test.go` (new), `tests/lang_test.go` (new); delete `wizard_test.go`, `sensitive_test.go` |
| 5 | `README.md` |

## Acceptance

- `gotest` green.
- `grep -rn "items \.\.\.any\|\.(interface{ GetSteps" --include=*.go .` → empty.
- `grep -rn "\"Wizard cancelled\"\|\"Error: \"" --include=*.go . | grep -v "lang.Text"` → empty.
- `test ! -e wizard_test.go && test ! -e sensitive_test.go`.
- `git diff --stat main` lists every file of the stages table and deletes only the two moved test files.
