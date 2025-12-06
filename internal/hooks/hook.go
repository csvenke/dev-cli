package hooks

type Hook interface {
	ShouldRun() bool
	Run(projectName string) error
}

func Run(hooks []Hook, projectName string) {
	for _, hook := range hooks {
		if hook.ShouldRun() {
			_ = hook.Run(projectName)
		}
	}
}
