package semconv

// CmdStreamCommandStatus represents the Command execution status.
type CmdStreamCommandStatus string

const (
	// Ok indicates the Command completed successfully.
	Ok CmdStreamCommandStatus = "OK"

	// Failed indicates the Command failed to complete successfully.
	Failed CmdStreamCommandStatus = "FAILED"

	// Timeout indicates the Command timed out before completion.
	Timeout CmdStreamCommandStatus = "TIMEOUT"
)
