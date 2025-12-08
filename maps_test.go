package bulk

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mapInvertTests = []struct {
	name     string
	input    map[int]string
	expected map[string]int
}{
	{
		name:     "nil",
		input:    nil,
		expected: map[string]int{},
	},
	{
		name:     "empty",
		input:    map[int]string{},
		expected: map[string]int{},
	},
	{
		name:     "single_pair",
		input:    map[int]string{1: "one"},
		expected: map[string]int{"one": 1},
	},
	{
		name:     "multiple_pairs",
		input:    map[int]string{1: "one", 2: "two", 3: "three"},
		expected: map[string]int{"one": 1, "two": 2, "three": 3},
	},
	{
		name:     "zero_key",
		input:    map[int]string{0: "zero", 1: "one"},
		expected: map[string]int{"zero": 0, "one": 1},
	},
	{
		name:     "empty_string_value",
		input:    map[int]string{1: "", 2: "two"},
		expected: map[string]int{"": 1, "two": 2},
	},
	{
		name:     "simple_no_duplicates",
		input:    map[int]string{1: "first", 2: "second", 3: "third"},
		expected: map[string]int{"first": 1, "second": 2, "third": 3},
	},
	{
		name:     "numeric_string_values",
		input:    map[int]string{10: "10", 20: "20", 30: "30"},
		expected: map[string]int{"10": 10, "20": 20, "30": 30},
	},
	{
		name:     "negative_keys",
		input:    map[int]string{-1: "negative", 0: "zero", 1: "positive"},
		expected: map[string]int{"negative": -1, "zero": 0, "positive": 1},
	},
	{
		name:     "special_characters_values",
		input:    map[int]string{1: "hello world", 2: "test@example.com", 3: "special!@#$%"},
		expected: map[string]int{"hello world": 1, "test@example.com": 2, "special!@#$%": 3},
	},
	{
		name:     "unicode_values",
		input:    map[int]string{1: "café", 2: "naïve", 3: "résumé"},
		expected: map[string]int{"café": 1, "naïve": 2, "résumé": 3},
	},
	{
		name:     "long_string_values",
		input:    map[int]string{1: "this_is_a_very_long_string_value_to_test_behavior", 2: "short"},
		expected: map[string]int{"this_is_a_very_long_string_value_to_test_behavior": 1, "short": 2},
	},
	{
		name:     "large_numbers",
		input:    map[int]string{1000000: "million", 2000000: "two_million"},
		expected: map[string]int{"million": 1000000, "two_million": 2000000},
	},
}

var mapInvertDuplicateTests = []struct {
	name          string
	input         map[int]string
	expectedKeys  []string
	possibleValue map[string][]int // For duplicate values, any of these keys could win
}{
	{
		name:          "two_duplicate_values",
		input:         map[int]string{1: "same", 2: "same"},
		expectedKeys:  []string{"same"},
		possibleValue: map[string][]int{"same": {1, 2}},
	},
	{
		name:          "three_duplicate_values",
		input:         map[int]string{1: "dup", 2: "dup", 3: "dup"},
		expectedKeys:  []string{"dup"},
		possibleValue: map[string][]int{"dup": {1, 2, 3}},
	},
	{
		name:          "mixed_duplicates",
		input:         map[int]string{1: "a", 2: "b", 3: "a", 4: "c", 5: "b"},
		expectedKeys:  []string{"a", "b", "c"},
		possibleValue: map[string][]int{"a": {1, 3}, "b": {2, 5}, "c": {4}},
	},
	{
		name:          "all_same_value",
		input:         map[int]string{1: "all", 2: "all", 3: "all", 4: "all"},
		expectedKeys:  []string{"all"},
		possibleValue: map[string][]int{"all": {1, 2, 3, 4}},
	},
}

