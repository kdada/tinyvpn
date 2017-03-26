package state

import "testing"

func TestBaseMachine(t *testing.T) {
	stateA := NewBaseHandler("A")
	stateB := NewBaseHandler("B")
	stateC := NewBaseHandler("C")
	machine := NewBaseMachine()
	machine.AddStateHandler(stateA, stateB, stateC)
	machine.AddTransition("AB", "A", "B")
	machine.AddTransition("AC", "A", "C")
	machine.AddTransition("BB", "B", "B")
	machine.AddTransition("CA", "C", "A")
	check(t, machine, "A", machine.Start("A", nil))
	check(t, machine, "C", machine.Trigger("AC", nil))
	check(t, machine, "A", machine.Trigger("CA", nil))
	check(t, machine, "B", machine.Trigger("AB", nil))
	check(t, machine, "B", machine.Trigger("BB", nil))
}

func check(t *testing.T, m *BaseMachine, expect State, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if m.State != expect {
		t.Fatalf("state should %s, but got %s", expect, m.State)
	}
}
