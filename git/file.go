package git

const (
	StagedStatus   = "staged"
	UnstagedStatus = "unstaged"
)

type File struct {
	name   string
	status string
}

func (f File) Name() string {
	return f.name
}

func (f File) Diff() ([]string, error) {
	if f.status == StagedStatus {
		return git("--no-pager", "diff", "--staged", "--unified=0", f.name)
	} else {
		return git("--no-pager", "diff", "--unified=0", f.name)
	}
}

func (f File) ToggleStatus() {
	if f.status == StagedStatus {
		f.status = UnstagedStatus
	} else {
		f.status = StagedStatus
	}
}

func (f File) Staged() bool {
	return f.status == StagedStatus
}
