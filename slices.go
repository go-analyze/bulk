package bulk

// SliceFilter returns elements that pass the predicate function.
// May return the original slice if all elements pass (no allocation).
func SliceFilter[T any](predicate func(val T) bool, slices ...[]T) []T {
	switch len(slices) {
	case 0:
		return nil
	case 1:
		result, _ := singleSliceFilter(predicate, slices[0])
		return result
	}

	results := make([][]T, 0, len(slices))
	concatInPlace := true
	for i, slice := range slices {
		partResult, view := singleSliceFilter(predicate, slice)
		if len(results) == 1 && len(results[0]) == 0 {
			// Our head slice is empty, check if we should prefer this slice instead
			if len(partResult) > 0 ||
				// if no results this will be a view, but maybe we can retain a larger view for the user
				cap(partResult) > cap(results[0]) {
				results[0] = partResult
				concatInPlace = !view
				continue
			}
		}
		// head not replaced, check if we should append
		if i == 0 {
			concatInPlace = !view // if first slice uses a view we have to copy in concat
			results = append(results, partResult)
		} else if len(partResult) > 0 {
			results = append(results, partResult)
		}
	}
	return sliceConcat(results, concatInPlace)
}

// singleSliceFilter filters a single slice based on the predicate function.
// Returns (filteredElements, isView) where isView indicates if the result is a view of the original slice.
func singleSliceFilter[T any](predicate func(val T) bool, slice []T) ([]T, bool) {
	for falseIdx, val := range slice {
		if predicate(val) {
			continue // continue till first false is found
		}

		// build and return result
		if falseIdx == 0 {
			// Track state transitions for view optimization
			firstTrueIdx := -1

			// Find first true element
			for i := falseIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					firstTrueIdx = i
					break
				}
			}
			if firstTrueIdx == -1 {
				return slice[:0], true // No true elements found
			}

			// Check if all remaining elements are consecutive and true
			consecutiveEnd := firstTrueIdx
			nonConsecutiveStart := -1
			for i := firstTrueIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					consecutiveEnd = i
					continue
				}
				// Found a false element, break to check if consecutive section continues
				nonConsecutiveStart = i
				break
			}
			if nonConsecutiveStart < 0 {
				return slice[firstTrueIdx:], true // All elements from firstTrueIdx to end are true
			}

			// if any more trues, we have to allocate and append, otherwise return a view
			for j := nonConsecutiveStart + 1; j < len(slice); j++ {
				if predicate(slice[j]) {
					// Found another true after false, not consecutive - need to allocate
					// worst case size: (consecutiveEnd-firstTrueIdx+1) + 1 + (len(slice) - j - 1) (+1 -1 simplified out)
					result := make([]T, 0, (consecutiveEnd-firstTrueIdx+1)+capGuess(len(slice)-j))
					result = append(result, slice[firstTrueIdx:consecutiveEnd+1]...)
					result = append(result, slice[j])

					// Continue appending remaining true elements
					return SliceFilterInto(result, predicate, slice[j+1:]), false
				}
			}
			// No more true elements found, return consecutive view
			return slice[firstTrueIdx : consecutiveEnd+1], true
		} else { // Started true, now first false found
			// Find first true element after falseIdx
			secondTrueIdx := -1
			for i := falseIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					secondTrueIdx = i
					break
				}
			}
			if secondTrueIdx < 0 {
				return slice[:falseIdx], true // No true elements in suffix, return prefix only
			}

			// true+ -> false+ -> true - We must allocate at this point
			result := make([]T, 0, falseIdx+capGuess(len(slice)-secondTrueIdx))
			result = append(result, slice[:falseIdx]...)
			result = append(result, slice[secondTrueIdx])
			return SliceFilterInto(result, predicate, slice[secondTrueIdx+1:]), false
		}
	}
	return slice, true // all records tested to true
}

