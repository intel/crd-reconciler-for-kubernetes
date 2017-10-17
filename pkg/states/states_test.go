package states_test

import (
	"testing"

	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
)

func getFSM(t *testing.T) *states.FSM {
	fsm := states.NewFSM("a", "b", "c", "d")
	err := fsm.SetAdj("a", "b")
	if err != nil {
		t.Fatal("getFSM: SetAdj should return nil but it returned an error")
	}
	err = fsm.SetAdj("b", "c")
	if err != nil {
		t.Fatal("getFSM: SetAdj should return nil but it returned an error")
	}
	err = fsm.SetAdj("a", "d")
	if err != nil {
		t.Fatal("getFSM: SetAdj should return nil but it returned an error")
	}
	return fsm
}

func TestPathExists(t *testing.T) {
	fsm := getFSM(t)
	if !fsm.PathExists("a", "c") {
		t.Fatal("PathExists: a path exists between 'a' and 'c', but " +
			"PathExists returned false")
	}

	if fsm.PathExists("b", "d") {
		t.Fatal("PathExists: a path does not exist between 'b' and 'd', but " +
			"PathExists returned true")
	}
}

func TestValidTransition(t *testing.T) {
	fsm := getFSM(t)
	if !fsm.ValidTransition("a", "b") {
		t.Fatal("ValidTransition: 'a' -> 'b' is a valid transition, but " +
			"ValidTransition returned false")
	}
	if fsm.ValidTransition("a", "c") {
		t.Fatal("ValidTransition: 'a' -> 'c' is not a valid transition, but " +
			"ValidTransition returned true")
	}
}
