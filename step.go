package wizard

import (
	"webtyp.com/context"
	"webtyp.com/lang"
)

// Step represents a single interaction or execution unit in the wizard.
// It is designed to be used as a literal for easy instantiation.
type Step struct {
	LabelText lang.Text
	DefaultFn func(ctx *context.Context) string
	OnInputFn func(input string, ctx *context.Context) (continueFlow bool, err error)
	OnShowFn  func(log func(message ...any))
	// Sensitive marks this step's collected value as a secret: the wizard
	// never logs the raw input for this step, and any consumer that checks
	// Wizard.Sensitive() (e.g. webtyp/devtui, via tui.Sensitive) masks it
	// on screen while it is being typed.
	Sensitive bool
	// NextFn, when set, runs after this step's input is accepted (OnInputFn
	// returned true with no error). The steps it returns are inserted right
	// after this step, before any step that was already queued.
	NextFn func(ctx *context.Context) []*Step
}

// Label returns the prompt text for the UI.
func (s *Step) Label() string {
	return lang.Translate(s.LabelText).String()
}

// DefaultValue returns a suggestion based on the current context.
func (s *Step) DefaultValue(ctx *context.Context) string {
	if s.DefaultFn == nil {
		return ""
	}
	return s.DefaultFn(ctx)
}

// OnInput executes the step logic using a mutable context.
func (s *Step) OnInput(input string, ctx *context.Context) (bool, error) {
	if s.OnInputFn == nil {
		return true, nil
	}
	return s.OnInputFn(input, ctx)
}

// OnShow executes optional display logic for the step when it becomes active.
func (s *Step) OnShow(log func(message ...any)) {
	if s.OnShowFn != nil {
		s.OnShowFn(log)
	}
}

// IsSensitive reports Sensitive. A method, not the field itself, because
// orchestratorStep needs it and a Go type cannot have a field and a method
// share one name.
func (s *Step) IsSensitive() bool { return s.Sensitive }

// Next returns the steps produced by NextFn, or nil if none.
func (s *Step) Next(ctx *context.Context) []*Step {
	if s.NextFn == nil {
		return nil
	}
	return s.NextFn(ctx)
}