// sliceConcat is similar to slices.Concat (which is why API is not elevated),
// but differs in that if provided a single slice it's returned without a copy.
func sliceConcat[T any](slices [][]T, inPlace bool) []T {
	switch len(slices) {
	case 0:
		return nil
	case 1:
		return slices[0]
	case 2:
		if len(slices[0]) == 0 { // empty first slice, easy result choice
			if len(slices[1]) > 0 || cap(slices[1]) > cap(slices[0]) {
				return slices[1]
			}
			return slices[0] // both empty, 0 is largest capacity
		}
	}

	if totalSize := sliceTotalSize(slices); inPlace && totalSize <= cap(slices[0]) {
		result := slices[0]
		for i := 1; i < len(slices); i++ { // skip the first slice (set directly)
			result = append(result, slices[i]...)
		}
		return result
	} else { // allocate and copy
		result := make([]T, 0, totalSize)
		for _, slice := range slices {
			result = append(result, slice...)
		}
		return result
	}
}

// SliceFilterInto appends elements that pass the predicate function from the input slices into dest.
func SliceFilterInto[T any](dest []T, predicate func(T) bool, inputs ...[]T) []T {
	for _, input := range inputs {
		for _, val := range input {
			if predicate(val) {
				dest = append(dest, val)
			}
		}
	}
	return dest
}

// SliceFilterTransform returns transformed elements that pass the predicate function.
// Combines filtering and transformation in a single efficient pass.
func SliceFilterTransform[I any, R any](predicate func(I) bool, transform func(I) R, inputs ...[]I) []R {
	errTransform := func(i I) (R, error) {
		return transform(i), nil
	}

	result, _ := SliceFilterTransformErr(predicate, errTransform, inputs...)
	return result
}

// SliceFilterTransformErr returns transformed elements that pass the predicate function.
// Combines filtering and transformation in a single efficient pass.
// If the conversion function returns an error, appending will stop with the partial result returned and the original error.
func SliceFilterTransformErr[I any, R any](predicate func(I) bool, transform func(I) (R, error), inputs ...[]I) ([]R, error) {
	switch len(inputs) {
	case 0:
		return nil, nil
	case 1:
		return singleSliceFilterTransform(predicate, transform, inputs[0])
	}

	var results [][]R
	for i, slice := range inputs {
		// simpler because empty results will never have capacity, so we only have to care about parts with results
		partResult, err := singleSliceFilterTransform(predicate, transform, slice)
		if len(partResult) > 0 {
			if results == nil {
				results = make([][]R, 0, len(inputs)-i)
			}
			results = append(results, partResult)
		}
		if err != nil {
			return sliceConcat(results, true), err
		}
	}
	return sliceConcat(results, true), nil
}

// singleSliceFilterTransform filters and transforms a single slice based on the predicate and transform functions.
func singleSliceFilterTransform[I any, R any](predicate func(I) bool, transform func(I) (R, error), slice []I) ([]R, error) {
	for falseIdx, val := range slice {
		if predicate(val) {
			continue // continue till first false is found
		}

		// build and return result
		if falseIdx == 0 {
			// Track state transitions for view optimization
			firstTrueIdx := -1

			// Find first true element
			for i := falseIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					firstTrueIdx = i
					break
				}
			}
			if firstTrueIdx == -1 {
				return nil, nil // No true elements found, return empty slice
			}

			// Check if all remaining elements are consecutive and true
			consecutiveEnd := firstTrueIdx
			nonConsecutiveStart := -1
			for i := firstTrueIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					consecutiveEnd = i
					continue
				}
				// Found a false element, break to check if consecutive section continues
				nonConsecutiveStart = i
				break
			}
			if nonConsecutiveStart < 0 {
				// All elements from firstTrueIdx to end are true - transform consecutive section
				return SliceTransformErr(transform, slice[firstTrueIdx:])
			}

			// if any more trues, we have to allocate and append, otherwise return transformed slice
			for j := nonConsecutiveStart + 1; j < len(slice); j++ {
				if predicate(slice[j]) {
					// Found another true after false, not consecutive - need to allocate
					// worst case size: (consecutiveEnd-firstTrueIdx+1) + 1 + (len(slice) - j - 1) (+1 -1 simplified out)
					result := make([]R, 0, (consecutiveEnd-firstTrueIdx+1)+capGuess(len(slice)-j))
					for i := firstTrueIdx; i <= consecutiveEnd; i++ {
						val, err := transform(slice[i])
						if err != nil {
							return result, err
						}
						result = append(result, val)
					}
					val, err := transform(slice[j])
					if err != nil {
						return result, err
					}
					result = append(result, val)

					// Continue appending remaining true elements
					return SliceFilterTransformErrInto(result, predicate, transform, slice[j+1:])
				}
			}
			// No more true elements found, return consecutive transformed slice
			return SliceTransformErr(transform, slice[firstTrueIdx:consecutiveEnd+1])
		} else { // Started true, now first false found
			// Find first true element after falseIdx
			secondTrueIdx := -1
			for i := falseIdx + 1; i < len(slice); i++ {
				if predicate(slice[i]) {
					secondTrueIdx = i
					break
				}
			}
			if secondTrueIdx < 0 {
				// No true elements in suffix, return transformed prefix only
				return SliceTransformErr(transform, slice[:falseIdx])
			}

			// true+ -> false+ -> true - We must allocate at this point
			result := make([]R, 0, falseIdx+capGuess(len(slice)-secondTrueIdx))
			for i := 0; i < falseIdx; i++ {
				val, err := transform(slice[i])
				if err != nil {
					return result, err
				}
				result = append(result, val)
			}
			val, err := transform(slice[secondTrueIdx])
			if err != nil {
				return result, err
			}
			result = append(result, val)
			return SliceFilterTransformErrInto(result, predicate, transform, slice[secondTrueIdx+1:])
		}
	}
	// all records tested to true - transform all elements
	return SliceTransformErr(transform, slice)
}

