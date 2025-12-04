package bulk

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sliceLargeInput = []int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
	27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
	27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
}
var sliceTestCases = []struct {
	name        string
	input       []int
	testFunc    func(v int) bool
	expectTrue  []int
	expectFalse []int
	trueCapMin  int
	trueCapMax  int
	falseCapMin int
	falseCapMax int
}{
	{
		name:        "nil",
		input:       nil,
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  nil,
		expectFalse: nil,
		trueCapMin:  0,
		trueCapMax:  0,
		falseCapMin: 0,
		falseCapMax: 0,
	},
	{
		name:        "empty",
		input:       []int{},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  nil,
		expectFalse: nil,
		trueCapMin:  0,
		trueCapMax:  0,
		falseCapMin: 0,
		falseCapMax: 0,
	},
	{
		name:        "all_true",
		input:       []int{2, 4, 6},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6},
		expectFalse: nil,
		// should return original slice
		trueCapMin: 3,
		trueCapMax: 3,
		// empty slice, but may retain original capacity as view
		falseCapMin: 0,
		falseCapMax: 3,
	},
	{
		name:        "all_false",
		input:       []int{1, 3, 5},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  nil,
		expectFalse: []int{1, 3, 5},
		// empty slice, but may retain original capacity as view
		trueCapMin: 0,
		trueCapMax: 3,
		// should return original slice
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "all_true_large",
		input:       sliceLargeInput,
		testFunc:    func(v int) bool { return true },
		expectTrue:  sliceLargeInput,
		expectFalse: nil,
		// should return original slice
		trueCapMin: 101,
		trueCapMax: 101,
		// empty slice, but may retain original capacity as view
		falseCapMin: 0,
		falseCapMax: 101,
	},
	{
		name:        "all_false_large",
		input:       sliceLargeInput,
		testFunc:    func(v int) bool { return false },
		expectTrue:  nil,
		expectFalse: sliceLargeInput,
		// empty slice, but may retain original capacity as view
		trueCapMin: 0,
		trueCapMax: 101,
		// should return original slice
		falseCapMin: 101,
		falseCapMax: 101,
	},
	{
		name:        "single_true",
		input:       []int{2},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2},
		expectFalse: nil,
		// should return original slice
		trueCapMin: 1,
		trueCapMax: 1,
		// empty slice, but may retain original capacity as view
		falseCapMin: 0,
		falseCapMax: 1,
	},
	{
		name:        "single_false",
		input:       []int{1},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  nil,
		expectFalse: []int{1},
		// empty slice, but may retain original capacity as view
		trueCapMin: 0,
		trueCapMax: 1,
		// should return original slice
		falseCapMin: 1,
		falseCapMax: 1,
	},
	{
		name:        "one_true",
		input:       []int{1, 2, 3, 4, 6, 8},
		testFunc:    func(v int) bool { return v == 1 },
		expectTrue:  []int{1},
		expectFalse: []int{2, 3, 4, 6, 8},
		trueCapMin:  1,
		trueCapMax:  6,
		falseCapMin: 5,
		falseCapMax: 5,
	},
	{
		name:        "one_true_large_first",
		input:       sliceLargeInput,
		testFunc:    func(v int) bool { return v == 0 },
		expectTrue:  []int{0},
		expectFalse: sliceLargeInput[1:],
		trueCapMin:  1,
		trueCapMax:  101,
		falseCapMin: 100,
		falseCapMax: 100,
	},
	{
		name:       "two_true_large_middle",
		input:      sliceLargeInput,
		testFunc:   func(v int) bool { return v == 40 },
		expectTrue: []int{40, 40},
		expectFalse: []int{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
			27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
			27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
		},
		// scattered elements, may allocate
		trueCapMin: 12,
		trueCapMax: 101,
		// scattered elements, may allocate
		falseCapMin: 100,
		falseCapMax: 101,
	},
	{
		name:        "true_end",
		input:       []int{2, 3, 4, 6, 8, 1, 1},
		testFunc:    func(v int) bool { return v == 1 },
		expectTrue:  []int{1, 1},
		expectFalse: []int{2, 3, 4, 6, 8},
		trueCapMin:  2,
		trueCapMax:  2,
		falseCapMin: 6,
		falseCapMax: 7,
	},
	{
		name:        "mixed_split_first",
		input:       []int{1, 2, 1, 4, 6, 8},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8},
		expectFalse: []int{1, 1},
		// scattered elements, may allocate
		trueCapMin: 4,
		trueCapMax: 5,
		// scattered elements, may allocate
		falseCapMin: 5,
		falseCapMax: 6,
	},
	{
		name:        "mixed_split_second",
		input:       []int{2, 1, 4, 6, 8},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8},
		expectFalse: []int{1},
		// prefix + suffix, may allocate
		trueCapMin: 4,
		trueCapMax: 5,
		// single element, may allocate
		falseCapMin: 4,
		falseCapMax: 5,
	},
	{
		name:        "mixed_split_middle",
		input:       []int{2, 4, 6, 1, 3, 5, 8},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8},
		expectFalse: []int{1, 3, 5},
		// prefix + suffix, may allocate
		trueCapMin: 4,
		trueCapMax: 7,
		// consecutive middle section, may return view
		falseCapMin: 4,
		falseCapMax: 7,
	},
	{
		name:        "mixed_split_last",
		input:       []int{2, 4, 6, 1},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6},
		expectFalse: []int{1},
		// prefix only, may return view
		trueCapMin: 3,
		trueCapMax: 4,
		// single element at end, may return view
		falseCapMin: 1,
		falseCapMax: 1,
	},
	{
		name:     "mixed_split_large",
		input:    sliceLargeInput,
		testFunc: func(v int) bool { return v%2 == 0 },
		expectTrue: []int{
			0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26,
			28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50,
			2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26,
			28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50,
		},
		expectFalse: []int{
			1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25,
			27, 29, 31, 33, 35, 37, 39, 41, 43, 45, 47, 49,
			1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25,
			27, 29, 31, 33, 35, 37, 39, 41, 43, 45, 47, 49,
		},
		trueCapMin:  100,
		trueCapMax:  101,
		falseCapMin: 99,
		falseCapMax: 101,
	},
	{
		name:        "consecutive_start_after_one_false",
		input:       []int{1, 2, 3, 4, 5},
		testFunc:    func(v int) bool { return v >= 2 },
		expectTrue:  []int{2, 3, 4, 5},
		expectFalse: []int{1},
		// consecutive suffix, may return view
		trueCapMin: 4,
		trueCapMax: 4,
		// single element prefix, may return view
		falseCapMin: 4,
		falseCapMax: 5,
	},
	{
		name:        "consecutive_start_after_multiple_false",
		input:       []int{1, 3, 5, 6, 7, 8},
		testFunc:    func(v int) bool { return v >= 6 },
		expectTrue:  []int{6, 7, 8},
		expectFalse: []int{1, 3, 5},
		// consecutive suffix, may return view
		trueCapMin: 3,
		trueCapMax: 3,
		// non-consecutive prefix, may allocate
		falseCapMin: 5,
		falseCapMax: 6,
	},
	{
		name:        "consecutive_middle_chunk_only",
		input:       []int{1, 3, 6, 7, 8, 9, 5, 11},
		testFunc:    func(v int) bool { return v >= 6 && v <= 9 },
		expectTrue:  []int{6, 7, 8, 9},
		expectFalse: []int{1, 3, 5, 11},
		// consecutive middle chunk, may return view
		trueCapMin: 6,
		trueCapMax: 6,
		// scattered elements, may allocate
		falseCapMin: 4,
		falseCapMax: 8,
	},
	{
		name:        "consecutive_end_chunk_only",
		input:       []int{1, 3, 5, 8, 9, 10},
		testFunc:    func(v int) bool { return v >= 8 },
		expectTrue:  []int{8, 9, 10},
		expectFalse: []int{1, 3, 5},
		// consecutive suffix, may return view
		trueCapMin: 3,
		trueCapMax: 3,
		// non-consecutive prefix, may allocate
		falseCapMin: 5,
		falseCapMax: 6,
	},
	{
		name:        "prefix_plus_consecutive_suffix",
		input:       []int{2, 4, 1, 3, 8, 10, 12},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 8, 10, 12},
		expectFalse: []int{1, 3},
		// prefix + suffix, may allocate
		trueCapMin: 5,
		trueCapMax: 7,
		// consecutive middle section, may return view
		falseCapMin: 5,
		falseCapMax: 7,
	},
	{
		name:        "prefix_plus_non_consecutive_suffix",
		input:       []int{2, 4, 1, 8, 9, 10},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 8, 10},
		expectFalse: []int{1, 9},
		// prefix + suffix, may allocate
		trueCapMin: 5,
		trueCapMax: 6,
		// scattered elements, may allocate
		falseCapMin: 3,
		falseCapMax: 6,
	},
	{
		name:        "consecutive_chunk_then_gap_then_more",
		input:       []int{1, 6, 7, 8, 3, 5, 12, 14},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{6, 8, 12, 14},
		expectFalse: []int{1, 7, 3, 5},
		// scattered elements, may allocate
		trueCapMin: 6,
		trueCapMax: 8,
		// scattered elements, may allocate
		falseCapMin: 7,
		falseCapMax: 8,
	},
	{
		name:        "single_true_after_false",
		input:       []int{1, 3, 5, 8},
		testFunc:    func(v int) bool { return v == 8 },
		expectTrue:  []int{8},
		expectFalse: []int{1, 3, 5},
		// single element at end, may return view
		trueCapMin: 1,
		trueCapMax: 1,
		// non-consecutive prefix, may allocate
		falseCapMin: 3,
		falseCapMax: 4,
	},
	{
		name:        "consecutive_at_very_end",
		input:       []int{1, 3, 5, 7, 8, 10, 12},
		testFunc:    func(v int) bool { return v >= 8 },
		expectTrue:  []int{8, 10, 12},
		expectFalse: []int{1, 3, 5, 7},
		// consecutive suffix, may return view
		trueCapMin: 3,
		trueCapMax: 3,
		// non-consecutive prefix, may allocate
		falseCapMin: 6,
		falseCapMax: 7,
	},
	{
		name:        "prefix_only_no_suffix_true",
		input:       []int{2, 4, 6, 1, 3, 5},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6},
		expectFalse: []int{1, 3, 5},
		// prefix only, may return view
		trueCapMin: 5,
		trueCapMax: 6,
		// consecutive suffix, may return view
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "long_consecutive_in_middle",
		input:       []int{1, 3, 10, 11, 12, 13, 14, 15, 5, 7},
		testFunc:    func(v int) bool { return v >= 10 && v <= 15 },
		expectTrue:  []int{10, 11, 12, 13, 14, 15},
		expectFalse: []int{1, 3, 5, 7},
		// long consecutive middle chunk, may return view
		trueCapMin: 8,
		trueCapMax: 8,
		// scattered elements, may allocate
		falseCapMin: 4,
		falseCapMax: 10,
	},
	{
		name:        "multiple_gaps_non_consecutive",
		input:       []int{1, 2, 4, 5, 6, 8, 9, 10},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8, 10},
		expectFalse: []int{1, 5, 9},
		// scattered elements, may allocate
		trueCapMin: 6,
		trueCapMax: 8,
		// scattered elements, may allocate
		falseCapMin: 6,
		falseCapMax: 8,
	},
	{
		name:        "consecutive_with_single_gap",
		input:       []int{1, 6, 7, 8, 9, 5, 12},
		testFunc:    func(v int) bool { return v >= 6 },
		expectTrue:  []int{6, 7, 8, 9, 12},
		expectFalse: []int{1, 5},
		// consecutive chunk + suffix, may allocate
		trueCapMin: 5,
		trueCapMax: 7,
		// scattered elements, may allocate
		falseCapMin: 3,
		falseCapMax: 7,
	},
	{
		name:        "large_consecutive_chunk_optimization",
		input:       []int{1, 3, 5, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
		testFunc:    func(v int) bool { return v >= 10 },
		expectTrue:  []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
		expectFalse: []int{1, 3, 5},
		// large consecutive suffix, may return view
		trueCapMin: 16,
		trueCapMax: 16,
		// non-consecutive prefix, may allocate
		falseCapMin: 18,
		falseCapMax: 19,
	},
}

