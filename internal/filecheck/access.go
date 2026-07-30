package filecheck

type Outcome string

const (
	Allowed Outcome = "allowed"
	Denied  Outcome = "denied"
	Unknown Outcome = "unknown"
)

type Result struct {
	Path   string
	Read   Outcome
	Write  Outcome
	Reason string
}

func Access(path string) Result {
	read, readReason := probe(path, true)
	write, writeReason := probe(path, false)
	reason := readReason
	if reason == "" {
		reason = writeReason
	}
	return Result{Path: path, Read: read, Write: write, Reason: reason}
}