// SliceFilterTransformInto appends transformed elements that pass the predicate function from the input slices into dest.
func SliceFilterTransformInto[I any, R any](dest []R, predicate func(I) bool, transform func(I) R, inputs ...[]I) []R {
	errTransform := func(i I) (R, error) { return transform(i), nil }
	result, _ := SliceFilterTransformErrInto(dest, predicate, errTransform, inputs...)
	return result
}

// SliceFilterTransformErrInto appends transformed elements that pass the predicate function from the input slices into dest.
// If the conversion function returns an error, operation stops and returns the partial result with the original error.
func SliceFilterTransformErrInto[I any, R any](dest []R, predicate func(I) bool, transform func(I) (R, error), inputs ...[]I) ([]R, error) {
	for _, input := range inputs {
		for _, val := range input {
			if predicate(val) {
				finalVal, err := transform(val)
				if err != nil {
					return dest, err
				}
				dest = append(dest, finalVal)
			}
		}
	}
	return dest, nil
}

// SliceFilterInPlace returns elements that pass the predicate function.
// The input slice is modified and must be discarded after calling.
func SliceFilterInPlace[T any](predicate func(val T) bool, slices ...[]T) []T {
	switch len(slices) {
	case 0:
		return nil
	case 1:
		return singleSliceFilterInPlace(predicate, slices[0])
	}

	results := make([][]T, 0, len(slices))
	firstCapacity := len(slices[0]) // initialize with the raw initial length to consider as capacity (NOT full cap range)
	for i, slice := range slices {
		partResult := singleSliceFilterInPlace(predicate, slice)
		if len(results) == 1 && len(results[0]) == 0 {
			// Our head slice is empty, check if we should prefer this slice instead
			if i == len(slices)-1 { // last record, pick the ideal result and break
				if len(partResult) > 0 || cap(partResult) > cap(results[0]) {
					results[0] = partResult
				}
				break
			}
			// if the current result offers more capacity retain it instead
			// otherwise just fall through and both will be retained
			// if the current is the only added slice sliceConcat will optimize away the empty head slice
			// if more slices are added (3+ total), then sliceConcat may use this slice to avoid allocations
			currCapacity := len(slice) - len(partResult)
			if firstCapacity < currCapacity {
				firstCapacity = currCapacity
				results[0] = partResult
				continue
			}
		} // if not continue above, fall to check below
		if i == 0 || len(partResult) > 0 {
			firstCapacity -= len(partResult)
			results = append(results, partResult)
		}
	}
	return sliceConcat(results, firstCapacity >= 0)
}

func singleSliceFilterInPlace[T any](predicate func(v T) bool, slice []T) []T {
	var n int
	for i := range slice {
		if predicate(slice[i]) {
			slice[n] = slice[i]
			n++
		}
	}
	return slice[:n]
}

