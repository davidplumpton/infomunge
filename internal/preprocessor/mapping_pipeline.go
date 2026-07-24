package preprocessor

import "infomunge/internal/sourcemap"

// composeMappings composes two mappings:
// local maps stage output positions -> stage input positions,
// prev maps stage input positions -> original input positions.
// The returned mapping maps stage output positions -> original input positions.
func composeMappings(prev, local []int) []int {
	return sourcemap.ComposePositions(prev, local)
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
		anchor := sourcemap.ClampIndex(oldStart, len(before))
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

func identityMapping(length int) []int {
	return sourcemap.IdentityPositions(length)
}