var sliceMultipleTestCases = []struct {
	name        string
	slices      [][]int
	testFunc    func(v int) bool
	expectTrue  []int
	expectFalse []int
	trueCapMin  int
	trueCapMax  int
	falseCapMin int
	falseCapMax int
}{
	{
		name:        "two_empty_slices",
		slices:      [][]int{{}, {}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  nil,
		expectFalse: nil,
		trueCapMin:  0,
		trueCapMax:  0,
		falseCapMin: 0,
		falseCapMax: 0,
	},
	{
		name:        "empty_first_nonempty_second",
		slices:      [][]int{{}, {1, 2, 3}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3},
		expectFalse: []int{2},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "nonempty_first_empty_second",
		slices:      [][]int{{1, 2, 3}, {}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3},
		expectFalse: []int{2},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "both_with_matching_elements",
		slices:      [][]int{{1, 2, 3}, {4, 5, 6}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3, 5},
		expectFalse: []int{2, 4, 6},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "first_no_matches_second_has_matches",
		slices:      [][]int{{2, 4, 6}, {1, 3, 5}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3, 5},
		expectFalse: []int{2, 4, 6},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "first_has_matches_second_no_matches",
		slices:      [][]int{{1, 3, 5}, {2, 4, 6}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3, 5},
		expectFalse: []int{2, 4, 6},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "three_slices_mixed_results",
		slices:      [][]int{{1, 2}, {}, {3, 4, 5}},
		testFunc:    func(v int) bool { return v >= 3 },
		expectTrue:  []int{3, 4, 5},
		expectFalse: []int{1, 2},
		trueCapMin:  3,
		trueCapMax:  5,
		falseCapMin: 2,
		falseCapMax: 5,
	},
	{
		name:        "all_elements_match",
		slices:      [][]int{{1, 3}, {5, 7}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3, 5, 7},
		expectFalse: nil,
		trueCapMin:  4,
		trueCapMax:  4,
		falseCapMin: 0,
		falseCapMax: 4,
	},
	{
		name:        "no_elements_match",
		slices:      [][]int{{2, 4}, {6, 8}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  nil,
		expectFalse: []int{2, 4, 6, 8},
		trueCapMin:  0,
		trueCapMax:  4,
		falseCapMin: 4,
		falseCapMax: 4,
	},
	{
		name:        "single_slice_compatibility",
		slices:      [][]int{{1, 2, 3, 4, 5}},
		testFunc:    func(v int) bool { return v >= 3 },
		expectTrue:  []int{3, 4, 5},
		expectFalse: []int{1, 2},
		trueCapMin:  3,
		trueCapMax:  5,
		falseCapMin: 2,
		falseCapMax: 5,
	},
	{
		name:        "capacity_preservation_test",
		slices:      [][]int{make([]int, 0, 10), {1, 2, 3}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2, 3},
		expectFalse: nil,
		trueCapMin:  3,
		trueCapMax:  10,
		falseCapMin: 0,
		falseCapMax: 10,
	},
	{
		name:        "multiple_empty_slices_with_capacity",
		slices:      [][]int{make([]int, 0, 5), make([]int, 0, 8), {1, 2}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2},
		expectFalse: nil,
		trueCapMin:  2,
		trueCapMax:  8,
		falseCapMin: 0,
		falseCapMax: 8,
	},
	{
		name:        "zero_slices",
		slices:      [][]int{},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  nil,
		expectFalse: nil,
		trueCapMin:  0,
		trueCapMax:  0,
		falseCapMin: 0,
		falseCapMax: 0,
	},
	{
		name:        "first_empty_second_empty_third_has_results",
		slices:      [][]int{{}, {}, {1, 2, 3}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3},
		expectFalse: []int{2},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "many_empties_then_results",
		slices:      [][]int{{}, {}, {}, {}, {5, 6, 7}},
		testFunc:    func(v int) bool { return v > 5 },
		expectTrue:  []int{6, 7},
		expectFalse: []int{5},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "trigger_single_result_optimization_filter",
		slices:      [][]int{{}, {}, {1, 2, 3}, {}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2, 3},
		expectFalse: nil,
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 0,
		falseCapMax: 3,
	},
	{
		name:        "alternating_empty_nonempty",
		slices:      [][]int{{}, {1}, {}, {2}, {}, {3}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  []int{1, 3},
		expectFalse: []int{2},
		trueCapMin:  2,
		trueCapMax:  2,
		falseCapMin: 1,
		falseCapMax: 2,
	},
	{
		name:        "all_but_first_empty",
		slices:      [][]int{{1, 2, 3}, {}, {}, {}},
		testFunc:    func(v int) bool { return v >= 2 },
		expectTrue:  []int{2, 3},
		expectFalse: []int{1},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "split_one_empty_true_one_empty_false",
		slices:      [][]int{{}, {2, 4, 6}},
		testFunc:    func(v int) bool { return v%2 == 1 },
		expectTrue:  nil,
		expectFalse: []int{2, 4, 6},
		trueCapMin:  0,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "split_triggers_dual_single_optimization",
		slices:      [][]int{{}, {1, 2, 3}},
		testFunc:    func(v int) bool { return v >= 2 },
		expectTrue:  []int{2, 3},
		expectFalse: []int{1},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "empty_capacity_first_higher",
		slices:      [][]int{make([]int, 0, 20), make([]int, 0, 5), {1, 2, 3}},
		testFunc:    func(v int) bool { return v < 0 },
		expectTrue:  nil,
		expectFalse: []int{1, 2, 3},
		trueCapMin:  0,
		trueCapMax:  20,
		falseCapMin: 3,
		falseCapMax: 20,
	},
	{
		name:        "empty_capacity_second_higher",
		slices:      [][]int{make([]int, 0, 5), make([]int, 0, 20), {1, 2, 3}},
		testFunc:    func(v int) bool { return v < 0 },
		expectTrue:  nil,
		expectFalse: []int{1, 2, 3},
		trueCapMin:  0,
		trueCapMax:  20,
		falseCapMin: 3,
		falseCapMax: 20,
	},
	{
		name:        "first_slice_filtered_fully",
		slices:      [][]int{make([]int, 3), {1, 2}, {3, 4}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2, 3, 4},
		expectFalse: []int{0, 0, 0},
		trueCapMin:  4,
		trueCapMax:  4,
		falseCapMin: 3,
		falseCapMax: 4,
	},
	{
		name:        "first_slice_insufficient_capacity",
		slices:      [][]int{{0, 1}, {2}, {3, 4}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2, 3, 4},
		expectFalse: []int{0},
		trueCapMin:  4,
		trueCapMax:  4,
		falseCapMin: 1,
		falseCapMax: 4,
	},
	{
		name:        "first_slice_provides_capacity",
		slices:      [][]int{{6, 1, 6, 1, 2, 3, 4, 5}, {6, 7}, {8, 9, 10}},
		testFunc:    func(v int) bool { return v > 5 },
		expectTrue:  []int{6, 6, 6, 7, 8, 9, 10},
		expectFalse: []int{1, 1, 2, 3, 4, 5},
		trueCapMin:  7,
		trueCapMax:  8,
		falseCapMin: 6,
		falseCapMax: 8,
	},
	{
		name:        "first_slice_view_prevents_concat_inplace",
		slices:      [][]int{{2, 4, 6}, {1, 3, 8}},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8},
		expectFalse: []int{1, 3},
		trueCapMin:  4,
		trueCapMax:  4,
		falseCapMin: 2,
		falseCapMax: 3,
	},
	{
		name:        "first_slice_allocation_allows_concat_inplace",
		slices:      [][]int{{1, 2, 4, 3, 6}, {8, 10, 12}},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6, 8, 10, 12},
		expectFalse: []int{1, 3},
		trueCapMin:  6,
		trueCapMax:  6,
		falseCapMin: 3,
		falseCapMax: 5,
	},
	{
		name:        "head_slice_replacement_with_capacity_preference",
		slices:      [][]int{make([]int, 0, 2), make([]int, 0, 8), {1, 2, 3}},
		testFunc:    func(v int) bool { return v > 1 },
		expectTrue:  []int{2, 3},
		expectFalse: []int{1},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "head_slice_not_replaced_when_has_results",
		slices:      [][]int{{2}, make([]int, 0, 10), {4, 5, 6}},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6},
		expectFalse: []int{5},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 1,
		falseCapMax: 3,
	},
	{
		name:        "capacity_tracking_heavy_filtering",
		slices:      [][]int{{1, 2, 3, 4, 5}, {6}, {7, 8, 9}},
		testFunc:    func(v int) bool { return v >= 6 },
		expectTrue:  []int{6, 7, 8, 9},
		expectFalse: []int{1, 2, 3, 4, 5},
		trueCapMin:  4,
		trueCapMax:  5,
		falseCapMin: 5,
		falseCapMax: 5,
	},
	{
		name:        "capacity_insufficient_forces_regular_concat",
		slices:      [][]int{{1}, {2, 3, 4, 5, 6, 7, 8, 9, 10}},
		testFunc:    func(v int) bool { return v >= 2 },
		expectTrue:  []int{2, 3, 4, 5, 6, 7, 8, 9, 10},
		expectFalse: []int{1},
		trueCapMin:  9,
		trueCapMax:  9,
		falseCapMin: 1,
		falseCapMax: 1,
	},
	{
		name:        "split_first_slice_view_true_prevents_inplace",
		slices:      [][]int{{2, 4, 6}, {1, 7, 9}},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 4, 6},
		expectFalse: []int{1, 7, 9},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "split_first_slice_view_false_prevents_inplace",
		slices:      [][]int{{1, 3, 5}, {2, 8, 9}},
		testFunc:    func(v int) bool { return v%2 == 0 },
		expectTrue:  []int{2, 8},
		expectFalse: []int{1, 3, 5, 9},
		trueCapMin:  2,
		trueCapMax:  3,
		falseCapMin: 4,
		falseCapMax: 4,
	},
	{
		name:        "firstCapacity_boundary_equal_zero",
		slices:      [][]int{{1, 2, 3, 4, 5}, {6, 7, 8}},
		testFunc:    func(v int) bool { return v == 7 },
		expectTrue:  []int{7},
		expectFalse: []int{1, 2, 3, 4, 5, 6, 8},
		trueCapMin:  1,
		trueCapMax:  3,
		falseCapMin: 7,
		falseCapMax: 8,
	},
	{
		name:        "firstCapacity_negative_prevents_inplace",
		slices:      [][]int{{1}, {2, 3, 4, 5, 6}},
		testFunc:    func(v int) bool { return v >= 2 },
		expectTrue:  []int{2, 3, 4, 5, 6},
		expectFalse: []int{1},
		trueCapMin:  5,
		trueCapMax:  5,
		falseCapMin: 1,
		falseCapMax: 5,
	},
	{
		name:        "last_slice_head_replacement_capacity",
		slices:      [][]int{{}, {1, 2, 3}, {}},
		testFunc:    func(v int) bool { return false },
		expectTrue:  nil,
		expectFalse: []int{1, 2, 3},
		trueCapMin:  0,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "last_slice_head_replacement_results",
		slices:      [][]int{{}, {1, 2, 3}, {4, 5, 6}},
		testFunc:    func(v int) bool { return v >= 4 },
		expectTrue:  []int{4, 5, 6},
		expectFalse: []int{1, 2, 3},
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 3,
		falseCapMax: 3,
	},
	{
		name:        "firstCapacity_tracking_multiple",
		slices:      [][]int{{1, 2, 3, 4, 5, 6, 7, 8}, {9}, {10}},
		testFunc:    func(v int) bool { return v >= 7 },
		expectTrue:  []int{7, 8, 9, 10},
		expectFalse: []int{1, 2, 3, 4, 5, 6},
		trueCapMin:  4,
		trueCapMax:  8,
		falseCapMin: 6,
		falseCapMax: 8,
	},
	{
		name:        "currCapacity_vs_firstCapacity",
		slices:      [][]int{{1, 2, 3}, {4, 5, 6, 7, 8}},
		testFunc:    func(v int) bool { return v == 4 },
		expectTrue:  []int{4},
		expectFalse: []int{1, 2, 3, 5, 6, 7, 8},
		trueCapMin:  1,
		trueCapMax:  5,
		falseCapMin: 7,
		falseCapMax: 7,
	},
	{
		name:        "head_not_replaced_has_results",
		slices:      [][]int{{}, {1, 2}, {}},
		testFunc:    func(v int) bool { return v > 0 },
		expectTrue:  []int{1, 2},
		expectFalse: nil,
		trueCapMin:  2,
		trueCapMax:  2,
		falseCapMin: 0,
		falseCapMax: 2,
	},
	{
		name:        "concat_boundary_totalSize_equals_capacity",
		slices:      [][]int{{1}, {2, 3}},
		testFunc:    func(v int) bool { return true },
		expectTrue:  []int{1, 2, 3},
		expectFalse: nil,
		trueCapMin:  3,
		trueCapMax:  3,
		falseCapMin: 0,
		falseCapMax: 3,
	},
	{
		name:        "concat_boundary_totalSize_exceeds_capacity",
		slices:      [][]int{{1}, {2, 3, 4}},
		testFunc:    func(v int) bool { return true },
		expectTrue:  []int{1, 2, 3, 4},
		expectFalse: nil,
		trueCapMin:  4,
		trueCapMax:  4,
		falseCapMin: 0,
		falseCapMax: 4,
	},
}

func sliceDup[T any](s []T) []T {
	result := make([]T, len(s))
	copy(result, s)
	return result
}

func TestSliceFilter(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice := SliceFilter(tt.testFunc, tt.input)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)

			falseSlice := SliceFilter(func(v int) bool { return !tt.testFunc(v) }, tt.input)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice := SliceFilter(tt.testFunc, tt.slices...)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)

			falseSlice := SliceFilter(func(v int) bool { return !tt.testFunc(v) }, tt.slices...)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	t.Run("zero_slices_direct", func(t *testing.T) {
		result := SliceFilter(func(v int) bool { return v > 0 })
		assert.Nil(t, result)
	})
}