// SliceSplit partitions elements based on the predicate function.
// Returns (trueElements, falseElements).
func SliceSplit[T any](predicate func(val T) bool, slices ...[]T) ([]T, []T) {
	switch len(slices) {
	case 0:
		return nil, nil
	case 1:
		tSlice, fSlice, _, _ := singleSliceSplit(predicate, slices[0])
		return tSlice, fSlice
	}

	trueResults, falseResults := make([][]T, 0, len(slices)), make([][]T, 0, len(slices))
	trueConcatInPlace, falseConcatInPlace := true, true
	for i, slice := range slices {
		tSlice, fSlice, tView, fView := singleSliceSplit(predicate, slice)
		if i == 0 { // ensure at least one true result
			trueResults = append(trueResults, tSlice)
			trueConcatInPlace = !tView // if the first result is a view, we can't concatenate in place
		} else if len(trueResults) == 1 && len(trueResults[0]) == 0 && // check for empty head to replace
			(len(tSlice) > 0 || // replace with actual results
				(!tView && !trueConcatInPlace) || // can be used to upgrade into an in place copy
				(!trueConcatInPlace && cap(tSlice) > cap(trueResults[0]))) { // won't downgrade and has more capacity
			trueResults[0] = tSlice
			trueConcatInPlace = !tView
		} else if len(tSlice) > 0 {
			trueResults = append(trueResults, tSlice)
		}
		if i == 0 { // ensure at least one false result
			falseResults = append(falseResults, fSlice)
			falseConcatInPlace = !fView // if the first result is a view, we can't concatenate in place
		} else if len(falseResults) == 1 && len(falseResults[0]) == 0 && // check for empty head to replace
			(len(fSlice) > 0 || // replace with actual results
				(!fView && !falseConcatInPlace) || // can be used to upgrade into an in place copy
				(!falseConcatInPlace && cap(fSlice) > cap(falseResults[0]))) { // won't downgrade and has more capacity
			falseResults[0] = fSlice
			falseConcatInPlace = !fView
		} else if len(fSlice) > 0 {
			falseResults = append(falseResults, fSlice)
		}
	}
	trueResult := sliceConcat(trueResults, trueConcatInPlace)
	falseResult := sliceConcat(falseResults, falseConcatInPlace)
	return trueResult, falseResult
}

// singleSliceSplit partitions a single slice based on the predicate function.
// Returns (trueElements, falseElements, trueIsView, falseIsView).
func singleSliceSplit[T any](predicate func(val T) bool, slice []T) ([]T, []T, bool, bool) {
	if len(slice) == 0 {
		return slice, nil, true, false
	}

	var splitIdx int
	first := predicate(slice[0])
	for splitIdx = 1; splitIdx < len(slice); splitIdx++ {
		if first != predicate(slice[splitIdx]) {
			break
		}
	}

	// If all are the same, return early
	if splitIdx == len(slice) {
		if first {
			return slice, nil, true, false
		} else {
			return nil, slice, false, true
		}
	}

	// Allocate slices and copy first segment
	remainingBuff := capGuess(len(slice) - splitIdx)
	var tSlice, fSlice []T
	if first {
		tSlice = append(make([]T, 0, splitIdx+remainingBuff-1), slice[:splitIdx]...)
		fSlice = append(make([]T, 0, remainingBuff), slice[splitIdx])
	} else {
		fSlice = append(make([]T, 0, splitIdx+remainingBuff-1), slice[:splitIdx]...)
		tSlice = append(make([]T, 0, remainingBuff), slice[splitIdx])
	}
	// Finish iterating appending remaining elements
	for i := splitIdx + 1; i < len(slice); i++ {
		if predicate(slice[i]) {
			tSlice = append(tSlice, slice[i])
		} else {
			fSlice = append(fSlice, slice[i])
		}
	}

	return tSlice, fSlice, false, false
}

// SliceSplitInPlace partitions elements based on the predicate function.
// The input slice is modified and must be discarded after calling. Resulting slices maintain original order.
// If order is not important, use SliceSplitInPlaceUnstable for better performance.
// Returns (trueElements, falseElements).
func SliceSplitInPlace[T any](predicate func(val T) bool, slice []T) ([]T, []T) {
	n := len(slice)
	if n == 0 {
		return slice, nil
	}

	match := predicate(slice[0])
	matchList := slice[:1]
	var otherList []T
	for i := 1; i < n; i++ {
		val := slice[i]
		if predicate(val) == match {
			matchList = append(matchList, val) // writes to earlier indices only; safe
		} else {
			if otherList == nil { // Allocate when we discover the split
				otherList = make([]T, 0, capGuess(n-i))
			}
			otherList = append(otherList, val)
		}
	}
	if match {
		return matchList, otherList
	} else {
		return otherList, matchList
	}
}