func TestMapInvert(t *testing.T) {
	t.Parallel()

	for i, tt := range mapInvertTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := MapInvert(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test duplicate value behavior separately since map iteration order is non-deterministic
	for i, tt := range mapInvertDuplicateTests {
		t.Run(strconv.Itoa(i)+"-duplicate-"+tt.name, func(t *testing.T) {
			result := MapInvert(tt.input)

			// Check that result has correct keys
			assert.Len(t, result, len(tt.expectedKeys))
			for _, key := range tt.expectedKeys {
				assert.Contains(t, result, key)
			}

			// Check that values are one of the possible values
			for key, value := range result {
				possibleValues, exists := tt.possibleValue[key]
				assert.True(t, exists)
				assert.Contains(t, possibleValues, value)
			}
		})
	}

	t.Run("string_to_int", func(t *testing.T) {
		input := map[string]int{"one": 1, "two": 2, "three": 3}
		result := MapInvert(input)
		expected := map[int]string{1: "one", 2: "two", 3: "three"}
		assert.Equal(t, expected, result)
	})

	t.Run("comparable_struct_keys", func(t *testing.T) {
		type Point struct{ X, Y int }
		input := map[Point]string{{1, 2}: "first", {3, 4}: "second"}
		result := MapInvert(input)
		expected := map[string]Point{"first": {1, 2}, "second": {3, 4}}
		assert.Equal(t, expected, result)
	})

	t.Run("comparable_struct_values", func(t *testing.T) {
		type Point struct{ X, Y int }
		input := map[string]Point{"first": {1, 2}, "second": {3, 4}}
		result := MapInvert(input)
		expected := map[Point]string{{1, 2}: "first", {3, 4}: "second"}
		assert.Equal(t, expected, result)
	})
}

func TestMapInvertInto(t *testing.T) {
	t.Parallel()

	for i, tt := range mapInvertTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := make(map[string]int)
			MapInvertInto(result, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test duplicate value behavior separately
	for i, tt := range mapInvertDuplicateTests {
		t.Run(strconv.Itoa(i)+"-duplicate-"+tt.name, func(t *testing.T) {
			result := make(map[string]int)
			MapInvertInto(result, tt.input)

			// Check that result has correct keys
			assert.Len(t, result, len(tt.expectedKeys))
			for _, key := range tt.expectedKeys {
				assert.Contains(t, result, key)
			}

			// Check that values are one of the possible values
			for key, value := range result {
				possibleValues, exists := tt.possibleValue[key]
				assert.True(t, exists)
				assert.Contains(t, possibleValues, value)
			}
		})
	}

	t.Run("append_to_existing_map", func(t *testing.T) {
		existing := map[string]int{"existing": 99}
		input := map[int]string{1: "one", 2: "two"}

		MapInvertInto(existing, input)

		expected := map[string]int{"existing": 99, "one": 1, "two": 2}
		assert.Equal(t, expected, existing)
	})

	t.Run("overwrite_existing_keys", func(t *testing.T) {
		existing := map[string]int{"one": 999, "existing": 99}
		input := map[int]string{1: "one", 2: "two"}

		MapInvertInto(existing, input)

		expected := map[string]int{"existing": 99, "one": 1, "two": 2}
		assert.Equal(t, expected, existing)
	})

	t.Run("nil_destination_panics", func(t *testing.T) {
		input := map[int]string{1: "one"}
		assert.Panics(t, func() {
			MapInvertInto(nil, input)
		})
	})

	t.Run("multiple_calls_accumulate", func(t *testing.T) {
		result := make(map[string]int)

		MapInvertInto(result, map[int]string{1: "one"})
		MapInvertInto(result, map[int]string{2: "two"})
		MapInvertInto(result, map[int]string{3: "three"})

		expected := map[string]int{"one": 1, "two": 2, "three": 3}
		assert.Equal(t, expected, result)
	})

	t.Run("empty_input_no_change", func(t *testing.T) {
		existing := map[string]int{"existing": 99}
		input := map[int]string{}

		MapInvertInto(existing, input)

		expected := map[string]int{"existing": 99}
		assert.Equal(t, expected, existing)
	})

	t.Run("nil_input_no_change", func(t *testing.T) {
		existing := map[string]int{"existing": 99}
		var input map[int]string

		MapInvertInto(existing, input)

		expected := map[string]int{"existing": 99}
		assert.Equal(t, expected, existing)
	})
}

var mapKeyValueSliceTests = []struct {
	name           string
	input          map[int]string
	expectedKeys   []int
	expectedValues []string
}{
	{
		name:           "nil",
		input:          nil,
		expectedKeys:   nil,
		expectedValues: nil,
	},
	{
		name:           "single_pair",
		input:          map[int]string{1: "one"},
		expectedKeys:   []int{1},
		expectedValues: []string{"one"},
	},
	{
		name:           "multiple_pairs",
		input:          map[int]string{1: "one", 2: "two", 3: "three"},
		expectedKeys:   []int{1, 2, 3},
		expectedValues: []string{"one", "two", "three"},
	},
	{
		name:           "zero_key",
		input:          map[int]string{0: "zero", 1: "one"},
		expectedKeys:   []int{0, 1},
		expectedValues: []string{"zero", "one"},
	},
	{
		name:           "negative_keys",
		input:          map[int]string{-1: "negative", 0: "zero", 1: "positive"},
		expectedKeys:   []int{-1, 0, 1},
		expectedValues: []string{"negative", "zero", "positive"},
	},
	{
		name:           "large_numbers",
		input:          map[int]string{1000000: "million", 2000000: "two_million"},
		expectedKeys:   []int{1000000, 2000000},
		expectedValues: []string{"million", "two_million"},
	},
	{
		name:           "empty_string_value",
		input:          map[int]string{1: "", 2: "two"},
		expectedKeys:   []int{1, 2},
		expectedValues: []string{"", "two"},
	},
	{
		name:           "unicode_values",
		input:          map[int]string{1: "café", 2: "naïve", 3: "résumé"},
		expectedKeys:   []int{1, 2, 3},
		expectedValues: []string{"café", "naïve", "résumé"},
	},
	{
		name:           "special_characters",
		input:          map[int]string{1: "hello world", 2: "test@example.com", 3: "special!@#$%"},
		expectedKeys:   []int{1, 2, 3},
		expectedValues: []string{"hello world", "test@example.com", "special!@#$%"},
	},
	{
		name:           "duplicate_values",
		input:          map[int]string{1: "same", 2: "same", 3: "same"},
		expectedKeys:   []int{1, 2, 3},
		expectedValues: []string{"same", "same", "same"},
	},
}

func TestMapKeysSlice(t *testing.T) {
	t.Parallel()

	for i, tt := range mapKeyValueSliceTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := MapKeysSlice(tt.input)
			assert.Len(t, result, len(tt.expectedKeys))
			assert.ElementsMatch(t, tt.expectedKeys, result)
			assert.Equal(t, len(result), cap(result))
		})
	}

	t.Run("empty", func(t *testing.T) {
		result := MapKeysSlice(map[string]int{})
		assert.Nil(t, result)
	})

	t.Run("struct", func(t *testing.T) {
		type Point struct{ X, Y int }
		input := map[Point]string{{1, 2}: "first", {3, 4}: "second"}
		result := MapKeysSlice(input)
		expected := []Point{{1, 2}, {3, 4}}
		assert.Len(t, result, len(expected))
		assert.ElementsMatch(t, expected, result)
	})
}

