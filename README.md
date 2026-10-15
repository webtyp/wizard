# webtyp/wizard
<img src="docs/img/badges.svg">

A modular, pluggable wizard orchestrator for TinyGo applications.

## LLM Context / Core Logic
The library orchestrates a sequence of `Step` struct helpers. It manages state via a **mutable** `webtyp/context`.

### Steps
```go
type Step struct {
    LabelText  lang.Text
    DefaultFn  func(ctx *context.Context) string
    OnInputFn  func(input string, ctx *context.Context) (continueFlow bool, err error)
    // NextFn dynamically inserts returned steps immediately following the current step.
    NextFn     func(ctx *context.Context) []*Step
}
```

### Flow Control
- **New(onComplete func(ctx *context.Context), sources ...StepSource)**: Initializes orchestration. `StepSource` is an interface with `GetSteps() []*Step`.
- **Change(input)**: Executes `OnInput` for the current step.
- If `error != nil`, the wizard **stays on the current step**.
- If `error == nil` and `continueFlow == true`, the wizard **advances to the next step**.

## Features
- **Zero dynamic allocations** (fixed capacity context).
- **Mutable State**: Steps modify the context directly via `ctx.Set()`.
- **Easy Literals**: Define steps as simple structs without boilerplate.
- **TUI-ready**: Implements `Handler` and `Loggable` patterns.

## Usage

```go
// 1. Define a StepSource
type MyModule struct {
    Steps []*wizard.Step
}
func (m MyModule) Name() string { return "My Module" }
func (m MyModule) GetSteps() []*wizard.Step {
    return m.Steps
}

// 2. Create steps
steps := []*wizard.Step{
    {
        LabelText: "Project Name",
        DefaultFn: func(ctx *context.Context) string { return "my-app" },
        OnInputFn: func(in string, ctx *context.Context) (bool, error) {
            if in == "" { return false, nil }
            err := ctx.Set("name", in) // Mutate context in-place
            return true, err
        },
    },
}

// 3. Run wizard
wiz := wizard.New(func(ctx *context.Context) { println("Done!") }, &MyModule{Steps: steps})
wiz.Change("new-project")
```

## Translations

Use `wizard.Dictionary()` to get the embedded `lang.json` translations, suitable for loading with `lang.Load()`. Step `LabelText` uses `lang.Text` and will be translated automatically.
