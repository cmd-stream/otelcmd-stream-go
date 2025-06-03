package semconv

import "go.opentelemetry.io/otel/attribute"

const (
	// CmdStreamCommandSeqKey is the attribute Key conforming to the
	// "cmd-stream.command.seq" semantic conventions. It represents the sequence
	// number of the command in the stream.
	//
	// Type: int
	// RequirementLevel: Recommended
	// Stability: Experimental
	//
	// Examples: 1, 42, 999
	CmdStreamCommandSeqKey = attribute.Key("cmd-stream.command.seq")

	// CmdStreamCommandTypeKey is the attribute Key conforming to the
	// "cmd-stream.command.type" semantic conventions. It represents the type of
	// the command, such as "Create", "Update", or "Delete".
	//
	// Type: string
	// RequirementLevel: Recommended
	// Stability: Experimental
	//
	// Examples: "Create", "Update", "Delete"
	CmdStreamCommandTypeKey = attribute.Key("cmd-stream.command.type")

	// CmdStreamCommandSizeKey is the attribute Key conforming to the
	// "cmd-stream.command.size" semantic conventions. It represents the size of
	// the command payload in bytes.
	//
	// Type: int
	// RequirementLevel: Optional
	// Stability: Experimental
	//
	// Examples: 128, 4096, 1048576
	CmdStreamCommandSizeKey = attribute.Key("cmd-stream.command.size")

	// CmdStreamCommandStatusKey is the attribute Key conforming to the
	// "cmd-stream.command.status" semantic conventions. It represents the final
	// status of the command.
	//
	// Type: string (enum)
	// RequirementLevel: Recommended
	// Stability: Experimental
	//
	// Examples: "OK", "FAILED", "TIMEOUT"
	CmdStreamCommandStatusKey = attribute.Key("cmd-stream.command.status")
)

const (
	// CmdStreamResultSeqKey is the attribute Key conforming to the
	// "cmd-stream.result.seq" semantic conventions. It represents the sequence
	// number of the result in the stream.
	//
	// Type: int
	// RequirementLevel: Recommended
	// Stability: Experimental
	//
	// Examples: 1, 42, 999
	CmdStreamResultSeqKey = attribute.Key("cmd-stream.result.seq")

	// CmdStreamResultTypeKey is the attribute Key conforming to the
	// "cmd-stream.result.type" semantic conventions. It represents the type of
	// the result, such as "Ok", "Failed", or "Reply".
	//
	// Type: string
	// RequirementLevel: Recommended
	// Stability: Experimental
	//
	// Examples: "Ok", "Failed", "Reply"
	CmdStreamResultTypeKey = attribute.Key("cmd-stream.result.type")

	// CmdStreamResultSizeKey is the attribute Key conforming to the
	// "cmd-stream.result.size" semantic conventions. It represents the size of
	// the result payload in bytes.
	//
	// Type: int
	// RequirementLevel: Optional
	// Stability: Experimental
	//
	// Examples: 256, 8192, 2097152
	CmdStreamResultSizeKey = attribute.Key("cmd-stream.result.size")
)
