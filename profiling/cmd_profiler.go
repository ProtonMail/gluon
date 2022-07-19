package profiling

const (
	CmdTypeSelect      = 0
	CmdTypeCreate      = 1
	CmdTypeDelete      = 2
	CmdTypeRename      = 3
	CmdTypeSubscribe   = 4
	CmdTypeUnsubscribe = 5
	CmdTypeList        = 6
	CmdTypeLSub        = 7
	CmdTypeStatus      = 8
	CmdTypeAppend      = 9
	CmdTypeCheck       = 10
	CmdTypeClose       = 11
	CmdTypeExpunge     = 12
	CmdTypeSearch      = 13
	CmdTypeFetch       = 14
	CmdTypeStore       = 15
	CmdTypeCopy        = 16
	CmdTypeUID         = 17
	CmdTypeNoop        = 18
	CmdTypeIdle        = 19
	CmdTypeMove        = 20
	CmdTypeID          = 21
	CmdTypeLogout      = 22
	CmdTypeUnselect    = 23
	CmdTypeLogin       = 24
	CmdTypeExamine     = 25
	CmdTypeTotal       = 26
)

func CmdTypeToString(cmdType int) string {
	switch cmdType {
	case CmdTypeSelect:
		return "SELECT "
	case CmdTypeCreate:
		return "CREATE "
	case CmdTypeDelete:
		return "DELETE "
	case CmdTypeRename:
		return "RENAME "
	case CmdTypeSubscribe:
		return "SUB    "
	case CmdTypeUnsubscribe:
		return "USUB   "
	case CmdTypeList:
		return "LIST   "
	case CmdTypeLSub:
		return "LSUB   "
	case CmdTypeStatus:
		return "STATUS "
	case CmdTypeAppend:
		return "APPEND "
	case CmdTypeCheck:
		return "CHECK  "
	case CmdTypeClose:
		return "CLOSE  "
	case CmdTypeExpunge:
		return "EXPUNGE"
	case CmdTypeSearch:
		return "SEARCH "
	case CmdTypeFetch:
		return "FETCH  "
	case CmdTypeStore:
		return "STORE  "
	case CmdTypeCopy:
		return "COPY   "
	case CmdTypeUID:
		return "UID    "
	case CmdTypeNoop:
		return "NOOP   "
	case CmdTypeIdle:
		return "IDLE   "
	case CmdTypeMove:
		return "MOVE   "
	case CmdTypeID:
		return "ID     "
	case CmdTypeLogout:
		return "LOGOUT "
	case CmdTypeUnselect:
		return "USELECT"
	case CmdTypeLogin:
		return "LOGIN  "
	case CmdTypeExamine:
		return "EXAMINE"

	default:
		return "Unknown"
	}
}

// CmdProfiler is the interface that can be used to perform measurements related to the execution
// scope of incoming IMAP commands.
type CmdProfiler interface {
	// Start will be called once the command has been received and interpreted.
	Start(cmdType int)
	// Stop will be called once the command has finished executing and all the replies sent to the client.
	Stop(cmdType int)
}

// CmdProfilerBuilder is the interface through which an instance of the CmdProfiler gets created. One of these will be
// created for each connecting IMAP client.
type CmdProfilerBuilder interface {
	// New creates a new CmdProfiler instance.
	New() CmdProfiler

	// Collect will be called when the IMAP client has disconnected/logged out.
	Collect(profiler CmdProfiler)
}

// NullCmdProfiler represents a null implementation of CmdProfiler.
type NullCmdProfiler struct{}

func (*NullCmdProfiler) Start(int) {}

func (*NullCmdProfiler) Stop(int) {}

type NullCmdExecProfilerBuilder struct{}

func (*NullCmdExecProfilerBuilder) New() CmdProfiler {
	return &NullCmdProfiler{}
}

func (*NullCmdExecProfilerBuilder) Collect(CmdProfiler) {}