var sliceFilterTransformTestCases = []struct {
	name           string
	input          []int
	predicate      func(int) bool
	transform      func(int) string
	expectedResult []string
}{
	{
		name:           "nil",
		input:          nil,
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: nil,
	},
	{
		name:           "empty",
		input:          []int{},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{},
	},
	{
		name:           "all_true",
		input:          []int{1, 2, 3, 4, 5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{"val_1", "val_2", "val_3", "val_4", "val_5"},
	},
	{
		name:           "all_false",
		input:          []int{-1, -2, -3, -4, -5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{},
	},
	{
		name:           "mixed_pattern",
		input:          []int{1, -2, 3, -4, 5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{"val_1", "val_3", "val_5"},
	},
	{
		name:           "true_prefix",
		input:          []int{1, 2, 3, -4, -5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{"val_1", "val_2", "val_3"},
	},
	{
		name:           "true_suffix",
		input:          []int{-1, -2, 3, 4, 5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{"val_3", "val_4", "val_5"},
	},
	{
		name:           "single_true",
		input:          []int{5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{"val_5"},
	},
	{
		name:           "single_false",
		input:          []int{-5},
		predicate:      func(v int) bool { return v > 0 },
		transform:      func(v int) string { return fmt.Sprintf("val_%d", v) },
		expectedResult: []string{},
	},
	{
		name:           "type_change_int_to_float",
		input:          []int{2, 4, 6, 8, 10},
		predicate:      func(v int) bool { return v%4 == 0 },
		transform:      func(v int) string { return fmt.Sprintf("%.1f", float64(v)/2) },
		expectedResult: []string{"2.0", "4.0"},
	},
}

func TestSliceFilterTransform(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceFilterTransformTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := SliceFilterTransform(tt.predicate, tt.transform, tt.input)
			if len(tt.expectedResult) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}

	intNoTransform := func(i int) int { return i }

	for i, tt := range sliceTestCases {
		t.Run("basic-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueResult := SliceFilterTransform(tt.testFunc, intNoTransform, tt.input)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueResult)
			} else {
				assert.Equal(t, tt.expectTrue, trueResult)
			}

			falseResult := SliceFilterTransform(func(i int) bool { return !tt.testFunc(i) }, intNoTransform, tt.input)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseResult)
			} else {
				assert.Equal(t, tt.expectFalse, falseResult)
			}
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice := SliceFilterTransform(tt.testFunc, intNoTransform, tt.slices...)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)

			falseSlice := SliceFilterTransform(func(v int) bool { return !tt.testFunc(v) }, intNoTransform, tt.slices...)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	t.Run("zero_slices_direct", func(t *testing.T) {
		result := SliceFilterTransform(func(v int) bool { return v > 0 }, func(v int) string { return fmt.Sprintf("val_%d", v) })
		assert.Nil(t, result)
	})

	t.Run("multiple_slices", func(t *testing.T) {
		slice1 := []int{1, -2, 3}
		slice2 := []int{-4, 5, 6}
		slice3 := []int{7, -8}

		result := SliceFilterTransform(
			func(v int) bool { return v > 0 },
			func(v int) string { return fmt.Sprintf("val_%d", v) },
			slice1, slice2, slice3,
		)
		expected := []string{"val_1", "val_3", "val_5", "val_6", "val_7"}
		assert.Equal(t, expected, result)
	})
}

func TestSliceFilterTransformErr(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceFilterTransformTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result, err := SliceFilterTransformErr(tt.predicate, func(i int) (string, error) {
				return tt.transform(i), nil
			}, tt.input)
			require.NoError(t, err)
			if len(tt.expectedResult) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}

	intNoTransform := func(i int) (int, error) { return i, nil }

	for i, tt := range sliceTestCases {
		t.Run("basic-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueResult, err := SliceFilterTransformErr(tt.testFunc, intNoTransform, tt.input)
			require.NoError(t, err)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueResult)
			} else {
				assert.Equal(t, tt.expectTrue, trueResult)
			}

			falseResult, err := SliceFilterTransformErr(func(i int) bool { return !tt.testFunc(i) },
				intNoTransform, tt.input)
			require.NoError(t, err)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseResult)
			} else {
				assert.Equal(t, tt.expectFalse, falseResult)
			}
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice, err := SliceFilterTransformErr(tt.testFunc, intNoTransform, tt.slices...)
			require.NoError(t, err)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}

			falseSlice, err := SliceFilterTransformErr(func(v int) bool { return !tt.testFunc(v) },
				intNoTransform, tt.slices...)
			require.NoError(t, err)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
		})
	}

	t.Run("error on first element", func(t *testing.T) {
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i > 0 }, // filter positive numbers
			func(i int) (string, error) { return "", errors.New("transform error") },
			[]int{1, 2, 3},
		)
		require.Error(t, err)
		assert.Equal(t, "transform error", err.Error())
		assert.Empty(t, result)
	})

	t.Run("error in middle with partial results", func(t *testing.T) {
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i > 0 }, // filter positive numbers
			func(i int) (string, error) {
				if i == 3 {
					return "", errors.New("error on 3")
				}
				return strconv.Itoa(i * 10), nil
			},
			[]int{1, 2, 3, 4},
		)
		require.Error(t, err)
		assert.Equal(t, "error on 3", err.Error())
		assert.Equal(t, []string{"10", "20"}, result) // partial results before error
	})

	t.Run("error in consecutive section", func(t *testing.T) {
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i%2 == 0 }, // filter even numbers
			func(i int) (string, error) {
				if i == 4 {
					return "", errors.New("error on 4")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2, 4, 6}, // 2,4,6 pass filter but error on 4
		)
		require.Error(t, err)
		assert.Equal(t, "error on 4", err.Error())
		assert.Equal(t, []string{"2"}, result) // partial results before error
	})

	t.Run("error in multi-slice scenario", func(t *testing.T) {
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i > 0 },
			func(i int) (string, error) {
				if i == 4 {
					return "", errors.New("error on 4")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2}, []int{3, 4, 5}, // error occurs in second slice
		)
		require.Error(t, err)
		assert.Equal(t, "error on 4", err.Error())
		assert.Equal(t, []string{"1", "2", "3"}, result) // results from first slice + partial from second
	})

	t.Run("error after filter skips elements", func(t *testing.T) {
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i%2 == 1 }, // filter odd numbers
			func(i int) (string, error) {
				if i == 5 {
					return "", errors.New("error on 5")
				}
				return strconv.Itoa(i * 100), nil
			},
			[]int{1, 2, 3, 4, 5, 6, 7}, // 1,3,5,7 pass filter but error on 5
		)
		require.Error(t, err)
		assert.Equal(t, "error on 5", err.Error())
		assert.Equal(t, []string{"100", "300"}, result) // partial results before error
	})

	t.Run("error in consecutive section inner loop", func(t *testing.T) {
		// This targets the error path while processing consecutive true elements after finding non-consecutive trues
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i != 4 }, // filter all except 4
			func(i int) (string, error) {
				if i == 2 { // error on second element in consecutive section
					return "", errors.New("error on 2")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2, 3, 4, 5}, // 1,2,3 consecutive, 4 skipped, 5 true -> triggers non-consecutive path
		)
		require.Error(t, err)
		assert.Equal(t, "error on 2", err.Error())
		assert.Equal(t, []string{"1"}, result) // partial result before error in consecutive section
	})

	t.Run("error in true-false-true pattern", func(t *testing.T) {
		// This targets the specific error path where an error occurs in the true+false+true section after processing the initial true elements
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i != 3 }, // filter all except 3 (creates true+false+true pattern)
			func(i int) (string, error) {
				if i == 4 { // error on element in second true section
					return "", errors.New("error on 4")
				}
				return strconv.Itoa(i * 10), nil
			},
			[]int{1, 2, 3, 4, 5}, // 1,2 true, 3 false, 4,5 true -> triggers true+false+true path
		)
		require.Error(t, err)
		assert.Equal(t, "error on 4", err.Error())
		assert.Equal(t, []string{"10", "20"}, result) // results from initial true section only
	})

	t.Run("error in non-consecutive allocation path - consecutive section", func(t *testing.T) {
		// This targets error during consecutive section transformation
		// Pattern: false, true+, false+, true -> triggers "Found another true after false" path
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i != 1 && i != 4 }, // filter out 1 and 4
			func(i int) (string, error) {
				if i == 3 { // error in consecutive section (2,3 are consecutive trues)
					return "", errors.New("error on 3")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2, 3, 4, 5}, // 1 false, 2,3 true consecutive, 4 false, 5 true -> non-consecutive
		)
		require.Error(t, err)
		assert.Equal(t, "error on 3", err.Error())
		assert.Equal(t, []string{"2"}, result) // partial result from consecutive section before error
	})

	t.Run("error in non-consecutive allocation path - slice[j]", func(t *testing.T) {
		// This targets error during slice[j] transformation
		// Pattern: false, true+, false+, true -> triggers "Found another true after false" path
		result, err := SliceFilterTransformErr(
			func(i int) bool { return i != 1 && i != 4 }, // filter out 1 and 4
			func(i int) (string, error) {
				if i == 5 { // error on slice[j] (the later true element)
					return "", errors.New("error on 5")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2, 3, 4, 5}, // 1 false, 2,3 true consecutive, 4 false, 5 true -> non-consecutive
		)
		require.Error(t, err)
		assert.Equal(t, "error on 5", err.Error())
		assert.Equal(t, []string{"2", "3"}, result) // consecutive section processed, error on slice[j]
	})
}

func TestSliceFilterTransformInto(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceFilterTransformTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := SliceFilterTransformInto([]string{}, tt.predicate, tt.transform, tt.input)
			if len(tt.expectedResult) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}

	intNoTransform := func(i int) int { return i }

	for i, tt := range sliceTestCases {
		t.Run("basic-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueResult := SliceFilterTransformInto([]int{}, tt.testFunc, intNoTransform, tt.input)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueResult)
			} else {
				assert.Equal(t, tt.expectTrue, trueResult)
			}

			falseResult := SliceFilterTransformInto([]int{}, func(i int) bool { return !tt.testFunc(i) }, intNoTransform, tt.input)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseResult)
			} else {
				assert.Equal(t, tt.expectFalse, falseResult)
			}
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice := SliceFilterTransformInto([]int{}, tt.testFunc, intNoTransform, tt.slices...)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}

			falseSlice := SliceFilterTransformInto([]int{}, func(v int) bool { return !tt.testFunc(v) }, intNoTransform, tt.slices...)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
		})
	}
}

func TestSliceFilterTransformErrInto(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceFilterTransformTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result, err := SliceFilterTransformErrInto([]string{}, tt.predicate, func(i int) (string, error) {
				return tt.transform(i), nil
			}, tt.input)
			require.NoError(t, err)
			if len(tt.expectedResult) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}

	intNoTransform := func(i int) (int, error) { return i, nil }

	for i, tt := range sliceTestCases {
		t.Run("basic-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueResult, err := SliceFilterTransformErrInto([]int{}, tt.testFunc, intNoTransform, tt.input)
			require.NoError(t, err)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueResult)
			} else {
				assert.Equal(t, tt.expectTrue, trueResult)
			}

			falseResult, err := SliceFilterTransformErrInto([]int{}, func(i int) bool { return !tt.testFunc(i) },
				intNoTransform, tt.input)
			require.NoError(t, err)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseResult)
			} else {
				assert.Equal(t, tt.expectFalse, falseResult)
			}
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice, err := SliceFilterTransformErrInto([]int{}, tt.testFunc, intNoTransform, tt.slices...)
			require.NoError(t, err)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}

			falseSlice, err := SliceFilterTransformErrInto([]int{}, func(v int) bool { return !tt.testFunc(v) },
				intNoTransform, tt.slices...)
			require.NoError(t, err)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
		})
	}

	t.Run("error on first element", func(t *testing.T) {
		dest := []string{"existing"}
		result, err := SliceFilterTransformErrInto(
			dest,
			func(i int) bool { return i > 0 }, // filter positive numbers
			func(i int) (string, error) { return "", errors.New("transform error") },
			[]int{1, 2, 3},
		)
		require.Error(t, err)
		assert.Equal(t, "transform error", err.Error())
		assert.Equal(t, []string{"existing"}, result) // no new elements added
	})

	t.Run("error in middle with partial results", func(t *testing.T) {
		dest := []string{"start"}
		result, err := SliceFilterTransformErrInto(
			dest,
			func(i int) bool { return i > 0 }, // filter positive numbers
			func(i int) (string, error) {
				if i == 3 {
					return "", errors.New("error on 3")
				}
				return strconv.Itoa(i * 10), nil
			},
			[]int{1, 2, 3, 4},
		)
		require.Error(t, err)
		assert.Equal(t, "error on 3", err.Error())
		assert.Equal(t, []string{"start", "10", "20"}, result) // partial results appended
	})

	t.Run("error across multiple input slices", func(t *testing.T) {
		dest := []string{}
		result, err := SliceFilterTransformErrInto(
			dest,
			func(i int) bool { return i%2 == 1 }, // filter odd numbers
			func(i int) (string, error) {
				if i == 5 {
					return "", errors.New("error on 5")
				}
				return strconv.Itoa(i * 100), nil
			},
			[]int{1, 2, 3}, []int{4, 5, 6, 7}, // error in second slice
		)
		require.Error(t, err)
		assert.Equal(t, "error on 5", err.Error())
		assert.Equal(t, []string{"100", "300"}, result) // results from first slice only
	})
}

