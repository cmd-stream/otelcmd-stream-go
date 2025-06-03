package semconv

const (
	// CmdStreamClientCommandSize is the metric conforming to the
	// "cmd-stream.client.command.size" semantic conventions. It represents the
	// size of commands sent by the client.
	// Instrument: histogram
	// Unit: By
	// Stability: Experimental
	CmdStreamClientCommandSizeName        = "cmd-stream.client.command.size"
	CmdStreamClientCommandSizeUnit        = "By"
	CmdStreamClientCommandSizeDescription = "Size of client commands."

	// CmdStreamServerCommandSize is the metric conforming to the
	// "cmd-stream.server.command.size" semantic conventions. It represents the
	// size of commands received by the server.
	// Instrument: histogram
	// Unit: By
	// Stability: Experimental
	CmdStreamServerCommandSizeName        = "cmd-stream.server.command.size"
	CmdStreamServerCommandSizeUnit        = "By"
	CmdStreamServerCommandSizeDescription = "Size of server commands."

	// CmdStreamClientResultSize is the metric conforming to the
	// "cmd-stream.client.result.size" semantic conventions. It represents the
	// size of results received by the client.
	// Instrument: histogram
	// Unit: By
	// Stability: Experimental
	CmdStreamClientResultSizeName        = "cmd-stream.client.result.size"
	CmdStreamClientResultSizeUnit        = "By"
	CmdStreamClientResultSizeDescription = "Size of client results."

	// CmdStreamServerResultSize is the metric conforming to the
	// "cmd-stream.server.result.size" semantic conventions. It represents the
	// size of results sent by the server.
	// Instrument: histogram
	// Unit: By
	// Stability: Experimental
	CmdStreamServerResultSizeName        = "cmd-stream.server.result.size"
	CmdStreamServerResultSizeUnit        = "By"
	CmdStreamServerResultSizeDescription = "Size of server results."

	// CmdStreamClientCommandDuration is the metric conforming to the
	// "cmd-stream.client.command.duration" semantic conventions. It represents
	// the duration taken by the client to execute commands.
	// Instrument: histogram
	// Unit: s
	// Stability: Experimental
	CmdStreamClientCommandDurationName        = "cmd-stream.client.command.duration"
	CmdStreamClientCommandDurationUnit        = "s"
	CmdStreamClientCommandDurationDescription = "Duration of client commands."

	// CmdStreamServerCommandDuration is the metric conforming to the
	// "cmd-stream.server.command.duration" semantic conventions. It represents
	// the duration taken by the server to execute commands.
	// Instrument: histogram
	// Unit: s
	// Stability: Experimental
	CmdStreamServerCommandDurationName        = "cmd-stream.server.command.duration"
	CmdStreamServerCommandDurationUnit        = "s"
	CmdStreamServerCommandDurationDescription = "Duration of server commands."

	// CmdStreamClientResultDuration is the metric conforming to the
	// "cmd-stream.client.result.duration" semantic conventions. It represents the
	// duration taken by the client to receive results.
	// Instrument: histogram
	// Unit: s
	// Stability: Experimental
	CmdStreamClientResultDurationName        = "cmd-stream.client.result.duration"
	CmdStreamClientResultDurationUnit        = "s"
	CmdStreamClientResultDurationDescription = "Duration of client results."

	// CmdStreamServerResultDuration is the metric conforming to the
	// "cmd-stream.server.result.duration" semantic conventions. It represents the
	// duration taken by the server to send results.
	// Instrument: histogram
	// Unit: s
	// Stability: Experimental
	CmdStreamServerResultDurationName        = "cmd-stream.server.result.duration"
	CmdStreamServerResultDurationUnit        = "s"
	CmdStreamServerResultDurationDescription = "Duration of server results."

	// CmdStreamClientCommandCount is the metric conforming to the
	// "cmd-stream.client.command.count" semantic conventions. It represents the
	// number of commands sent by the client.
	// Instrument: counter
	// Unit: {command}
	// Stability: Experimental
	CmdStreamClientCommandCountName        = "cmd-stream.client.command.count"
	CmdStreamClientCommandCountUnit        = "{command}"
	CmdStreamClientCommandCountDescription = "Number of client commands."

	// CmdStreamServerCommandCount is the metric conforming to the
	// "cmd-stream.server.command.count" semantic conventions. It represents the
	// number of commands received by the server.
	// Instrument: counter
	// Unit: {command}
	// Stability: Experimental
	CmdStreamServerCommandCountName        = "cmd-stream.server.command.count"
	CmdStreamServerCommandCountUnit        = "{command}"
	CmdStreamServerCommandCountDescription = "Number of server commands."

	// CmdStreamClientResultCount is the metric conforming to the
	// "cmd-stream.client.result.count" semantic conventions. It represents the
	// number of results received by the client.
	// Instrument: counter
	// Unit: {result}
	// Stability: Experimental
	CmdStreamClientResultCountName        = "cmd-stream.client.result.count"
	CmdStreamClientResultCountUnit        = "{result}"
	CmdStreamClientResultCountDescription = "Number of client results."

	// CmdStreamServerResultCount is the metric conforming to the
	// "cmd-stream.server.result.count" semantic conventions. It represents the
	// number of results sent back by the server.
	// Instrument: counter
	// Unit: {result}
	// Stability: Experimental
	CmdStreamServerResultCountName        = "cmd-stream.server.result.count"
	CmdStreamServerResultCountUnit        = "{result}"
	CmdStreamServerResultCountDescription = "Number of server results."
)
