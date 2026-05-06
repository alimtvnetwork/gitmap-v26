package clonepick

// picker.go: bubbletea TUI for `gitmap clone-pick --ask`. Renders a
// scrollable flat list of every tracked path returned by listRepoPaths
// so the user can pick which paths to sparse-checkout.
//
// Keys (spec §"--ask picker"):
//
//	up/k, down/j  move cursor
//	space         toggle current row
//	a             select all (excluding auto-greyed rows)
//	n             select none
//	s             save & continue (returns picked paths)
//	q / ctrl-c    cancel (returns ErrPickerCancelled -> exit 130)
//
// Pre-selected rows: anything in plan.Paths (the user-supplied list
// passed on the command line). Auto-greyed rows: anything matching
// constants.ClonePickAutoExclude -- still toggleable individually.

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrPickerCancelled is returned by RunPicker when the user pressed
// q / ctrl-c. The cmd layer maps this to exit code 130 and prints
// MsgClonePickUserCancelled.
var ErrPickerCancelled = errors.New("clone-pick: picker cancelled")

// RunPicker enumerates plan.RepoUrl, opens the bubbletea picker, and
// returns the user-picked subset. plan.Paths seeds the initial
// selection so re-running with the same args is a no-op confirmation.
func RunPicker(plan Plan) ([]string, error) {
	picked, tmp, err := RunPickerKeep(plan)
	if len(tmp) > 0 {
		os.RemoveAll(tmp)
	}

	return picked, err
}

// RunPickerKeep is the clone-once variant: returns the picked paths
// AND the temp metadata-clone directory so the executor can promote
// it instead of re-cloning. The caller owns tmp and must remove it
// (or pass it to the executor via Plan.PreClonedSrc, which moves it
// into place). On error or cancellation tmp is already cleaned up.
func RunPickerKeep(plan Plan) ([]string, string, error) {
	all, tmp, err := ListRepoPathsKeep(plan)
	if err != nil {
		return nil, "", err
	}
	model := newPickerModel(all, plan.Paths)
	prog := tea.NewProgram(model)
	final, runErr := prog.Run()
	if runErr != nil {
		os.RemoveAll(tmp)

		return nil, "", fmt.Errorf("clone-pick: picker run: %w", runErr)
	}
	finished, _ := final.(pickerModel)
	if finished.cancelled {
		os.RemoveAll(tmp)

		return nil, "", ErrPickerCancelled
	}

	return finished.selected(), tmp, nil
}

// pickerModel is the bubbletea model for the picker. Kept tiny so
// each method stays under the 15-line cap; rendering is delegated to
// picker_view.go.
type pickerModel struct {
	paths     []string
	picked    map[int]bool
	cursor    int
	// viewportHeight is the number of rows the row window can show
	// at once (terminal height minus header + footer chrome). Set
	// from tea.WindowSizeMsg; defaults to defaultViewportHeight
	// when the terminal hasn't reported a size yet.
	viewportHeight int
	// scrollOffset is the index of the first row currently visible.
	// Always in [0, len(paths)-viewportHeight] -- clamped by
	// clampScroll after every cursor move.
	scrollOffset int
	cancelled    bool
	done         bool
}

// defaultViewportHeight is the row-window size used until bubbletea
// reports a real terminal height via tea.WindowSizeMsg. 20 rows fits
// comfortably in any terminal we care about and matches the muscle
// memory of `less -F` users.
const defaultViewportHeight = 20

// chromeRows is the number of rows reserved for the header line and
// footer key-hint line (both newline-terminated). Subtracted from the
// terminal height so the row window doesn't push the footer offscreen.
const chromeRows = 3

func newPickerModel(all, preselected []string) pickerModel {
	picked := make(map[int]bool, len(preselected))
	preset := make(map[string]struct{}, len(preselected))
	for _, p := range preselected {
		preset[p] = struct{}{}
	}
	for i, path := range all {
		if _, ok := preset[path]; ok {
			picked[i] = true
		}
	}

	return pickerModel{
		paths:          all,
		picked:         picked,
		viewportHeight: defaultViewportHeight,
	}
}