func TestSliceFilterInto(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			resultTrue := SliceFilterInto(make([]int, 0), tt.testFunc, tt.input)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, resultTrue)
			} else {
				assert.Equal(t, tt.expectTrue, resultTrue)
			}

			resultFalse := SliceFilterInto(make([]int, 0), func(v int) bool { return !tt.testFunc(v) }, tt.input)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, resultFalse)
			} else {
				assert.Equal(t, tt.expectFalse, resultFalse)
			}
		})
	}
}

func TestSliceFilterInPlace(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			// make a copy of the test input to avoid changing it
			trueSlice := SliceFilterInPlace(tt.testFunc, sliceDup(tt.input))
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}

			// make a copy of the test input to avoid changing it
			falseSlice := SliceFilterInPlace(func(v int) bool { return !tt.testFunc(v) }, sliceDup(tt.input))
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			// make a copy of the test input to avoid changing it
			slices := make([][]int, len(tt.slices))
			for i := range slices {
				slices[i] = sliceDup(tt.slices[i])
			}

			trueSlice := SliceFilterInPlace(tt.testFunc, slices...)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)

			// replace with test case copies
			for i := range slices {
				slices[i] = sliceDup(tt.slices[i])
			}

			falseSlice := SliceFilterInPlace(func(v int) bool { return !tt.testFunc(v) }, slices...)
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	t.Run("slice_pick", func(t *testing.T) {
		slice1 := []int{1, 2, 3, 4, 5} // Will be filtered heavily, retained for capacity
		slice2 := []int{6, 7, 8, -8}   // result slice due to only one with results

		// This should hit the firstCapacity > 0 branch due to heavy filtering
		result := SliceFilterInPlace(func(v int) bool { return v >= 8 }, slice1, slice2)
		assert.Equal(t, []int{8}, result)
		assert.Equal(t, 4, cap(result))
	})

	t.Run("slice_concat_inplace_first", func(t *testing.T) {
		slice1 := []int{1, 2, 3, 4, 5} // Will be filtered heavily
		slice2 := make([]int, 0, 15)   // Empty but high capacity
		slice3 := []int{6, 7, 8, -8}   // first non-empty result, but co
		slice4 := []int{10}            // second slice with results

		// This should hit the firstCapacity > 0 branch due to heavy filtering
		result := SliceFilterInPlace(func(v int) bool { return v >= 8 }, slice1, slice2, slice3, slice4)
		assert.Equal(t, []int{8, 10}, result)
		assert.Equal(t, 5, cap(result))
	})

	t.Run("capacity_replacement_logic", func(t *testing.T) {
		// Test the specific "firstCapacity < currCapacity" branch
		slice1 := []int{1, 2}          // small capacity after filtering (all elements filtered out)
		slice2 := []int{3, 4, 5, 6, 7} // larger remaining capacity after filtering (all elements filtered out)

		result := SliceFilterInPlace(func(v int) bool { return false }, slice1, slice2)
		// Should prefer slice2's capacity since it has more remaining space after filtering
		assert.Empty(t, result)
		assert.Equal(t, 5, cap(result)) // Should use slice2's capacity
	})

	t.Run("capacity_replacement_with_results", func(t *testing.T) {
		// Test capacity replacement when one slice has results
		slice1 := make([]int, 0, 3)    // empty with small capacity
		slice2 := []int{1, 2, 3, 4, 5} // will have remaining capacity after filtering

		result := SliceFilterInPlace(func(v int) bool { return v > 3 }, slice1, slice2)
		assert.Equal(t, []int{4, 5}, result)
		// Should use slice2's capacity since it provided results and remaining capacity
		assert.Equal(t, 5, cap(result))
	})
}

func TestSliceSplit(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice, falseSlice := SliceSplit(tt.testFunc, tt.input)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice) // may be nil or input
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.Equal(t, tt.expectFalse, falseSlice)

			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	for i, tt := range sliceMultipleTestCases {
		t.Run("multi-"+strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			trueSlice, falseSlice := SliceSplit(tt.testFunc, tt.slices...)
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice)
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			if len(tt.expectFalse) == 0 {
				assert.Empty(t, falseSlice)
			} else {
				assert.Equal(t, tt.expectFalse, falseSlice)
			}
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}

	t.Run("zero_slices_direct", func(t *testing.T) {
		trueResult, falseResult := SliceSplit(func(v int) bool { return v > 0 })
		assert.Nil(t, trueResult)
		assert.Nil(t, falseResult)
	})

	t.Run("capacity_preference_when_concat_not_inplace_true", func(t *testing.T) {
		// Create scenario where views prevent inPlace concat and capacity selection is needed
		slice1 := make([]int, 0, 5)  // Empty with some capacity
		slice2 := make([]int, 0, 15) // Empty with higher capacity
		slice3 := []int{1, 2, 3, 4, 5}

		trueResult, falseResult := SliceSplit(func(v int) bool { return v > 10 }, slice1, slice2, slice3)
		assert.Empty(t, trueResult) // Should prefer higher capacity empty result
		assert.Equal(t, []int{1, 2, 3, 4, 5}, falseResult)
	})

	t.Run("capacity_preference_when_concat_not_inplace_false", func(t *testing.T) {
		// Create scenario where views prevent inPlace concat for false results
		slice1 := make([]int, 0, 5)  // Empty with some capacity
		slice2 := make([]int, 0, 15) // Empty with higher capacity
		slice3 := []int{11, 12, 13, 14, 15}

		trueResult, falseResult := SliceSplit(func(v int) bool { return v > 10 }, slice1, slice2, slice3)
		assert.Equal(t, []int{11, 12, 13, 14, 15}, trueResult)
		assert.Empty(t, falseResult) // Should prefer higher capacity empty result
	})

	t.Run("view_detection_prevents_inplace_concat", func(t *testing.T) {
		// Test case where singleSliceSplit returns views, preventing inPlace concat
		slice1 := []int{1, 2, 3, 4, 5} // Will return views
		slice2 := make([]int, 0, 20)   // High capacity empty slice
		slice3 := []int{6, 7, 8}       // More elements

		// This should exercise the view detection logic
		trueResult, falseResult := SliceSplit(func(v int) bool { return v%2 == 1 }, slice1, slice2, slice3)
		assert.Equal(t, []int{1, 3, 5, 7}, trueResult)
		assert.Equal(t, []int{2, 4, 6, 8}, falseResult)
	})

	t.Run("capacity_selection_with_view_preventing_inplace", func(t *testing.T) {
		// Force the specific branch: !trueConcatInPlace && len(trueResults[0]) == 0 && cap(trueResults[0]) < cap(tSlice)
		slice1 := make([]int, 0, 3)  // Empty with low capacity
		slice2 := []int{1, 3, 5}     // All true for predicate - will be a view, preventing inPlace
		slice3 := make([]int, 0, 15) // Empty with higher capacity - should be selected

		trueResult, falseResult := SliceSplit(func(v int) bool { return v%2 == 1 }, slice1, slice2, slice3)
		assert.Equal(t, []int{1, 3, 5}, trueResult)
		assert.Empty(t, falseResult)
		// The empty falseResult should use slice3's capacity due to the capacity preference logic
	})

	t.Run("mixed_view_nonview_capacity_selection", func(t *testing.T) {
		// Test the exact logic: first slice empty (non-view), second slice returns view, third slice has higher capacity
		slice1 := make([]int, 0, 2)    // Empty result with capacity 2
		slice2 := []int{1, 2, 3, 4, 5} // Mixed results - returns views, setting trueConcatInPlace=false
		slice3 := make([]int, 0, 20)   // Empty with much higher capacity

		// This should trigger both capacity selection branches due to view detection
		trueResult, falseResult := SliceSplit(func(v int) bool { return v > 3 }, slice1, slice2, slice3)
		assert.Equal(t, []int{4, 5}, trueResult)
		assert.Equal(t, []int{1, 2, 3}, falseResult)
	})

	t.Run("asymmetric_results_first_slice_mixed", func(t *testing.T) {
		slice1 := []int{1, 2}    // produces both true and false results
		slice2 := []int{3, 3, 3} // produces only false results

		trueResult, falseResult := SliceSplit(func(v int) bool { return v%2 == 0 }, slice1, slice2)
		assert.Equal(t, []int{2}, trueResult)
		assert.Equal(t, []int{1, 3, 3, 3}, falseResult)

		// trueResult should be a single allocated result
		assert.Len(t, trueResult, cap(trueResult))
		// falseResult should be concatenated from multiple slices
		assert.Len(t, falseResult, cap(falseResult))
	})

	t.Run("asymmetric_results_second_slice_mixed", func(t *testing.T) {
		slice1 := []int{4, 4, 4} // produces only true results
		slice2 := []int{1, 2}    // produces both true and false results

		trueResult, falseResult := SliceSplit(func(v int) bool { return v%2 == 0 }, slice1, slice2)
		assert.Equal(t, []int{4, 4, 4, 2}, trueResult)
		assert.Equal(t, []int{1}, falseResult)

		// trueResult should be concatenated (first slice is view + second slice allocated)
		assert.Len(t, trueResult, cap(trueResult))
		// falseResult should be a single allocated result
		assert.Len(t, falseResult, cap(falseResult))
	})
}

func TestSliceSplitInPlace(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			// make a copy of the test input to avoid changing it
			trueSlice, falseSlice := SliceSplitInPlace(tt.testFunc, sliceDup(tt.input))
			if len(tt.expectTrue) == 0 {
				assert.Empty(t, trueSlice) // may be nil or input
			} else {
				assert.Equal(t, tt.expectTrue, trueSlice)
			}
			assert.Equal(t, tt.expectFalse, falseSlice)

			// Validate capacity
			assert.GreaterOrEqual(t, cap(trueSlice), tt.trueCapMin)
			assert.LessOrEqual(t, cap(trueSlice), tt.trueCapMax)
			assert.GreaterOrEqual(t, cap(falseSlice), tt.falseCapMin)
			assert.LessOrEqual(t, cap(falseSlice), tt.falseCapMax)
		})
	}
}

func TestSliceSplitInPlaceUnstable(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceTestCases {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			// make a copy of the test input to avoid changing it
			trueSlice, falseSlice := SliceSplitInPlaceUnstable(tt.testFunc, sliceDup(tt.input))
			assert.ElementsMatch(t, tt.expectTrue, trueSlice)
			assert.ElementsMatch(t, tt.expectFalse, falseSlice)
		})
	}
}

func TestSliceFilterCapacity(t *testing.T) {
	t.Parallel()

	t.Run("all_pass_returns_original_slice", func(t *testing.T) {
		input := []int{2, 4, 6, 8}
		result := SliceFilter(func(v int) bool { return v%2 == 0 }, input)

		// Should return the exact same slice (view optimization)
		assert.Equal(t, input, result)
		assert.Equal(t, cap(input), cap(result))

		// Verify it's the same underlying array
		if len(input) > 0 && len(result) > 0 {
			assert.Equal(t, &input[0], &result[0])
		}
	})

	t.Run("consecutive_suffix_returns_view", func(t *testing.T) {
		input := []int{1, 3, 6, 7, 8, 9}
		result := SliceFilter(func(v int) bool { return v >= 6 }, input)

		// Should return a view of the suffix [6, 7, 8, 9]
		assert.Equal(t, []int{6, 7, 8, 9}, result)
		// Should share underlying array
		if len(result) > 0 {
			assert.Equal(t, &input[2], &result[0]) // result starts at input[2]
		}
	})

	t.Run("consecutive_prefix_returns_view", func(t *testing.T) {
		input := []int{2, 4, 6, 1, 3, 5}
		result := SliceFilter(func(v int) bool { return v%2 == 0 }, input)

		// Should return a view of prefix [2, 4, 6] - but this case requires allocation due to non-consecutive
		assert.Equal(t, []int{2, 4, 6}, result)
		// This case actually allocates because prefix + non-consecutive elements exist
		assert.GreaterOrEqual(t, cap(result), len(result))
	})

	t.Run("empty_result_retains_original_capacity", func(t *testing.T) {
		input := []int{1, 3, 5, 7}
		result := SliceFilter(func(v int) bool { return v%2 == 0 }, input)

		assert.Empty(t, result)
		// Empty result from SliceFilter returns slice[:0] which retains original capacity
		assert.Equal(t, cap(input), cap(result))
	})

	t.Run("non_consecutive_allocates_with_reasonable_capacity", func(t *testing.T) {
		input := []int{2, 1, 4, 3, 6, 5, 8}
		result := SliceFilter(func(v int) bool { return v%2 == 0 }, input)

		assert.Equal(t, []int{2, 4, 6, 8}, result)
		// Should allocate with capacity >= length of result
		assert.GreaterOrEqual(t, cap(result), len(result))
		// Should not over-allocate excessively
		assert.LessOrEqual(t, cap(result), len(input))
	})
}

