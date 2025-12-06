package projects

import "strings"

type Project struct {
	Name string
	Path string
}

func Filter(projects []Project, query string) []Project {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)
	var result []Project
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Path), query) {
			result = append(result, p)
		}
	}
	return result
}