// SliceSplitInPlaceUnstable partitions elements based on the predicate function.
// The input slice is modified and must be discarded after calling.
// Element order may change from the original input.
// Returns (trueElements, falseElements).
func SliceSplitInPlaceUnstable[T any](predicate func(val T) bool, slice []T) ([]T, []T) {
	if len(slice) == 0 {
		return slice, nil
	}

	i, j := 0, len(slice)-1
	for {
		for i <= j && predicate(slice[i]) {
			i++
		}
		for i <= j && !predicate(slice[j]) {
			j--
		}
		if i >= j {
			break
		}
		slice[i], slice[j] = slice[j], slice[i]
		i++
		j--
	}

	switch {
	case i == 0:
		return nil, slice
	case i == len(slice):
		return slice, nil
	default:
		return slice[:i], slice[i:]
	}
}

// SliceTransform converts each element using the conversion function.
func SliceTransform[I any, R any](conversion func(I) R, inputs ...[]I) []R {
	result := make([]R, 0, sliceTotalSize(inputs))
	return SliceFilterTransformInto(result, func(_ I) bool { return true }, conversion, inputs...)
}

// SliceTransformErr converts each element using the conversion function.
// If the conversion function returns an error, operation stops and returns the partial result with the original error.
func SliceTransformErr[I any, R any](conversion func(I) (R, error), inputs ...[]I) ([]R, error) {
	result := make([]R, 0, sliceTotalSize(inputs))
	return SliceFilterTransformErrInto(result, func(_ I) bool { return true }, conversion, inputs...)
}

// SliceToSet accepts slices of comparable types and returns a map with elements as keys.
// This provides a deduplicated union of slices and enables fast lookups using the returned map.
func SliceToSet[T comparable](slices ...[]T) map[T]struct{} {
	result := make(map[T]struct{}, sliceTotalSize(slices))
	SliceIntoSet(result, slices...)
	return result
}

// SliceIntoSet adds elements from slices to an existing set map.
func SliceIntoSet[T comparable](m map[T]struct{}, slices ...[]T) {
	for _, slice := range slices {
		for _, value := range slice {
			m[value] = struct{}{}
		}
	}
}

// SliceToSetBy transforms slice elements using keyfunc and stores results as keys in a map.
// This combines transformation and set creation in a single operation.
func SliceToSetBy[I any, R comparable](keyfunc func(I) R, slices ...[]I) map[R]struct{} {
	result := make(map[R]struct{}, sliceTotalSize(slices))
	SliceIntoSetBy(result, keyfunc, slices...)
	return result
}

// SliceIntoSetBy transforms slice elements using keyfunc and adds results as keys to an existing set map.
func SliceIntoSetBy[I any, R comparable](m map[R]struct{}, keyfunc func(I) R, slices ...[]I) {
	for _, slice := range slices {
		for _, inputVal := range slice {
			m[keyfunc(inputVal)] = struct{}{}
		}
	}
}

// SliceToCounts accepts slices of comparable types and returns a map with elements as keys and their counts as values.
func SliceToCounts[T comparable](slices ...[]T) map[T]int {
	result := make(map[T]int, sliceTotalSize(slices))
	SliceIntoCounts(result, slices...)
	return result
}

// SliceIntoCounts adds element counts from slices to an existing count map.
func SliceIntoCounts[T comparable](m map[T]int, slices ...[]T) {
	for _, slice := range slices {
		for _, value := range slice {
			m[value]++
		}
	}
}

// SliceToCountsBy transforms slice elements using keyfunc and returns a map with generated keys and their counts.
func SliceToCountsBy[T any, K comparable](keyfunc func(T) K, slices ...[]T) map[K]int {
	result := make(map[K]int)
	SliceIntoCountsBy(result, keyfunc, slices...)
	return result
}