func TestSliceSplitCapacity(t *testing.T) {
	t.Parallel()

	t.Run("all_true_returns_original_slice", func(t *testing.T) {
		input := []int{2, 4, 6, 8}
		trueSlice, falseSlice := SliceSplit(func(v int) bool { return v%2 == 0 }, input)

		assert.Equal(t, input, trueSlice)
		assert.Empty(t, falseSlice)
		assert.Equal(t, cap(input), cap(trueSlice))
	})

	t.Run("all_false_returns_original_slice", func(t *testing.T) {
		input := []int{1, 3, 5, 7}
		trueSlice, falseSlice := SliceSplit(func(v int) bool { return v%2 == 0 }, input)

		assert.Empty(t, trueSlice)
		assert.Equal(t, input, falseSlice)
		assert.Equal(t, cap(input), cap(falseSlice))
	})

	t.Run("split_allocates_appropriate_capacity", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6, 7, 8}
		trueSlice, falseSlice := SliceSplit(func(v int) bool { return v%2 == 0 }, input)

		assert.Equal(t, []int{2, 4, 6, 8}, trueSlice)
		assert.Equal(t, []int{1, 3, 5, 7}, falseSlice)

		// Both slices should have reasonable capacity
		assert.GreaterOrEqual(t, cap(trueSlice), len(trueSlice))
		assert.GreaterOrEqual(t, cap(falseSlice), len(falseSlice))

		// Capacity should not be excessive
		assert.LessOrEqual(t, cap(trueSlice), len(input))
		assert.LessOrEqual(t, cap(falseSlice), len(input))
	})

	t.Run("first_element_true_capacity_allocation", func(t *testing.T) {
		input := []int{2, 1, 4, 3, 6}
		trueSlice, falseSlice := SliceSplit(func(v int) bool { return v%2 == 0 }, input)

		assert.Equal(t, []int{2, 4, 6}, trueSlice)
		assert.Equal(t, []int{1, 3}, falseSlice)

		// Check capacity is reasonable for both slices
		assert.GreaterOrEqual(t, cap(trueSlice), len(trueSlice))
		assert.GreaterOrEqual(t, cap(falseSlice), len(falseSlice))
	})

	t.Run("first_element_false_capacity_allocation", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		trueSlice, falseSlice := SliceSplit(func(v int) bool { return v%2 == 0 }, input)

		assert.Equal(t, []int{2, 4}, trueSlice)
		assert.Equal(t, []int{1, 3, 5}, falseSlice)

		// Check capacity is reasonable for both slices
		assert.GreaterOrEqual(t, cap(trueSlice), len(trueSlice))
		assert.GreaterOrEqual(t, cap(falseSlice), len(falseSlice))
	})
}

func TestSliceSplitInPlaceCapacity(t *testing.T) {
	t.Parallel()

	t.Run("first_true_reuses_input_slice", func(t *testing.T) {
		input := []int{2, 1, 4, 3, 6}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlace(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Equal(t, []int{2, 4, 6}, trueSlice)
		assert.Equal(t, []int{1, 3}, falseSlice)

		// True slice should reuse original capacity
		assert.Equal(t, originalCap, cap(trueSlice))
		// False slice should have reasonable capacity
		assert.GreaterOrEqual(t, cap(falseSlice), len(falseSlice))
	})

	t.Run("first_false_reuses_input_slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlace(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Equal(t, []int{2, 4}, trueSlice)
		assert.Equal(t, []int{1, 3, 5}, falseSlice)

		// False slice should reuse original capacity
		assert.Equal(t, originalCap, cap(falseSlice))
		// True slice should have reasonable capacity
		assert.GreaterOrEqual(t, cap(trueSlice), len(trueSlice))
	})

	t.Run("all_true_reuses_input_capacity", func(t *testing.T) {
		input := []int{2, 4, 6, 8}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlace(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Equal(t, input, trueSlice)
		assert.Empty(t, falseSlice)
		assert.Equal(t, originalCap, cap(trueSlice))
	})

	t.Run("all_false_reuses_input_capacity", func(t *testing.T) {
		input := []int{1, 3, 5, 7}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlace(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Empty(t, trueSlice)
		assert.Equal(t, input, falseSlice)
		assert.Equal(t, originalCap, cap(falseSlice))
	})
}

func TestSliceSplitInPlaceUnstableCapacity(t *testing.T) {
	t.Parallel()

	t.Run("reuses_input_slice_capacity", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlaceUnstable(func(v int) bool { return v%2 == 0 }, inputCopy)

		// Both slices should use the original array
		totalLen := len(trueSlice) + len(falseSlice)
		assert.Equal(t, len(input), totalLen)

		// Combined capacity should equal original capacity
		if len(trueSlice) > 0 && len(falseSlice) > 0 {
			// Both slices share the same underlying array
			assert.LessOrEqual(t, cap(trueSlice), originalCap)
			assert.LessOrEqual(t, cap(falseSlice), originalCap)
		} else if len(trueSlice) > 0 {
			assert.Equal(t, originalCap, cap(trueSlice))
		} else if len(falseSlice) > 0 {
			assert.Equal(t, originalCap, cap(falseSlice))
		}
	})

	t.Run("all_true_preserves_capacity", func(t *testing.T) {
		input := []int{2, 4, 6, 8}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlaceUnstable(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Equal(t, originalCap, cap(trueSlice))
		assert.Empty(t, falseSlice)
	})

	t.Run("all_false_preserves_capacity", func(t *testing.T) {
		input := []int{1, 3, 5, 7}
		originalCap := cap(input)
		inputCopy := sliceDup(input)

		trueSlice, falseSlice :=
			SliceSplitInPlaceUnstable(func(v int) bool { return v%2 == 0 }, inputCopy)

		assert.Empty(t, trueSlice)
		assert.Equal(t, originalCap, cap(falseSlice))
	})
}

func TestSliceTransform(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		var input []int
		result := SliceTransform(func(i int) int { return i }, input)
		assert.Empty(t, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := SliceTransform(func(i int) int { return i }, []int{})
		assert.Empty(t, result)
	})

	t.Run("int_to_string", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := SliceTransform(func(i int) string { return strconv.Itoa(i) }, input)
		assert.Equal(t, []string{"1", "2", "3"}, result)
	})

	t.Run("float_to_int", func(t *testing.T) {
		input := []float64{1.1, 2.2}
		result := SliceTransform(func(f float64) int { return int(f) }, input)
		assert.Equal(t, []int{1, 2}, result)
	})

	t.Run("multiple_slices_concatenated", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{3, 4}
		slice3 := []int{5}
		result := SliceTransform(func(i int) string { return strconv.Itoa(i) }, slice1, slice2, slice3)
		assert.Equal(t, []string{"1", "2", "3", "4", "5"}, result)
	})

	t.Run("multiple_slices_with_empty", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{}
		slice3 := []int{3, 4}
		result := SliceTransform(func(i int) int { return i * 2 }, slice1, slice2, slice3)
		assert.Equal(t, []int{2, 4, 6, 8}, result)
	})

	t.Run("multiple_slices_all_empty", func(t *testing.T) {
		result := SliceTransform(func(i int) int { return i }, []int{}, []int{}, []int{})
		assert.Empty(t, result)
	})

	t.Run("multiple_slices_with_nil", func(t *testing.T) {
		slice1 := []int{1}
		var slice2 []int
		slice3 := []int{2, 3}
		result := SliceTransform(func(i int) string { return strconv.Itoa(i) }, slice1, slice2, slice3)
		assert.Equal(t, []string{"1", "2", "3"}, result)
	})

	t.Run("single_slice_maintains_compatibility", func(t *testing.T) {
		input := []int{10, 20, 30}
		result := SliceTransform(func(i int) int { return i / 10 }, input)
		assert.Equal(t, []int{1, 2, 3}, result)
	})
}