func TestMapValuesSlice(t *testing.T) {
	t.Parallel()

	for i, tt := range mapKeyValueSliceTests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			result := MapValuesSlice(tt.input)
			assert.Len(t, result, len(tt.expectedValues))
			assert.ElementsMatch(t, tt.expectedValues, result)
			assert.Equal(t, len(result), cap(result))
		})
	}

	t.Run("empty", func(t *testing.T) {
		result := MapValuesSlice(map[string]int{})
		assert.Nil(t, result)
	})

	t.Run("struct", func(t *testing.T) {
		type Point struct{ X, Y int }
		input := map[string]Point{"first": {1, 2}, "second": {3, 4}}
		result := MapValuesSlice(input)
		expected := []Point{{1, 2}, {3, 4}}
		assert.Len(t, result, len(expected))
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("slice", func(t *testing.T) {
		input := map[string][]int{"a": {1, 2}, "b": {3, 4}}
		result := MapValuesSlice(input)
		assert.Len(t, result, 2)
		// Can't use ElementsMatch directly on [][]int, so check contents
		found := make(map[string]bool)
		for _, v := range result {
			if len(v) == 2 && v[0] == 1 && v[1] == 2 {
				found["a"] = true
			} else if len(v) == 2 && v[0] == 3 && v[1] == 4 {
				found["b"] = true
			}
		}
		assert.True(t, found["a"])
		assert.True(t, found["b"])
	})
}