// SliceIntoCountsBy transforms slice elements using keyfunc and adds counts to an existing count map.
func SliceIntoCountsBy[T any, K comparable](m map[K]int, keyfunc func(T) K, slices ...[]T) {
	for _, slice := range slices {
		for _, value := range slice {
			m[keyfunc(value)]++
		}
	}
}

// SliceToIndexBy creates an index map using keyfunc to generate keys from slice elements.
// Keys should be unique. Duplicate keys will overwrite earlier values.
func SliceToIndexBy[T any, K comparable](keyfunc func(T) K, slices ...[]T) map[K]T {
	result := make(map[K]T, sliceTotalSize(slices))
	SliceIntoIndexBy(result, keyfunc, slices...)
	return result
}

// SliceIntoIndexBy adds elements to an existing index map using keyfunc to generate keys.
// Keys should be unique. Duplicate keys will overwrite earlier values.
func SliceIntoIndexBy[T any, K comparable](m map[K]T, keyfunc func(T) K, slices ...[]T) {
	for _, slice := range slices {
		for _, value := range slice {
			key := keyfunc(value)
			m[key] = value
		}
	}
}

// SliceToGroupsBy groups slice elements by keys generated using keyfunc.
func SliceToGroupsBy[T any, K comparable](keyfunc func(T) K, slices ...[]T) map[K][]T {
	result := make(map[K][]T)
	SliceIntoGroupsBy(result, keyfunc, slices...)
	return result
}

// SliceIntoGroupsBy adds elements to existing groups using keyfunc to generate keys.
func SliceIntoGroupsBy[T any, K comparable](m map[K][]T, keyfunc func(T) K, slices ...[]T) {
	for _, slice := range slices {
		for _, value := range slice {
			key := keyfunc(value)
			m[key] = append(m[key], value)
		}
	}
}

// SliceIntersect returns elements that exist in both slices, preserving order from the first slice.
func SliceIntersect[T comparable](a, b []T) []T {
	if len(a) == 0 {
		return a
	} else if len(b) == 0 {
		return b
	}

	maxCount := len(a)
	if len(b) < maxCount {
		maxCount = len(b)
	}

	// Collect intersection, preserving order from slice a
	bLookup := SliceToSet(b)
	var result []T
	var seen map[T]struct{}
	for aIdx, val := range a {
		if _, exists := bLookup[val]; exists {
			if _, duplicate := seen[val]; !duplicate {
				if result == nil { // allocate based on potential remaining
					if aMax := len(a) - aIdx; aMax < maxCount {
						maxCount = aMax // conditional because b may still have been the min
					}
					seen = make(map[T]struct{}, maxCount)
					result = make([]T, 0, capGuess(maxCount))
				}
				seen[val] = struct{}{}
				result = append(result, val)
			}
		}
	}
	return result
}

// SliceDifference returns elements that exist in the first slice but not in the second, preserving order from the first slice.
func SliceDifference[T comparable](a, b []T) []T {
	if len(a) == 0 {
		return a
	}

	// Collect elements from a that are not in b, with deduplication
	exclude := SliceToSet(b)
	var result []T
	var seen map[T]struct{}
	for aIdx, val := range a {
		if _, exists := exclude[val]; !exists {
			if _, duplicate := seen[val]; !duplicate {
				if result == nil { // allocate based on potential remaining
					seen = make(map[T]struct{}, len(a)-aIdx)
					result = make([]T, 0, capGuess(len(a)-aIdx))
				}
				seen[val] = struct{}{}
				result = append(result, val)
			}
		}
	}
	return result
}

// SlicePrepend creates a new slice with elm at the front followed by elements from existing slices.
// Performs a single allocation sized exactly for the result, avoiding multiple allocations
// that would occur with append([]T{elm}, existing...).
func SlicePrepend[T any](elm T, existing ...[]T) []T {
	result := make([]T, 1, sliceTotalSize(existing)+1)
	result[0] = elm
	for _, slice := range existing {
		result = append(result, slice...)
	}
	return result
}

// capGuess estimates allocation size, reducing large allocations in case of significant filtering.
func capGuess(remaining int) int {
	if remaining > 2048 {
		return remaining / 2
	}
	return remaining
}

func sliceTotalSize[T any](slices [][]T) int {
	var size int
	for _, slice := range slices {
		size += len(slice)
	}
	return capGuess(size)
}