func TestSliceTransformErr(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		var input []int
		result, err := SliceTransformErr(func(i int) (int, error) { return i, nil }, input)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty", func(t *testing.T) {
		result, err := SliceTransformErr(func(i int) (int, error) { return i, nil }, []int{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("int_to_string", func(t *testing.T) {
		input := []int{1, 2, 3}
		result, err := SliceTransformErr(func(i int) (string, error) { return strconv.Itoa(i), nil }, input)
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, result)
	})

	t.Run("float_to_int", func(t *testing.T) {
		input := []float64{1.1, 2.2}
		result, err := SliceTransformErr(func(f float64) (int, error) { return int(f), nil }, input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, result)
	})

	t.Run("multiple_slices_concatenated", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{3, 4}
		slice3 := []int{5}
		result, err := SliceTransformErr(func(i int) (string, error) { return strconv.Itoa(i), nil }, slice1, slice2, slice3)
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3", "4", "5"}, result)
	})

	t.Run("multiple_slices_with_empty", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{}
		slice3 := []int{3, 4}
		result, err := SliceTransformErr(func(i int) (int, error) { return i * 2, nil }, slice1, slice2, slice3)
		require.NoError(t, err)
		assert.Equal(t, []int{2, 4, 6, 8}, result)
	})

	t.Run("multiple_slices_all_empty", func(t *testing.T) {
		result, err := SliceTransformErr(func(i int) (int, error) { return i, nil }, []int{}, []int{}, []int{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("multiple_slices_with_nil", func(t *testing.T) {
		slice1 := []int{1}
		var slice2 []int
		slice3 := []int{2, 3}
		result, err := SliceTransformErr(func(i int) (string, error) { return strconv.Itoa(i), nil }, slice1, slice2, slice3)
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, result)
	})

	t.Run("single_slice_maintains_compatibility", func(t *testing.T) {
		input := []int{10, 20, 30}
		result, err := SliceTransformErr(func(i int) (int, error) { return i / 10, nil }, input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("error on first element", func(t *testing.T) {
		result, err := SliceTransformErr(
			func(i int) (string, error) { return "", errors.New("transform error") },
			[]int{1, 2, 3},
		)
		require.Error(t, err)
		assert.Equal(t, "transform error", err.Error())
		assert.Empty(t, result)
	})

	t.Run("error in middle with partial results", func(t *testing.T) {
		result, err := SliceTransformErr(
			func(i int) (string, error) {
				if i == 3 {
					return "", errors.New("error on 3")
				}
				return strconv.Itoa(i * 10), nil
			},
			[]int{1, 2, 3, 4},
		)
		require.Error(t, err)
		assert.Equal(t, "error on 3", err.Error())
		assert.Equal(t, []string{"10", "20"}, result) // partial results before error
	})

	t.Run("error in multi-slice scenario", func(t *testing.T) {
		result, err := SliceTransformErr(
			func(i int) (string, error) {
				if i == 4 {
					return "", errors.New("error on 4")
				}
				return strconv.Itoa(i), nil
			},
			[]int{1, 2}, []int{3, 4, 5}, // error occurs in second slice
		)
		require.Error(t, err)
		assert.Equal(t, "error on 4", err.Error())
		assert.Equal(t, []string{"1", "2", "3"}, result) // results from first slice + partial from second
	})

	t.Run("error at end of first slice", func(t *testing.T) {
		result, err := SliceTransformErr(
			func(i int) (string, error) {
				if i == 2 {
					return "", errors.New("error on 2")
				}
				return strconv.Itoa(i * 100), nil
			},
			[]int{1, 2}, []int{3, 4}, // error at end of first slice
		)
		require.Error(t, err)
		assert.Equal(t, "error on 2", err.Error())
		assert.Equal(t, []string{"100"}, result) // partial result from first slice only
	})
}

var sliceToSetTests = []struct {
	name       string
	input      [][]int
	expectKeys []int
}{
	{
		name:  "nil",
		input: nil,
	},
	{
		name:  "empty",
		input: [][]int{},
	},
	{
		name:       "unique_values",
		input:      [][]int{{1, 2}, {3, 4}},
		expectKeys: []int{1, 2, 3, 4},
	},
	{
		name:       "middle_overlapping_values",
		input:      [][]int{{1, 2}, {2, 3}, {3, 4}},
		expectKeys: []int{1, 2, 3, 4},
	},
	{
		name:       "single_slice",
		input:      [][]int{{1, 2, 3}},
		expectKeys: []int{1, 2, 3},
	},
	{
		name:  "nested_empty_slices",
		input: [][]int{{}, {}},
	},
	{
		name:       "duplicate_values_in_slice",
		input:      [][]int{{1, 1, 2}, {2, 3, 3}},
		expectKeys: []int{1, 2, 3},
	},
	{
		name:  "single_empty_slice",
		input: [][]int{{}},
	},
	{
		name:       "all_same_values",
		input:      [][]int{{1, 1, 1}, {1, 1}, {1}},
		expectKeys: []int{1},
	},
	{
		name:       "large_input_with_duplicates",
		input:      [][]int{sliceLargeInput[:25], sliceLargeInput[20:]},
		expectKeys: sliceLargeInput[:51], // 0-50 unique values
	},
	{
		name:       "zero_values",
		input:      [][]int{{0}, {0, 1}, {1, 0}},
		expectKeys: []int{0, 1},
	},
	{
		name:       "negative_values",
		input:      [][]int{{-1, -2}, {-2, -3}, {-3, -4}},
		expectKeys: []int{-1, -2, -3, -4},
	},
}

func TestSliceToSet(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToSetTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToSet(tt.input...)

			assert.Len(t, got, len(tt.expectKeys))

			for _, key := range tt.expectKeys {
				_, ok := got[key]
				assert.True(t, ok)
			}

			for key := range got {
				assert.Contains(t, tt.expectKeys, key)
			}
		})
	}

	t.Run("string", func(t *testing.T) {
		result := SliceToSet([][]string{{"a", "b"}, {"b", "c"}, {"c", "d"}}...)

		expected := map[string]struct{}{
			"a": {},
			"b": {},
			"c": {},
			"d": {},
		}
		assert.Len(t, result, 4)
		assert.Equal(t, expected, result)
	})
}

var sliceToSetByTests = []struct {
	name        string
	inputSlices [][]int
	conversion  func(int) string
	expectKeys  []string
}{
	{
		name:        "nil_input",
		inputSlices: nil,
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{},
	},
	{
		name:        "empty_input",
		inputSlices: [][]int{},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{},
	},
	{
		name:        "single_slice",
		inputSlices: [][]int{{1, 2, 3}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"1", "2", "3"},
	},
	{
		name:        "multiple_slices_unique_values",
		inputSlices: [][]int{{1, 2}, {3, 4}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"1", "2", "3", "4"},
	},
	{
		name:        "multiple_slices_overlapping_values",
		inputSlices: [][]int{{1, 2}, {2, 3}, {3, 4}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"1", "2", "3", "4"},
	},
	{
		name:        "duplicate_values_in_slice",
		inputSlices: [][]int{{1, 1, 2}, {2, 3, 3}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"1", "2", "3"},
	},
	{
		name:        "empty_slices_mixed",
		inputSlices: [][]int{{}, {1, 2}, {}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"1", "2"},
	},
	{
		name:        "all_same_values",
		inputSlices: [][]int{{5, 5, 5}, {5, 5}, {5}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"5"},
	},
	{
		name:        "zero_values",
		inputSlices: [][]int{{0}, {0, 1}, {1, 0}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"0", "1"},
	},
	{
		name:        "negative_values",
		inputSlices: [][]int{{-1, -2}, {-2, -3}, {-3, -4}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys:  []string{"-1", "-2", "-3", "-4"},
	},
	{
		name:        "transform_to_same_key",
		inputSlices: [][]int{{1, 11, 21}, {31, 41, 51}},
		conversion:  func(i int) string { return strconv.Itoa(i % 10) }, // all mod 10 = 1
		expectKeys:  []string{"1"},
	},
	{
		name:        "large_input_with_duplicates",
		inputSlices: [][]int{sliceLargeInput[:25], sliceLargeInput[20:]},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectKeys: func() []string {
			keys := make([]string, 51) // 0-50 unique values
			for i := 0; i <= 50; i++ {
				keys[i] = strconv.Itoa(i)
			}
			return keys
		}(),
	},
}

func TestSliceToSetBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToSetByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToSetBy(tt.conversion, tt.inputSlices...)

			assert.Len(t, got, len(tt.expectKeys))

			for _, key := range tt.expectKeys {
				_, ok := got[key]
				assert.True(t, ok)
			}

			for key := range got {
				assert.Contains(t, tt.expectKeys, key, "unexpected key %q found", key)
			}
		})
	}

	t.Run("int_to_length_string", func(t *testing.T) {
		slice1 := []string{"a", "bb", "ccc"}
		slice2 := []string{"dd", "eeeee"}
		result := SliceToSetBy(func(s string) int { return len(s) }, slice1, slice2)

		expected := map[int]struct{}{
			1: {}, // "a"
			2: {}, // "bb", "dd"
			3: {}, // "ccc"
			5: {}, // "eeeee"
		}
		assert.Equal(t, expected, result)
	})

	t.Run("comparable_struct_keys", func(t *testing.T) {
		type Point struct{ X, Y int }
		points1 := []Point{{1, 2}, {3, 4}}
		points2 := []Point{{1, 2}, {5, 6}} // duplicate {1, 2}

		result := SliceToSetBy(func(p Point) Point { return p }, points1, points2)

		expected := map[Point]struct{}{
			{1, 2}: {},
			{3, 4}: {},
			{5, 6}: {},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("complex_transformation", func(t *testing.T) {
		nums1 := []int{10, 20, 30}
		nums2 := []int{15, 25, 35}

		// Transform to "even" or "odd" based on last digit
		result := SliceToSetBy(func(n int) string {
			if n%10%2 == 0 {
				return "even_ending"
			}
			return "odd_ending"
		}, nums1, nums2)

		expected := map[string]struct{}{
			"even_ending": {}, // 10, 20, 30
			"odd_ending":  {}, // 15, 25, 35
		}
		assert.Equal(t, expected, result)
	})
}

var sliceToCountsTests = []struct {
	name         string
	input        [][]int
	expectCounts map[int]int
}{
	{
		name:         "nil",
		input:        nil,
		expectCounts: map[int]int{},
	},
	{
		name:         "empty",
		input:        [][]int{},
		expectCounts: map[int]int{},
	},
	{
		name:         "unique_values",
		input:        [][]int{{1, 2}, {3, 4}},
		expectCounts: map[int]int{1: 1, 2: 1, 3: 1, 4: 1},
	},
	{
		name:         "overlapping_values",
		input:        [][]int{{1, 2}, {2, 3}, {3, 4}},
		expectCounts: map[int]int{1: 1, 2: 2, 3: 2, 4: 1},
	},
	{
		name:         "single_slice",
		input:        [][]int{{1, 2, 3}},
		expectCounts: map[int]int{1: 1, 2: 1, 3: 1},
	},
	{
		name:         "nested_empty_slices",
		input:        [][]int{{}, {}},
		expectCounts: map[int]int{},
	},
	{
		name:         "duplicate_values_in_slice",
		input:        [][]int{{1, 1, 2}, {2, 3, 3}},
		expectCounts: map[int]int{1: 2, 2: 2, 3: 2},
	},
	{
		name:         "single_empty_slice",
		input:        [][]int{{}},
		expectCounts: map[int]int{},
	},
	{
		name:         "all_same_values",
		input:        [][]int{{1, 1, 1}, {1, 1}, {1}},
		expectCounts: map[int]int{1: 6},
	},
	{
		name:         "zero_values",
		input:        [][]int{{0}, {0, 1}, {1, 0}},
		expectCounts: map[int]int{0: 3, 1: 2},
	},
	{
		name:         "negative_values",
		input:        [][]int{{-1, -2}, {-2, -3}, {-3, -4}},
		expectCounts: map[int]int{-1: 1, -2: 2, -3: 2, -4: 1},
	},
	{
		name:         "large_mixed_counts",
		input:        [][]int{{1, 2, 3, 2, 1}, {3, 4, 5, 4, 3}},
		expectCounts: map[int]int{1: 2, 2: 2, 3: 3, 4: 2, 5: 1},
	},
}

var sliceToCountsByTests = []struct {
	name         string
	inputSlices  [][]int
	conversion   func(int) string
	expectCounts map[string]int
}{
	{
		name:         "nil_input",
		inputSlices:  nil,
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{},
	},
	{
		name:         "empty_input",
		inputSlices:  [][]int{},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{},
	},
	{
		name:         "single_slice",
		inputSlices:  [][]int{{1, 2, 3}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"1": 1, "2": 1, "3": 1},
	},
	{
		name:         "multiple_slices_unique_values",
		inputSlices:  [][]int{{1, 2}, {3, 4}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"1": 1, "2": 1, "3": 1, "4": 1},
	},
	{
		name:         "multiple_slices_overlapping_values",
		inputSlices:  [][]int{{1, 2}, {2, 3}, {3, 4}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"1": 1, "2": 2, "3": 2, "4": 1},
	},
	{
		name:         "duplicate_values_in_slice",
		inputSlices:  [][]int{{1, 1, 2}, {2, 3, 3}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"1": 2, "2": 2, "3": 2},
	},
	{
		name:         "empty_slices_mixed",
		inputSlices:  [][]int{{}, {1, 2}, {}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"1": 1, "2": 1},
	},
	{
		name:         "all_same_values",
		inputSlices:  [][]int{{5, 5, 5}, {5, 5}, {5}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"5": 6},
	},
	{
		name:         "zero_values",
		inputSlices:  [][]int{{0}, {0, 1}, {1, 0}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"0": 3, "1": 2},
	},
	{
		name:         "negative_values",
		inputSlices:  [][]int{{-1, -2}, {-2, -3}, {-3, -4}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectCounts: map[string]int{"-1": 1, "-2": 2, "-3": 2, "-4": 1},
	},
	{
		name:         "transform_to_same_key",
		inputSlices:  [][]int{{1, 11, 21}, {31, 41, 51}},
		conversion:   func(i int) string { return strconv.Itoa(i % 10) },
		expectCounts: map[string]int{"1": 6},
	},
	{
		name:         "mod_10_grouping",
		inputSlices:  [][]int{{10, 20, 11, 21}, {12, 22, 13, 23}},
		conversion:   func(i int) string { return strconv.Itoa(i % 10) },
		expectCounts: map[string]int{"0": 2, "1": 2, "2": 2, "3": 2},
	},
}

func TestSliceToCounts(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToCountsTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToCounts(tt.input...)

			assert.Len(t, got, len(tt.expectCounts))
			for key, expectedCount := range tt.expectCounts {
				actualCount, ok := got[key]
				assert.True(t, ok)
				assert.Equal(t, expectedCount, actualCount)
			}

			for key := range got {
				assert.Contains(t, tt.expectCounts, key, "unexpected key %v found", key)
			}
		})
	}

	t.Run("string_values", func(t *testing.T) {
		result := SliceToCounts([][]string{{"a", "b"}, {"b", "c"}, {"c", "d", "a"}}...)

		expected := map[string]int{
			"a": 2,
			"b": 2,
			"c": 2,
			"d": 1,
		}
		assert.Equal(t, expected, result)
	})
}

func TestSliceToCountsBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToCountsByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToCountsBy(tt.conversion, tt.inputSlices...)

			assert.Len(t, got, len(tt.expectCounts))
			for key, expectedCount := range tt.expectCounts {
				actualCount, ok := got[key]
				assert.True(t, ok)
				assert.Equal(t, expectedCount, actualCount)
			}

			for key := range got {
				assert.Contains(t, tt.expectCounts, key, "unexpected key %q found", key)
			}
		})
	}

	t.Run("int_to_length_string", func(t *testing.T) {
		slice1 := []string{"a", "bb", "ccc", "a"}
		slice2 := []string{"dd", "eeeee", "bb"}
		result := SliceToCountsBy(func(s string) int { return len(s) }, slice1, slice2)

		expected := map[int]int{
			1: 2, // "a", "a"
			2: 3, // "bb", "dd", "bb"
			3: 1, // "ccc"
			5: 1, // "eeeee"
		}
		assert.Equal(t, expected, result)
	})

	t.Run("comparable_struct_keys", func(t *testing.T) {
		type Point struct{ X, Y int }
		points1 := []Point{{1, 2}, {3, 4}, {1, 2}}
		points2 := []Point{{1, 2}, {5, 6}, {3, 4}}

		result := SliceToCountsBy(func(p Point) Point { return p }, points1, points2)

		expected := map[Point]int{
			{1, 2}: 3, // appears 3 times total
			{3, 4}: 2, // appears 2 times total
			{5, 6}: 1, // appears 1 time
		}
		assert.Equal(t, expected, result)
	})

	t.Run("complex_transformation", func(t *testing.T) {
		nums1 := []int{10, 20, 30, 15}
		nums2 := []int{25, 35, 40, 45}

		result := SliceToCountsBy(func(n int) string {
			if n%10%2 == 0 {
				return "even_ending"
			}
			return "odd_ending"
		}, nums1, nums2)

		expected := map[string]int{
			"even_ending": 4, // 10, 20, 30, 40
			"odd_ending":  4, // 15, 25, 35, 45
		}
		assert.Equal(t, expected, result)
	})
}