// Init is required by tea.Model. Nothing to schedule on startup.
func (m pickerModel) Init() tea.Cmd { return nil }

// Update routes key events to handleKey and resize events to
// handleResize. Other message types are ignored -- the picker is a
// pure keyboard UI.
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(t)
	case tea.WindowSizeMsg:
		return m.handleResize(t), nil
	}

	return m, nil
}

// handleResize updates the viewport height to match the terminal,
// clamps the scroll offset, and returns the updated model. Reserves
// chromeRows for the header + footer so the row window never pushes
// either offscreen.
func (m pickerModel) handleResize(msg tea.WindowSizeMsg) pickerModel {
	height := msg.Height - chromeRows
	if height < 1 {
		height = 1
	}
	m.viewportHeight = height
	m.scrollOffset = clampScroll(m.cursor, m.scrollOffset, height, len(m.paths))

	return m
}

// handleKey implements every bound key. Returning tea.Quit on q / s
// is what unblocks tea.Program.Run() in RunPicker.
func (m pickerModel) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		m.cancelled = true

		return m, tea.Quit
	case "s", "enter":
		m.done = true

		return m, tea.Quit
	}

	return m.handleNavKey(k), nil
}

// handleNavKey handles cursor + selection toggles. Split out so
// handleKey stays under the function-length cap. Cursor movement
// re-clamps scrollOffset so the cursor row is always in view.
func (m pickerModel) handleNavKey(k tea.KeyMsg) pickerModel {
	m = applyCursorMove(m, k.String())
	switch k.String() {
	case " ":
		m.picked[m.cursor] = !m.picked[m.cursor]
	case "a":
		m.selectAll()
	case "n":
		m.picked = make(map[int]bool)
	}
	m.scrollOffset = clampScroll(m.cursor, m.scrollOffset,
		m.viewportHeight, len(m.paths))

	return m
}

// applyCursorMove updates the cursor for navigation keys. Page keys
// jump by viewportHeight; g / G are vim-style home / end. Toggle
// keys (space, a, n) fall through unchanged -- handleNavKey owns
// the selection mutation.
func applyCursorMove(m pickerModel, key string) pickerModel {
	switch key {
	case "up", "k":
		m.cursor = maxInt(0, m.cursor-1)
	case "down", "j":
		m.cursor = minInt(len(m.paths)-1, m.cursor+1)
	case "pgup", "ctrl+b":
		m.cursor = maxInt(0, m.cursor-m.viewportHeight)
	case "pgdown", "ctrl+f":
		m.cursor = minInt(len(m.paths)-1, m.cursor+m.viewportHeight)
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = maxInt(0, len(m.paths)-1)
	}

	return m
}

// clampScroll keeps the cursor in the visible window. Returns the
// new offset such that cursor in [offset, offset+height). Pure --
// no model mutation -- so the caller decides when to commit it.
func clampScroll(cursor, offset, height, total int) int {
	if height < 1 || total == 0 {
		return 0
	}
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+height {
		return cursor - height + 1
	}
	maxOffset := total - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		return maxOffset
	}

	return offset
}

// minInt / maxInt: tiny stdlib-free helpers so picker.go has no
// external dep beyond bubbletea. Kept private to the package.
func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// selectAll picks every non-auto-greyed row. Auto-greyed rows stay
// off so a careless "a" doesn't drag node_modules/ into the clone.
func (m *pickerModel) selectAll() {
	for i, path := range m.paths {
		if !IsAutoExcluded(path) {
			m.picked[i] = true
		}
	}
}

// selected returns the picked paths in their original order so the
// resulting Plan.Paths is stable across runs (matches normalisePaths
// which sorts -- the cmd layer re-normalises after the picker
// returns).
func (m pickerModel) selected() []string {
	out := make([]string, 0, len(m.picked))
	for i, path := range m.paths {
		if m.picked[i] {
			out = append(out, path)
		}
	}

	return out
}
