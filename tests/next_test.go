package tests

import (
	"testing"

	"webtyp.com/context"
	"webtyp.com/fmt"
	"webtyp.com/wizard"
)

func TestWizardNextFnAddsSteps(t *testing.T) {
	// N1: step 1 has NextFn returning steps A, B; after step 1 is accepted, the labels shown are A, B, then the originally queued step 2
	stepA := &wizard.Step{LabelText: "Step A", OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil }}
	stepB := &wizard.Step{LabelText: "Step B", OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil }}
	step2 := &wizard.Step{LabelText: "Step 2", OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil }}

	step1 := &wizard.Step{
		LabelText: "Step 1",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
		NextFn: func(ctx *context.Context) []*wizard.Step {
			return []*wizard.Step{stepA, stepB}
		},
	}

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1, step2}}
	w := wizard.New(nil, mod)

	if w.Label() != "Step 1" {
		t.Errorf("Expected Step 1, got %s", w.Label())
	}
	w.Change("val1")

	if w.Label() != "Step A" {
		t.Errorf("Expected Step A, got %s", w.Label())
	}
	w.Change("valA")

	if w.Label() != "Step B" {
		t.Errorf("Expected Step B, got %s", w.Label())
	}
	w.Change("valB")

	if w.Label() != "Step 2" {
		t.Errorf("Expected Step 2, got %s", w.Label())
	}
}

func TestWizardNextFnErrorNoInsert(t *testing.T) {
	// N2: OnInputFn returns an error -> NextFn is not called and no step is inserted
	stepA := &wizard.Step{LabelText: "Step A"}
	step2 := &wizard.Step{LabelText: "Step 2"}

	nextCalled := false
	step1 := &wizard.Step{
		LabelText: "Step 1",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			if input == "err" {
				return false, fmt.Err("error")
			}
			return true, nil
		},
		NextFn: func(ctx *context.Context) []*wizard.Step {
			nextCalled = true
			return []*wizard.Step{stepA}
		},
	}

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1, step2}}
	w := wizard.New(nil, mod)

	w.Change("err")

	if nextCalled {
		t.Errorf("Expected NextFn not to be called on error")
	}

	if w.Label() != "Step 1" {
		t.Errorf("Expected to stay on Step 1, got %s", w.Label())
	}
}

func TestWizardNextFnFalseNoInsert(t *testing.T) {
	// N3: OnInputFn returns false, nil -> NextFn is not called
	stepA := &wizard.Step{LabelText: "Step A"}
	step2 := &wizard.Step{LabelText: "Step 2"}

	nextCalled := false
	step1 := &wizard.Step{
		LabelText: "Step 1",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			return false, nil
		},
		NextFn: func(ctx *context.Context) []*wizard.Step {
			nextCalled = true
			return []*wizard.Step{stepA}
		},
	}

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1, step2}}
	w := wizard.New(nil, mod)

	w.Change("val")

	if nextCalled {
		t.Errorf("Expected NextFn not to be called on false return")
	}

	if w.Label() != "Step 1" {
		t.Errorf("Expected to stay on Step 1, got %s", w.Label())
	}
}

func TestWizardNextFnNil(t *testing.T) {
	// N4: NextFn returns nil -> flow continues to the next queued step; onComplete runs once at the end
	step2 := &wizard.Step{LabelText: "Step 2", OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil }}

	step1 := &wizard.Step{
		LabelText: "Step 1",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
		NextFn: func(ctx *context.Context) []*wizard.Step {
			return nil
		},
	}

	completedCount := 0
	onComplete := func(ctx *context.Context) { completedCount++ }

	mod := &MockModule{name: "Test", steps: []*wizard.Step{step1, step2}}
	w := wizard.New(onComplete, mod)

	w.Change("val1")

	if w.Label() != "Step 2" {
		t.Errorf("Expected Step 2, got %s", w.Label())
	}

	w.Change("val2")

	if completedCount != 1 {
		t.Errorf("Expected onComplete to be called once, got %d", completedCount)
	}
}
