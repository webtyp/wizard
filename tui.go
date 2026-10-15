package wizard

import "webtyp.com/lang"

const (
	textError     lang.Text = "Error"
	textCancelled lang.Text = "Wizard cancelled"
)

const (
	symbolConfirmed = "✓ "
	symbolMasked    = "••••••••"
)

// Handler Interface (compatible with app's TUI expectations)

func (w *Wizard) Name() string         { return w.stepMessage }
func (w *Wizard) Label() string        { return w.label }
func (w *Wizard) Value() string        { return w.currentValue }
func (w *Wizard) WaitingForUser() bool { return w.waitingForUser }

func (w *Wizard) Sensitive() bool {
	if w.currentStepIdx >= len(w.steps) {
		return false // wizard finished, no current step to be sensitive about
	}
	return w.steps[w.currentStepIdx].IsSensitive()
}

func (w *Wizard) Change(newValue string) {
	if w.currentStepIdx >= len(w.steps) {
		return
	}

	step := w.steps[w.currentStepIdx]
	continueFlow, err := step.OnInput(newValue, w.ctx)

	if err != nil {
		w.log(lang.Translate(textError).String() + ": " + err.Error())
		return
	}

	if !continueFlow {
		return
	}

	// Before advancing, log the user's input to preserve history — except the
	// raw value of a sensitive step, which must never reach the log (it
	// streams over SSE, see webtyp/devtui GET /logs).
	confirmed := newValue
	if step.IsSensitive() {
		confirmed = symbolMasked
	}
	w.log(symbolConfirmed + w.label + ": " + confirmed)

	nextSteps := step.Next(w.ctx)
	if len(nextSteps) > 0 {
		var oNext []orchestratorStep
		for _, s := range nextSteps {
			oNext = append(oNext, s)
		}
		// Insert right after the current step
		w.steps = append(w.steps[:w.currentStepIdx+1], append(oNext, w.steps[w.currentStepIdx+1:]...)...)
	}

	// Move to next step
	w.currentStepIdx++
	w.initCurrentStep()
}

// Loggable implementation
func (w *Wizard) SetLog(f func(message ...any)) {
	w.log = f
	w.hasRealLogger = true
}

// StreamingLoggable implementation
func (w *Wizard) AlwaysShowAllLogs() bool { return true }

// TabAware implementation
func (w *Wizard) OnTabActive() {
	// Let the wizard intercept TabAware so it can manually trigger
	// the OnShow hook of the current step when the tab becomes active.
	w.showCurrentStep()

	for _, source := range w.sources {
		if aware, ok := source.(interface{ OnTabActive() }); ok {
			aware.OnTabActive()
		}
	}
}

// Cancelable implementation
func (w *Wizard) Cancel() {
	w.waitingForUser = false
	w.log(lang.Translate(textCancelled).String())
}
