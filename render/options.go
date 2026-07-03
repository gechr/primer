package render

// DiffOptions configures optional external diff rendering.
type DiffOptions struct {
	// DeltaBin is the path to the delta binary. When empty, DiffStyled will
	// try to find delta on PATH.
	DeltaBin string
	// RepoURL enables GitHub blob hyperlinks in delta output when set together
	// with CommitSHA, e.g. "https://github.com/owner/repo".
	RepoURL string
	// CommitSHA is the commit used for blob hyperlinks in delta output.
	CommitSHA string
}
