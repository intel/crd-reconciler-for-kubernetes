package states

import (
	"errors"
)

// ErrInvalidState is the error returned when a given state does not exist
// in the state space for any job type.
var ErrInvalidState = errors.New("invalid state given")

// The State type's inhabitants comprise a job's state space.
type State string

// FSM is a type which can represent a finite state machine.
type FSM struct {
	adjm     adjMatrix
	strToIdx map[State]int
}

type adjMatrix [][]bool

// NewFSM returns an empty state machine.
func NewFSM(sts ...State) *FSM {
	count := len(sts)
	outer := make([][]bool, count)
	toIdx := make(map[State]int)
	for i, st := range sts {
		outer[i] = make([]bool, count)
		toIdx[st] = i
	}
	return &FSM{
		adjm:     outer,
		strToIdx: toIdx,
	}
}

// SetAdj creates an adjacency in the FSM.
func (f *FSM) SetAdj(from, to State) error {
	fromIdx, toIdx, err := f.getIndices(from, to)
	if err != nil {
		return ErrInvalidState
	}
	f.adjm[fromIdx][toIdx] = true
	return nil
}

// ValidTransition validates that the transition for `from` to `to` is
// valid.
func (f *FSM) ValidTransition(from, to State) bool {
	fromIdx, toIdx, err := f.getIndices(from, to)
	if err != nil {
		return false
	}
	return f.adjm[fromIdx][toIdx]
}

// PathExists validates that there exists a path from `from` -> `to`.
func (f *FSM) PathExists(from, to State) bool {
	fromIdx, toIdx, err := f.getIndices(from, to)
	if err != nil {
		return false
	}
	return f.pathExists(fromIdx, toIdx)
}

func (f *FSM) pathExists(from, to int) bool {
	if f.adjm[from][to] {
		return true
	}
	for idx, adj := range f.adjm[from] {
		if adj {
			if f.pathExists(idx, to) {
				return true
			}
		}
	}
	return false
}

func (f *FSM) getIndices(from, to State) (fromIdx int, toIdx int, err error) {
	fromIdx, ok := f.strToIdx[from]
	if !ok {
		return -1, -1, ErrInvalidState
	}
	toIdx, ok = f.strToIdx[to]
	if !ok {
		return -1, -1, ErrInvalidState
	}
	return fromIdx, toIdx, nil
}
