package rvb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {

	type testSequenceStep struct {
		Push         []int
		ReadN        int
		ExpectedNew  []int
		ExpectedOld  []int
		ExpectedSize int
	}

	type testCase struct {
		Name string

		BufferSize   int
		testSequence []testSequenceStep
	}

	for _, tc := range []testCase{
		{
			Name:       "simple overflow",
			BufferSize: 4,
			testSequence: []testSequenceStep{
				{
					Push:         []int{},
					ReadN:        4,
					ExpectedNew:  []int{},
					ExpectedOld:  []int{},
					ExpectedSize: 0,
				},
				{
					Push:         []int{1, 2},
					ReadN:        4,
					ExpectedNew:  []int{2, 1},
					ExpectedOld:  []int{1, 2},
					ExpectedSize: 2,
				},
				{
					Push:         []int{3, 4},
					ReadN:        4,
					ExpectedNew:  []int{4, 3, 2, 1},
					ExpectedOld:  []int{1, 2, 3, 4},
					ExpectedSize: 4,
				},
				{
					Push:         []int{5, 6},
					ReadN:        4,
					ExpectedNew:  []int{6, 5, 4, 3},
					ExpectedOld:  []int{3, 4, 5, 6},
					ExpectedSize: 4,
				},
			},
		},
		{
			Name:       "simple overflow with smaller reads",
			BufferSize: 4,
			testSequence: []testSequenceStep{
				{
					Push:         []int{},
					ReadN:        2,
					ExpectedNew:  []int{},
					ExpectedOld:  []int{},
					ExpectedSize: 0,
				},
				{
					Push:         []int{1, 2},
					ReadN:        2,
					ExpectedNew:  []int{2, 1},
					ExpectedOld:  []int{1, 2},
					ExpectedSize: 2,
				},
				{
					Push:         []int{3, 4},
					ReadN:        2,
					ExpectedNew:  []int{4, 3},
					ExpectedOld:  []int{1, 2},
					ExpectedSize: 4,
				},
				{
					Push:         []int{5, 6},
					ReadN:        2,
					ExpectedNew:  []int{6, 5},
					ExpectedOld:  []int{3, 4},
					ExpectedSize: 4,
				},
			},
		},
		{
			Name:       "simple overflow with greater reads",
			BufferSize: 4,
			testSequence: []testSequenceStep{
				{
					Push:         []int{},
					ReadN:        8,
					ExpectedNew:  []int{},
					ExpectedOld:  []int{},
					ExpectedSize: 0,
				},
				{
					Push:         []int{1, 2},
					ReadN:        8,
					ExpectedNew:  []int{2, 1},
					ExpectedOld:  []int{1, 2},
					ExpectedSize: 2,
				},
				{
					Push:         []int{3, 4},
					ReadN:        8,
					ExpectedNew:  []int{4, 3, 2, 1},
					ExpectedOld:  []int{1, 2, 3, 4},
					ExpectedSize: 4,
				},
				{
					Push:         []int{5, 6},
					ReadN:        8,
					ExpectedNew:  []int{6, 5, 4, 3},
					ExpectedOld:  []int{3, 4, 5, 6},
					ExpectedSize: 4,
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			rvb := NewBuffer[int](tc.BufferSize)

			for i, step := range tc.testSequence {
				t.Run(fmt.Sprintf("step %d", i), func(t *testing.T) {
					rvb.PushMany(step.Push)
					assert.Equal(t, step.ExpectedNew, rvb.ReadNew(step.ReadN))
					assert.Equal(t, step.ExpectedOld, rvb.ReadOld(step.ReadN))
					assert.Equal(t, step.ExpectedSize, rvb.GetCurrentSIze())
				})
			}
		})
	}
}

func TestCheckpoint(t *testing.T) {

	type stepAssertCheckpoint struct {
		ReadN           int
		ReadSkip        int
		Expected        []int
		ExpectedMissing Missing
		ExpectedSize    int
		ExpectedNew     int
	}

	type stepCheckpointSave struct{}

	type stepPush struct {
		Values []int
	}

	type testCase struct {
		Name string

		BufferSize   int
		testSequence []any
	}

	for _, tc := range []testCase{
		{
			Name:       "checkpoint entire buffer",
			BufferSize: 4,
			testSequence: []any{
				stepPush{[]int{1, 2}},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    2,
					ExpectedNew:     0,
				},
				stepPush{[]int{3}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    3,
					ExpectedNew:     1,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{3, 2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 3},
					ExpectedSize:    3,
					ExpectedNew:     0,
				},
				stepPush{[]int{4, 5, 6, 7, 8}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 3, Max: 3},
					ExpectedSize:    4,
					ExpectedNew:     5,
				},
				stepPush{[]int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 3, Max: 3},
					ExpectedSize:    4,
					ExpectedNew:     24,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{27, 26, 25, 24},
					ExpectedMissing: Missing{Reused: 0, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     0,
				},
				stepPush{[]int{28, 29}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{27, 26},
					ExpectedMissing: Missing{Reused: 2, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     2,
				},
				stepPush{[]int{30, 31}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 4, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     4,
				},
				stepPush{[]int{32, 33}},
				stepAssertCheckpoint{
					ReadN:           4,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 4, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     6,
				},
			},
		},
		{
			Name:       "checkpoint entire buffer with smaller reads and skip",
			BufferSize: 4,
			testSequence: []any{
				stepPush{[]int{1, 2}},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    2,
					ExpectedNew:     0,
				},
				stepPush{[]int{3}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    3,
					ExpectedNew:     1,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{3, 2},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    3,
					ExpectedNew:     0,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{1},
					ExpectedMissing: Missing{Reused: 0, Max: 1},
					ExpectedSize:    3,
					ExpectedNew:     0,
				},
				stepPush{[]int{4, 5, 6, 7, 8}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     5,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 1, Max: 1},
					ExpectedSize:    4,
					ExpectedNew:     5,
				},
				stepPush{[]int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     24,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{27, 26},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     0,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{25, 24},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     0,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        4,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 0, Max: 0},
					ExpectedSize:    4,
					ExpectedNew:     0,
				},
				stepPush{[]int{28, 29}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{27, 26},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     2,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     2,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        4,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 0, Max: 0},
					ExpectedSize:    4,
					ExpectedNew:     2,
				},
				stepPush{[]int{30, 31}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     4,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     4,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        4,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 0, Max: 0},
					ExpectedSize:    4,
					ExpectedNew:     4,
				},
				stepPush{[]int{32, 33}},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     6,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        2,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 2, Max: 2},
					ExpectedSize:    4,
					ExpectedNew:     6,
				},
				stepAssertCheckpoint{
					ReadN:           2,
					ReadSkip:        4,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 0, Max: 0},
					ExpectedSize:    4,
					ExpectedNew:     6,
				},
			},
		},
		{
			Name:       "checkpoint entire buffer with greater reads",
			BufferSize: 4,
			testSequence: []any{
				stepPush{[]int{1, 2}},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    2,
					ExpectedNew:     0,
				},
				stepPush{[]int{3}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 2},
					ExpectedSize:    3,
					ExpectedNew:     1,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{3, 2, 1},
					ExpectedMissing: Missing{Reused: 0, Max: 3},
					ExpectedSize:    3,
					ExpectedNew:     0,
				},
				stepPush{[]int{4, 5, 6, 7, 8}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 3, Max: 3},
					ExpectedSize:    4,
					ExpectedNew:     5,
				},
				stepPush{[]int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 3, Max: 3},
					ExpectedSize:    4,
					ExpectedNew:     24,
				},
				stepCheckpointSave{},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{27, 26, 25, 24},
					ExpectedMissing: Missing{Reused: 0, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     0,
				},
				stepPush{[]int{28, 29}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{27, 26},
					ExpectedMissing: Missing{Reused: 2, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     2,
				},
				stepPush{[]int{30, 31}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 4, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     4,
				},
				stepPush{[]int{32, 33}},
				stepAssertCheckpoint{
					ReadN:           8,
					ReadSkip:        0,
					Expected:        []int{},
					ExpectedMissing: Missing{Reused: 4, Max: 4},
					ExpectedSize:    4,
					ExpectedNew:     6,
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			rvb := NewBuffer[int](tc.BufferSize)

			var cp Checkpoint
			for i, step := range tc.testSequence {
				switch step := step.(type) {
				case stepAssertCheckpoint:
					t.Run(fmt.Sprintf("step %d", i), func(t *testing.T) {
						out, missing := rvb.ReadNewFromCheckpoint(cp, step.ReadSkip, step.ReadN)
						assert.Equal(t, step.Expected, out)
						assert.Equal(t, step.ExpectedMissing, missing)
						assert.Equal(t, step.ExpectedNew, rvb.NewItemsSince(cp))
						assert.Equal(t, step.ExpectedSize, rvb.GetCurrentSIze())
					})
				case stepPush:
					rvb.PushMany(step.Values)
				case stepCheckpointSave:
					cp = rvb.GetCheckpoint()
				default:
					t.Fatal(fmt.Errorf("unexpected step: %T", step))
				}
			}
		})
	}
}
