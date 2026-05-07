package constants

// gitmap install ctx — Windows right-click context menu (v1).
// Spec: spec/04-generic-cli/30-install-ctx.md.

// Tool name for the install dispatcher.
const (
	ToolCtx     = "ctx"
	ToolCtxDesc = "Add gitmap to Windows right-click context menu"
)

// Top-level cascade label and registry root key names. The two roots
// give us the menu both on folder backgrounds (clicking inside a
// folder) and on folder items (right-clicking the folder itself).
const (
	CtxRootKeyBackground = `HKCU\Software\Classes\Directory\Background\shell\gitmap`
	CtxRootKeyDirectory  = `HKCU\Software\Classes\Directory\shell\gitmap`
	CtxRootMUIVerb       = "gitmap"
)

// CtxMode controls how an entry is wired into the registry.
type CtxMode int

// CtxMode values.
const (
	CtxModeTerminal CtxMode = iota // pwsh -NoExit, command runs and window stays open
	CtxModeSilent                  // pwsh -WindowStyle Hidden, output via notifier
	CtxModePrefill                 // pwsh -NoExit + writes "gitmap " prompt, no command run
)

// User-facing labels and exec messages.
const (
	MsgCtxInstallStart    = "  Adding gitmap to Windows context menu...\n"
	MsgCtxInstallDone     = "  ✓ gitmap context menu installed (%d/%d registry keys).\n"
	MsgCtxUninstallStart  = "  Removing gitmap from Windows context menu...\n"
	MsgCtxUninstallDone   = "  ✓ gitmap context menu removed (%d/%d registry keys).\n"
	MsgCtxRegFail         = "  ! Registry command failed: %v\n"
	MsgCtxOSUnsupported   = "  Error: ctx is only supported on Windows (current OS: %s)\n"
	MsgCtxOpenTerminalLbl = "Open terminal here"
	MsgCtxDocsLbl         = "Docs"
)
