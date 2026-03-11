package preprocessor

// composeMappings composes two mappings:
// local maps stage output positions -> stage input positions,
// prev maps stage input positions -> original input positions.
// The returned mapping maps stage output positions -> original input positions.
func composeMappings(prev, local []int) []int {
	if len(local) == 0 {
		return nil
	}
	if len(prev) == 0 {
		out := make([]int, len(local))
		copy(out, local)
		return out
	}
	out := make([]int, len(local))
	for i, pos := range local {
		clamped := clampIndex(pos, len(prev))
		out[i] = prev[clamped]
	}
	return out
}

// inferStageMapping creates an approximate position mapping from transformed
// output back to the stage input. It keeps exact mappings for matching prefix
// and suffix runs and distributes unmatched middle segments proportionally.
func inferStageMapping(before, after string) []int {
	out := make([]int, len(after))
	if len(after) == 0 {
		return out
	}
	if len(before) == 0 {
		return out
	}

	leftBefore, leftAfter := 0, 0
	for leftBefore < len(before) && leftAfter < len(after) && before[leftBefore] == after[leftAfter] {
		out[leftAfter] = leftBefore
		leftBefore++
		leftAfter++
	}

	rightBefore, rightAfter := len(before)-1, len(after)-1
	for rightBefore >= leftBefore && rightAfter >= leftAfter && before[rightBefore] == after[rightAfter] {
		out[rightAfter] = rightBefore
		rightBefore--
		rightAfter--
	}

	if leftAfter > rightAfter {
		return out
	}

	oldStart, oldEnd := leftBefore, rightBefore
	if oldStart > oldEnd {
		anchor := clampIndex(oldStart, len(before))
		for i := leftAfter; i <= rightAfter; i++ {
			out[i] = anchor
		}
		return out
	}

	newSpan := rightAfter - leftAfter
	oldSpan := oldEnd - oldStart
	for i := leftAfter; i <= rightAfter; i++ {
		if newSpan <= 0 {
			out[i] = oldStart
			continue
		}
		rel := i - leftAfter
		out[i] = oldStart + (rel*oldSpan)/newSpan
	}

	return out
}

func clampIndex(pos, length int) int {
	if length <= 0 {
		return 0
	}
	if pos < 0 {
		return 0
	}
	if pos >= length {
		return length - 1
	}
	return pos
}

func identityMapping(length int) []int {
	if length <= 0 {
		return nil
	}
	mapping := make([]int, length)
	for i := range mapping {
		mapping[i] = i
	}
	return mapping
}
