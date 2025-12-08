package fuzzy

const (
	scoreMatch        = 1
	scoreConsecutive  = 2
	scoreWordBoundary = 3
)

func Score(query string, target string) int {
	if len(query) == 0 {
		return 1
	}
	if len(query) > len(target) {
		return 0
	}

	score := 0
	qi := 0
	prevMatch := -2

	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if len(target)-ti < len(query)-qi {
			return 0
		}

		if target[ti] == query[qi] {
			score += scoreMatch
			if ti == prevMatch+1 {
				score += scoreConsecutive
			}
			if ti == 0 || !isLetter(target[ti-1]) {
				score += scoreWordBoundary
			}
			prevMatch = ti
			qi++
		}
	}

	if qi == len(query) {
		return score
	}
	return 0
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
