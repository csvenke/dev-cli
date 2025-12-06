package selection

import (
	"testing"

	"dev/internal/projects"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestMoveCursorUp_DecrementsCursor(t *testing.T) {
	result := MoveCursorUp(5)

	if result != 4 {
		t.Errorf("expected 4, got %d", result)
	}
}

func TestMoveCursorUp_StopsAtZero(t *testing.T) {
	result := MoveCursorUp(0)

	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestMoveCursorDown_IncrementsCursor(t *testing.T) {
	result := MoveCursorDown(5, 10)

	if result != 6 {
		t.Errorf("expected 6, got %d", result)
	}
}

func TestMoveCursorDown_StopsAtMax(t *testing.T) {
	result := MoveCursorDown(10, 10)

	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}
}

func TestDeleteLastChar_RemovesCharacter(t *testing.T) {
	result := DeleteLastChar("hello")

	if result != "hell" {
		t.Errorf("expected 'hell', got %q", result)
	}
}

func TestDeleteLastChar_EmptyStringStaysEmpty(t *testing.T) {
	result := DeleteLastChar("")

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestAppendChar_AddsCharacter(t *testing.T) {
	result := AppendChar("hell", "o")

	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestAppendChar_WorksWithEmptyString(t *testing.T) {
	result := AppendChar("", "a")

	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
}

func TestSelectProject_ReturnsPath(t *testing.T) {
	filtered := []projects.Project{
		{Name: "project-a", Path: "/repos/project-a"},
		{Name: "project-b", Path: "/repos/project-b"},
	}

	result := SelectProject(filtered, 1)

	if result != "/repos/project-b" {
		t.Errorf("expected '/repos/project-b', got %q", result)
	}
}

func TestSelectProject_EmptyListReturnsEmpty(t *testing.T) {
	var filtered []projects.Project

	result := SelectProject(filtered, 0)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestSelectProject_OutOfBoundsReturnsEmpty(t *testing.T) {
	filtered := []projects.Project{
		{Name: "project-a", Path: "/repos/project-a"},
	}

	result := SelectProject(filtered, 5)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// Property-based tests

func TestMoveCursorUp_NeverNegative(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("result is never negative", prop.ForAll(
		func(cursor int) bool {
			// Only test with reasonable cursor values
			if cursor < 0 {
				cursor = 0
			}
			result := MoveCursorUp(cursor)
			return result >= 0
		},
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

func TestMoveCursorUp_NeverIncreases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("result is never greater than input", prop.ForAll(
		func(cursor int) bool {
			if cursor < 0 {
				return true
			}
			result := MoveCursorUp(cursor)
			return result <= cursor
		},
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

func TestMoveCursorDown_NeverExceedsMaxWhenStartingAtOrBelowMax(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("result never exceeds maxIndex when starting at or below max", prop.ForAll(
		func(cursor, maxIndex int) bool {
			if cursor < 0 || maxIndex < 0 || cursor > maxIndex {
				return true // Skip invalid inputs
			}
			result := MoveCursorDown(cursor, maxIndex)
			return result <= maxIndex
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

func TestMoveCursorDown_NeverDecreases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("result is never less than input", prop.ForAll(
		func(cursor, maxIndex int) bool {
			if cursor < 0 || maxIndex < 0 {
				return true
			}
			result := MoveCursorDown(cursor, maxIndex)
			return result >= cursor
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

func TestDeleteLastChar_LengthDecreases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("length decreases by 1 or stays 0", prop.ForAll(
		func(query string) bool {
			result := DeleteLastChar(query)
			if len(query) == 0 {
				return len(result) == 0
			}
			return len(result) == len(query)-1
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

func TestAppendChar_LengthIncreases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("length increases by char length", prop.ForAll(
		func(query, char string) bool {
			result := AppendChar(query, char)
			return len(result) == len(query)+len(char)
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

func TestSelectProject_ReturnsEmptyForInvalidCursor(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("returns empty for cursor >= length", prop.ForAll(
		func(numProjects, cursor int) bool {
			if numProjects < 0 {
				numProjects = 0
			}
			if cursor < 0 {
				return true
			}

			var filtered []projects.Project
			for i := 0; i < numProjects; i++ {
				filtered = append(filtered, projects.Project{
					Name: "project",
					Path: "/path",
				})
			}

			result := SelectProject(filtered, cursor)

			if cursor >= numProjects {
				return result == ""
			}
			return result == "/path"
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 20),
	))

	properties.TestingRun(t)
}
