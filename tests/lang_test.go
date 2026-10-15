package tests

import (
	"strings"
	"testing"

	stdFmt "fmt"

	"webtyp.com/context"
	"webtyp.com/lang"
	"webtyp.com/wizard"
)

func TestLangCancelTranslated(t *testing.T) {
	// T1: with lang.OutLang("es") and lang.Load(wizard.Dictionary()), Cancel() logs Asistente cancelado
	lang.Load(wizard.Dictionary())
	lang.OutLang("es")

	mod := &MockModule{name: "Test", steps: []*wizard.Step{}}
	w := wizard.New(nil, mod)

	var captured []string
	w.SetLog(func(msgs ...any) {
		captured = append(captured, stdFmt.Sprint(msgs...))
	})

	w.Cancel()

	found := false
	for _, msg := range captured {
		if strings.Contains(msg, "Asistente cancelado") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected translated text 'Asistente cancelado' in log, got %v", captured)
	}

	// Reset for other tests
	lang.OutLang("en")
}

func TestLangMissingDictionaryEntry(t *testing.T) {
	// T2: a LabelText with no dictionary entry is shown unchanged
	lang.Load(wizard.Dictionary())
	lang.OutLang("es")

	step1 := &wizard.Step{
		LabelText: "Unknown Label",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1}}
	w := wizard.New(nil, mod)

	if w.Label() != "Unknown Label" {
		t.Fatalf("expected 'Unknown Label', got %s", w.Label())
	}

	// Reset
	lang.OutLang("en")
}

func TestSensitiveStepInsertedByNextFn(t *testing.T) {
	// S1: a Sensitive step inserted by NextFn is still masked in the log (symbolMasked)
	stepSens := &wizard.Step{
		LabelText: "Sensitive Step",
		Sensitive: true,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}

	step1 := &wizard.Step{
		LabelText: "Step 1",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
		NextFn: func(ctx *context.Context) []*wizard.Step {
			return []*wizard.Step{stepSens}
		},
	}

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1}}
	w := wizard.New(nil, mod)

	var captured []string
	w.SetLog(func(msgs ...any) {
		captured = append(captured, stdFmt.Sprint(msgs...))
	})

	w.Change("val1") // Accept step 1
	w.Change("super-secret") // Accept sensitive step

	for _, msg := range captured {
		if strings.Contains(msg, "super-secret") {
			t.Fatalf("log must NOT contain raw sensitive value, got %q", msg)
		}
	}

	found := false
	for _, msg := range captured {
		if strings.Contains(msg, "••••••••") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected masked value '••••••••' in log, got %v", captured)
	}
}