var sliceToIndexByTests = []struct {
	name        string
	inputSlices [][]int
	conversion  func(int) string
	expectIndex map[string]int
}{
	{
		name:        "nil_input",
		inputSlices: nil,
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{},
	},
	{
		name:        "empty_input",
		inputSlices: [][]int{},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{},
	},
	{
		name:        "single_slice",
		inputSlices: [][]int{{1, 2, 3}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"1": 1, "2": 2, "3": 3},
	},
	{
		name:        "multiple_slices_unique_keys",
		inputSlices: [][]int{{1, 2}, {3, 4}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"1": 1, "2": 2, "3": 3, "4": 4},
	},
	{
		name:        "duplicate_keys_last_wins",
		inputSlices: [][]int{{1, 2}, {2, 3}, {3, 4}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"1": 1, "2": 2, "3": 3, "4": 4}, // last occurrence wins
	},
	{
		name:        "empty_slices_mixed",
		inputSlices: [][]int{{}, {1, 2}, {}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"1": 1, "2": 2},
	},
	{
		name:        "overwrite_values",
		inputSlices: [][]int{{10, 20}, {30, 10}},
		conversion:  func(i int) string { return strconv.Itoa(i % 10) },
		expectIndex: map[string]int{"0": 10}, // "0" key maps to last 10 (from 30, then 10)
	},
	{
		name:        "zero_values",
		inputSlices: [][]int{{0, 1}, {2, 0}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"0": 0, "1": 1, "2": 2}, // last 0 wins
	},
	{
		name:        "negative_values",
		inputSlices: [][]int{{-1, -2}, {-3, -2}},
		conversion:  func(i int) string { return strconv.Itoa(i) },
		expectIndex: map[string]int{"-1": -1, "-2": -2, "-3": -3}, // last -2 wins
	},
}

var sliceToGroupsByTests = []struct {
	name         string
	inputSlices  [][]int
	conversion   func(int) string
	expectGroups map[string][]int
}{
	{
		name:         "nil_input",
		inputSlices:  nil,
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{},
	},
	{
		name:         "empty_input",
		inputSlices:  [][]int{},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{},
	},
	{
		name:         "single_slice",
		inputSlices:  [][]int{{1, 2, 3}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"1": {1}, "2": {2}, "3": {3}},
	},
	{
		name:         "multiple_slices_unique_keys",
		inputSlices:  [][]int{{1, 2}, {3, 4}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"1": {1}, "2": {2}, "3": {3}, "4": {4}},
	},
	{
		name:         "multiple_slices_overlapping_keys",
		inputSlices:  [][]int{{1, 2}, {2, 3}, {3, 4}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"1": {1}, "2": {2, 2}, "3": {3, 3}, "4": {4}},
	},
	{
		name:         "empty_slices_mixed",
		inputSlices:  [][]int{{}, {1, 2}, {}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"1": {1}, "2": {2}},
	},
	{
		name:         "group_by_mod_10",
		inputSlices:  [][]int{{10, 21, 32}, {13, 24, 30}},
		conversion:   func(i int) string { return strconv.Itoa(i % 10) },
		expectGroups: map[string][]int{"0": {10, 30}, "1": {21}, "2": {32}, "3": {13}, "4": {24}},
	},
	{
		name:         "zero_values",
		inputSlices:  [][]int{{0, 1}, {2, 0}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"0": {0, 0}, "1": {1}, "2": {2}},
	},
	{
		name:         "negative_values",
		inputSlices:  [][]int{{-1, -2}, {-3, -2}},
		conversion:   func(i int) string { return strconv.Itoa(i) },
		expectGroups: map[string][]int{"-1": {-1}, "-2": {-2, -2}, "-3": {-3}},
	},
	{
		name:        "group_by_sign",
		inputSlices: [][]int{{-5, -2, 3}, {7, -1, 0}},
		conversion: func(i int) string {
			if i < 0 {
				return "negative"
			} else if i > 0 {
				return "positive"
			}
			return "zero"
		},
		expectGroups: map[string][]int{"negative": {-5, -2, -1}, "positive": {3, 7}, "zero": {0}},
	},
}

func TestSliceToIndexBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToIndexByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToIndexBy(tt.conversion, tt.inputSlices...)

			assert.Len(t, got, len(tt.expectIndex))
			for key, expectedValue := range tt.expectIndex {
				actualValue, ok := got[key]
				assert.True(t, ok)
				assert.Equal(t, expectedValue, actualValue)
			}

			for key := range got {
				assert.Contains(t, tt.expectIndex, key, "unexpected key %q found", key)
			}
		})
	}

	t.Run("struct_indexing", func(t *testing.T) {
		type Person struct {
			ID   int
			Name string
		}
		people1 := []Person{{1, "Alice"}, {2, "Bob"}}
		people2 := []Person{{3, "Charlie"}, {1, "Alice_Updated"}} // ID 1 gets updated

		result := SliceToIndexBy(func(p Person) int { return p.ID }, people1, people2)

		expected := map[int]Person{
			1: {1, "Alice_Updated"}, // last wins
			2: {2, "Bob"},
			3: {3, "Charlie"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("string_to_length_index", func(t *testing.T) {
		slice1 := []string{"a", "bb", "ccc"}
		slice2 := []string{"dd", "eeeee", "f"} // "f" overwrites "a" (both len=1)

		result := SliceToIndexBy(func(s string) int { return len(s) }, slice1, slice2)

		expected := map[int]string{
			1: "f",     // last string with length 1
			2: "dd",    // last string with length 2
			3: "ccc",   // only string with length 3
			5: "eeeee", // only string with length 5
		}
		assert.Equal(t, expected, result)
	})
}

func TestSliceToGroupsBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToGroupsByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := SliceToGroupsBy(tt.conversion, tt.inputSlices...)

			assert.Len(t, got, len(tt.expectGroups))
			for key, expectedGroup := range tt.expectGroups {
				actualGroup, ok := got[key]
				assert.True(t, ok)
				assert.Equal(t, expectedGroup, actualGroup)
			}

			for key := range got {
				assert.Contains(t, tt.expectGroups, key, "unexpected key %q found", key)
			}
		})
	}

	t.Run("struct_grouping", func(t *testing.T) {
		type Person struct {
			Dept string
			Name string
		}
		people1 := []Person{{"eng", "Alice"}, {"sales", "Bob"}}
		people2 := []Person{{"eng", "Charlie"}, {"sales", "Dave"}}

		result := SliceToGroupsBy(func(p Person) string { return p.Dept }, people1, people2)

		expected := map[string][]Person{
			"eng":   {{Dept: "eng", Name: "Alice"}, {Dept: "eng", Name: "Charlie"}},
			"sales": {{Dept: "sales", Name: "Bob"}, {Dept: "sales", Name: "Dave"}},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("string_to_length_groups", func(t *testing.T) {
		slice1 := []string{"a", "bb", "ccc", "d"}
		slice2 := []string{"ee", "f", "gggg"}

		result := SliceToGroupsBy(func(s string) int { return len(s) }, slice1, slice2)

		expected := map[int][]string{
			1: {"a", "d", "f"},
			2: {"bb", "ee"},
			3: {"ccc"},
			4: {"gggg"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("complex_grouping_with_duplicates", func(t *testing.T) {
		nums1 := []int{10, 15, 20, 11}
		nums2 := []int{25, 30, 16}

		result := SliceToGroupsBy(func(n int) string {
			if n%10 < 5 {
				return "low_digit"
			}
			return "high_digit"
		}, nums1, nums2)

		expected := map[string][]int{
			"low_digit":  {10, 20, 11, 30}, // 10, 20, 11, 30 (digits 0, 0, 1, 0)
			"high_digit": {15, 25, 16},     // 15, 25, 16 (digits 5, 5, 6)
		}
		assert.Equal(t, expected, result)
	})
}

var sliceSetOperationTests = []struct {
	name               string
	sliceA             []int
	sliceB             []int
	expectedIntersect  []int
	expectedDifference []int
}{
	{
		name:               "nil",
		sliceA:             nil,
		sliceB:             nil,
		expectedIntersect:  nil,
		expectedDifference: nil,
	},
	{
		name:               "empty_both",
		sliceA:             []int{},
		sliceB:             []int{},
		expectedIntersect:  nil,
		expectedDifference: nil,
	},
	{
		name:               "empty_a",
		sliceA:             []int{},
		sliceB:             []int{1, 2, 3},
		expectedIntersect:  nil,
		expectedDifference: nil,
	},
	{
		name:               "empty_b",
		sliceA:             []int{1, 2, 3},
		sliceB:             []int{},
		expectedIntersect:  nil,
		expectedDifference: []int{1, 2, 3},
	},
	{
		name:               "no_overlap",
		sliceA:             []int{1, 2, 3},
		sliceB:             []int{4, 5, 6},
		expectedIntersect:  nil,
		expectedDifference: []int{1, 2, 3},
	},
	{
		name:               "complete_overlap",
		sliceA:             []int{1, 2, 3},
		sliceB:             []int{1, 2, 3},
		expectedIntersect:  []int{1, 2, 3},
		expectedDifference: nil,
	},
	{
		name:               "partial_overlap",
		sliceA:             []int{1, 2, 3, 4},
		sliceB:             []int{3, 4, 5, 6},
		expectedIntersect:  []int{3, 4},
		expectedDifference: []int{1, 2},
	},
	{
		name:               "single_overlap",
		sliceA:             []int{1, 2, 3},
		sliceB:             []int{3, 4, 5},
		expectedIntersect:  []int{3},
		expectedDifference: []int{1, 2},
	},
	{
		name:               "duplicates_in_a",
		sliceA:             []int{1, 2, 2, 3, 3, 3},
		sliceB:             []int{2, 3, 4},
		expectedIntersect:  []int{2, 3},
		expectedDifference: []int{1},
	},
	{
		name:               "duplicates_in_b",
		sliceA:             []int{1, 2, 3},
		sliceB:             []int{2, 2, 3, 3, 4, 4},
		expectedIntersect:  []int{2, 3},
		expectedDifference: []int{1},
	},
	{
		name:               "duplicates_in_both",
		sliceA:             []int{1, 1, 2, 2, 3, 3},
		sliceB:             []int{2, 2, 3, 3, 4, 4},
		expectedIntersect:  []int{2, 3},
		expectedDifference: []int{1},
	},
	{
		name:               "a_subset_of_b",
		sliceA:             []int{2, 4},
		sliceB:             []int{1, 2, 3, 4, 5},
		expectedIntersect:  []int{2, 4},
		expectedDifference: nil,
	},
	{
		name:               "b_subset_of_a",
		sliceA:             []int{1, 2, 3, 4, 5},
		sliceB:             []int{2, 4},
		expectedIntersect:  []int{2, 4},
		expectedDifference: []int{1, 3, 5},
	},
	{
		name:               "single_elements",
		sliceA:             []int{1},
		sliceB:             []int{1},
		expectedIntersect:  []int{1},
		expectedDifference: nil,
	},
	{
		name:               "single_elements_no_match",
		sliceA:             []int{1},
		sliceB:             []int{2},
		expectedIntersect:  nil,
		expectedDifference: []int{1},
	},
	{
		name:               "large_slices",
		sliceA:             []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19},
		sliceB:             []int{2, 4, 5, 6, 8, 9, 10, 12, 13, 14},
		expectedIntersect:  []int{5, 9, 13},
		expectedDifference: []int{1, 3, 7, 11, 15, 17, 19},
	},
}

func TestSliceIntersect(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceSetOperationTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := SliceIntersect(tt.sliceA, tt.sliceB)
			if tt.expectedIntersect == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedIntersect, result)
			}
		})
	}

	t.Run("string_values", func(t *testing.T) {
		a := []string{"apple", "banana", "cherry", "date"}
		b := []string{"banana", "cherry", "elderberry", "fig"}
		result := SliceIntersect(a, b)
		expected := []string{"banana", "cherry"}
		assert.Equal(t, expected, result)
	})

	t.Run("order_preserved_from_a", func(t *testing.T) {
		a := []int{3, 1, 4, 1, 5, 9}
		b := []int{9, 5, 1, 3}
		result := SliceIntersect(a, b)
		expected := []int{3, 1, 5, 9} // Order from a, duplicates removed
		assert.Equal(t, expected, result)
	})
}

func TestSliceDifference(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceSetOperationTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := SliceDifference(tt.sliceA, tt.sliceB)
			if tt.expectedDifference == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedDifference, result)
			}
		})
	}

	t.Run("string_values", func(t *testing.T) {
		a := []string{"apple", "banana", "cherry", "date"}
		b := []string{"banana", "cherry", "elderberry", "fig"}
		result := SliceDifference(a, b)
		expected := []string{"apple", "date"}
		assert.Equal(t, expected, result)
	})

	t.Run("order_preserved_from_a", func(t *testing.T) {
		a := []int{5, 3, 1, 4, 1, 5}
		b := []int{1, 3}
		result := SliceDifference(a, b)
		expected := []int{5, 4} // Order from a preserved, duplicates removed for consistency
		assert.Equal(t, expected, result)
	})

	t.Run("empty_b_returns_a", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{}
		result := SliceDifference(a, b)
		assert.Equal(t, a, result) // Should return original slice a
	})
}

