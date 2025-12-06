package selection

import "dev/internal/projects"

func MoveCursorUp(cursor int) int {
	if cursor > 0 {
		return cursor - 1
	}
	return cursor
}

func MoveCursorDown(cursor, maxIndex int) int {
	if cursor < maxIndex {
		return cursor + 1
	}
	return cursor
}

func DeleteLastChar(query string) string {
	if len(query) > 0 {
		return query[:len(query)-1]
	}
	return query
}

func AppendChar(query, char string) string {
	return query + char
}

func SelectProject(filtered []projects.Project, cursor int) string {
	if len(filtered) > 0 && cursor < len(filtered) {
		return filtered[cursor].Path
	}
	return ""
}