func TestSliceTotalSize(t *testing.T) {
	t.Parallel()

	sliceTotalSizeTests := []struct {
		name   string
		slices [][]int
		expect int
	}{
		{
			name:   "nil",
			slices: nil,
		},
		{
			name:   "empty",
			slices: [][]int{},
		},
		{
			name:   "single_nil_slice",
			slices: [][]int{nil},
		},
		{
			name:   "single_empty_slice",
			slices: [][]int{{}},
		},
		{
			name:   "multiple_empty_slices",
			slices: [][]int{{}, {}, {}},
		},
		{
			name:   "single_slice",
			slices: [][]int{{1, 2, 3}},
			expect: 3,
		},
		{
			name:   "multiple_slices_same_size",
			slices: [][]int{{1, 2}, {3, 4}, {5, 6}},
			expect: 6,
		},
		{
			name:   "multiple_slices_different_sizes",
			slices: [][]int{{1}, {2, 3}, {4, 5, 6, 7}},
			expect: 7,
		},
		{
			name:   "mixed_empty_and_non_empty",
			slices: [][]int{{}, {1, 2}, {}, {3, 4, 5}},
			expect: 5,
		},
		{
			name:   "mixed_nil_and_non_nil",
			slices: [][]int{nil, {1, 2}, nil, {3, 4, 5}},
			expect: 5,
		},
		{
			name:   "large_single_slice",
			slices: [][]int{sliceLargeInput},
			expect: 101,
		},
		{
			name:   "multiple_large_slices",
			slices: [][]int{sliceLargeInput[:25], sliceLargeInput[25:50], sliceLargeInput[50:]},
			expect: 101,
		},
		{
			name:   "zero_length_but_non_nil",
			slices: [][]int{make([]int, 0, 10), make([]int, 0, 5)},
			expect: 0,
		},
	}

	for i, tt := range sliceTotalSizeTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			got := sliceTotalSize(tt.slices)
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestSliceIntoSet(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToSetTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[int]struct{})
			SliceIntoSet(result, tt.input...)

			assert.Len(t, result, len(tt.expectKeys))

			for _, key := range tt.expectKeys {
				_, ok := result[key]
				assert.True(t, ok)
			}

			for key := range result {
				assert.Contains(t, tt.expectKeys, key)
			}
		})
	}
}

func TestSliceIntoSetBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToSetByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[string]struct{})
			SliceIntoSetBy(result, tt.conversion, tt.inputSlices...)

			assert.Len(t, result, len(tt.expectKeys))

			for _, key := range tt.expectKeys {
				_, ok := result[key]
				assert.True(t, ok)
			}

			for key := range result {
				assert.Contains(t, tt.expectKeys, key, "unexpected key %q found", key)
			}
		})
	}
}

func TestSliceIntoCounts(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToCountsTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[int]int)
			SliceIntoCounts(result, tt.input...)

			assert.Len(t, result, len(tt.expectCounts))
			for key, expectedCount := range tt.expectCounts {
				actualCount, ok := result[key]
				assert.True(t, ok)
				assert.Equal(t, expectedCount, actualCount)
			}

			for key := range result {
				assert.Contains(t, tt.expectCounts, key)
			}
		})
	}
}

func TestSliceIntoCountsBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToCountsByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[string]int)
			SliceIntoCountsBy(result, tt.conversion, tt.inputSlices...)

			assert.Len(t, result, len(tt.expectCounts))
			for key, expectedCount := range tt.expectCounts {
				actualCount, ok := result[key]
				assert.True(t, ok)
				assert.Equal(t, expectedCount, actualCount)
			}

			for key := range result {
				assert.Contains(t, tt.expectCounts, key)
			}
		})
	}
}

func TestSliceIntoIndexBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToIndexByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[string]int)
			SliceIntoIndexBy(result, tt.conversion, tt.inputSlices...)

			assert.Len(t, result, len(tt.expectIndex))
			for key, expectedValue := range tt.expectIndex {
				actualValue, ok := result[key]
				assert.True(t, ok)
				assert.Equal(t, expectedValue, actualValue)
			}

			for key := range result {
				assert.Contains(t, tt.expectIndex, key)
			}
		})
	}
}

func TestSliceIntoGroupsBy(t *testing.T) {
	t.Parallel()

	for i, tt := range sliceToGroupsByTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[string][]int)
			SliceIntoGroupsBy(result, tt.conversion, tt.inputSlices...)

			assert.Len(t, result, len(tt.expectGroups))
			for key, expectedGroup := range tt.expectGroups {
				actualGroup, ok := result[key]
				assert.True(t, ok)
				assert.Equal(t, expectedGroup, actualGroup)
			}

			for key := range result {
				assert.Contains(t, tt.expectGroups, key)
			}
		})
	}
}

func TestSliceConcat(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		result := sliceConcat([][]int{}, false)
		assert.Nil(t, result)
	})

	t.Run("single slice", func(t *testing.T) {
		original := []int{1, 2, 3}
		input := [][]int{original}
		result := sliceConcat(input, false)

		// Should return same slice (no copy)
		assert.Same(t, &original[0], &result[0])
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("single empty slice", func(t *testing.T) {
		original := []int{}
		input := [][]int{original}
		result := sliceConcat(input, false)

		// Should return same slice reference
		assert.Equal(t, []int{}, result)
	})

	t.Run("multiple slices", func(t *testing.T) {
		input := [][]int{{1, 2}, {3, 4}, {5, 6}}
		result := sliceConcat(input, false)

		expected := []int{1, 2, 3, 4, 5, 6}
		assert.Equal(t, expected, result)

		// Should be a new allocation (copy made)
		assert.NotSame(t, &input[0][0], &result[0])
	})

	t.Run("multiple slices with empty", func(t *testing.T) {
		input := [][]int{{1, 2}, {}, {3, 4}}
		result := sliceConcat(input, false)

		expected := []int{1, 2, 3, 4}
		assert.Equal(t, expected, result)
	})

	t.Run("multiple empty slices", func(t *testing.T) {
		input := [][]int{{}, {}, {}}
		result := sliceConcat(input, false)

		expected := []int{}
		assert.Equal(t, expected, result)
	})

	t.Run("single slice with nil elements", func(t *testing.T) {
		input := [][]int{nil}
		result := sliceConcat(input, false)

		// Should return the nil slice directly
		assert.Nil(t, result)
	})

	t.Run("multiple slices with nil elements", func(t *testing.T) {
		input := [][]int{{1, 2}, nil, {3, 4}}
		result := sliceConcat(input, false)

		expected := []int{1, 2, 3, 4}
		assert.Equal(t, expected, result)
	})

	t.Run("in_place_true_sufficient_capacity", func(t *testing.T) {
		// First slice has enough capacity for all elements
		first := make([]int, 2, 10)
		first[0], first[1] = 1, 2
		input := [][]int{first, {3, 4}, {5}}
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3, 4, 5}
		assert.Equal(t, expected, result)

		// Should reuse first slice's underlying array
		assert.Same(t, &first[0], &result[0])
		assert.GreaterOrEqual(t, cap(result), 10) // Preserves original capacity
	})

	t.Run("in_place_true_insufficient_capacity", func(t *testing.T) {
		// First slice doesn't have enough capacity - should allocate new
		first := make([]int, 2, 3)
		first[0], first[1] = 1, 2
		input := [][]int{first, {3, 4}, {5}}
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3, 4, 5}
		assert.Equal(t, expected, result)

		// Should NOT reuse first slice's array due to insufficient capacity
		assert.NotSame(t, &first[0], &result[0])
	})

	t.Run("in_place_false_forces_allocation", func(t *testing.T) {
		// Even with sufficient capacity, inPlace=false should allocate
		first := make([]int, 2, 10)
		first[0], first[1] = 1, 2
		input := [][]int{first, {3, 4}}
		result := sliceConcat(input, false)

		expected := []int{1, 2, 3, 4}
		assert.Equal(t, expected, result)

		// Should NOT reuse first slice's array due to inPlace=false
		assert.NotSame(t, &first[0], &result[0])
	})

	t.Run("in_place_with_empty_slices", func(t *testing.T) {
		first := make([]int, 1, 5)
		first[0] = 1
		input := [][]int{first, {}, {2, 3}, {}}
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3}
		assert.Equal(t, expected, result)

		// Should reuse first slice's underlying array
		assert.Same(t, &first[0], &result[0])
	})

	t.Run("in_place_exact_capacity_fit", func(t *testing.T) {
		// Total size exactly matches first slice capacity
		first := make([]int, 1, 4)
		first[0] = 1
		input := [][]int{first, {2, 3, 4}}
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3, 4}
		assert.Equal(t, expected, result)

		// Should reuse first slice's underlying array
		assert.Same(t, &first[0], &result[0])
		assert.Equal(t, 4, cap(result)) // Exact capacity
	})

	t.Run("two_slices_both_empty_first_larger_capacity", func(t *testing.T) {
		first := make([]int, 0, 10) // Empty with larger capacity
		second := make([]int, 0, 5) // Empty with smaller capacity
		input := [][]int{first, second}
		result := sliceConcat(input, false)

		assert.Equal(t, []int{}, result)
		// Should return first slice since second doesn't have larger capacity
		assert.Equal(t, 10, cap(result))
	})

	t.Run("two_slices_both_empty_second_larger_capacity", func(t *testing.T) {
		first := make([]int, 0, 5)   // Empty with smaller capacity
		second := make([]int, 0, 10) // Empty with larger capacity
		input := [][]int{first, second}
		result := sliceConcat(input, false)

		assert.Equal(t, []int{}, result)
		// Should return second slice due to larger capacity
		assert.Equal(t, 10, cap(result))
	})

	t.Run("two_slices_first_empty_second_has_values", func(t *testing.T) {
		first := make([]int, 0, 3) // Empty
		second := []int{1, 2, 3}   // Has values
		input := [][]int{first, second}
		result := sliceConcat(input, false)

		expected := []int{1, 2, 3}
		assert.Equal(t, expected, result)
		// Should return second slice since it has values
		assert.Same(t, &second[0], &result[0])
	})

	t.Run("in_place_boundary_condition", func(t *testing.T) {
		// Test exact boundary condition where totalSize == cap(slices[0])
		first := make([]int, 1, 3) // exactly enough capacity
		first[0] = 1
		input := [][]int{first, {2, 3}} // total size = 3
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3}
		assert.Equal(t, expected, result)
		assert.Same(t, &first[0], &result[0]) // should reuse
		assert.Equal(t, 3, cap(result))       // should preserve exact capacity
	})

	t.Run("in_place_boundary_condition_insufficient", func(t *testing.T) {
		// Test boundary condition where totalSize > cap(slices[0]) by 1
		first := make([]int, 1, 3) // one less than needed capacity
		first[0] = 1
		input := [][]int{first, {2, 3, 4}} // total size = 4 > capacity 3
		result := sliceConcat(input, true)

		expected := []int{1, 2, 3, 4}
		assert.Equal(t, expected, result)
		assert.NotSame(t, &first[0], &result[0]) // should allocate new
		assert.GreaterOrEqual(t, cap(result), 4) // should have sufficient capacity
	})
}

func TestSlicePrepend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elem     int
		slices   [][]int
		expected []int
	}{
		{
			name:     "no_slices",
			elem:     1,
			slices:   nil,
			expected: []int{1},
		},
		{
			name:     "empty_slices",
			elem:     1,
			slices:   [][]int{},
			expected: []int{1},
		},
		{
			name:     "single_empty_slice",
			elem:     1,
			slices:   [][]int{{}},
			expected: []int{1},
		},
		{
			name:     "single_nil_slice",
			elem:     1,
			slices:   [][]int{nil},
			expected: []int{1},
		},
		{
			name:     "single_slice",
			elem:     0,
			slices:   [][]int{{1, 2, 3}},
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "multiple_slices",
			elem:     0,
			slices:   [][]int{{1, 2}, {3, 4}, {5}},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "multiple_slices_with_empty",
			elem:     99,
			slices:   [][]int{{1}, {}, {2, 3}},
			expected: []int{99, 1, 2, 3},
		},
		{
			name:     "mixed_nil_and_non_nil",
			elem:     42,
			slices:   [][]int{nil, {1, 2}, nil, {3}},
			expected: []int{42, 1, 2, 3},
		},
		{
			name:     "prepend_to_large_slice",
			elem:     -1,
			slices:   [][]int{sliceLargeInput},
			expected: append([]int{-1}, sliceLargeInput...),
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := SlicePrepend(tt.elem, tt.slices...)
			assert.Equal(t, tt.expected, result)
			// Verify exact allocation size
			assert.Len(t, result, len(tt.expected))
			assert.Equal(t, len(tt.expected), cap(result))
		})
	}

	// Test with different types
	t.Run("string_type", func(t *testing.T) {
		result := SlicePrepend("first", []string{"second", "third"}, []string{"fourth"})
		expected := []string{"first", "second", "third", "fourth"}
		assert.Equal(t, expected, result)
	})

	t.Run("struct_type", func(t *testing.T) {
		type person struct {
			name string
			age  int
		}
		p1 := person{"Alice", 30}
		p2 := person{"Bob", 25}
		p3 := person{"Charlie", 35}
		result := SlicePrepend(p1, []person{p2}, []person{p3})
		expected := []person{p1, p2, p3}
		assert.Equal(t, expected, result)
	})
}
