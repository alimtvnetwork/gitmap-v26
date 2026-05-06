package clonepick

// replay.go: load a previously persisted Plan from the DB and refresh
// its CreatedAt timestamp so most-recently-replayed selections sort to
// the top of `gitmap clone-pick --list` (planned).
//
// Replay rules (spec/01-app/100-clone-pick.md §"Replay rules"):
//   - Numeric ref  -> SELECT by SelectionId
//   - Non-numeric  -> SELECT by Name (case-sensitive, newest match wins)
//   - Replay does NOT insert a duplicate row; it bumps CreatedAt instead.
//   - --dry-run never writes to the DB (Touch is skipped by the caller).

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/gitmap-v16/gitmap/constants"
)

// Loader is the read surface needed by --replay. Split from Persister
// (the write surface) so a future read-only test fake can implement
// just the lookup half.
type Loader interface {
	LoadClonePickByID(id int64) (Plan, error)
	LoadClonePickByName(name string) (Plan, error)
	TouchClonePickCreatedAt(id int64) error
}

// LoadFromDB resolves ref to a saved Plan. Numeric refs hit the ID
// index; everything else is treated as a Name lookup. Returns a
// wrapped error using the user-facing "no saved selection" message
// when nothing matches so the cmd layer can print verbatim.
func LoadFromDB(loader Loader, ref string) (Plan, error) {
	if loader == nil {
		return Plan{}, fmt.Errorf("clone-pick: --replay requires database access")
	}
	trimmed := strings.TrimSpace(ref)
	if len(trimmed) == 0 {
		return Plan{}, fmt.Errorf(constants.MsgClonePickReplayNotFound, ref)
	}
	if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		plan, loadErr := loader.LoadClonePickByID(id)
		if loadErr != nil {
			return Plan{}, fmt.Errorf(constants.MsgClonePickReplayNotFound, ref)
		}

		return plan, nil
	}
	plan, err := loader.LoadClonePickByName(trimmed)
	if err != nil {
		return Plan{}, fmt.Errorf(constants.MsgClonePickReplayNotFound, ref)
	}

	return plan, nil
}

// TouchAfterReplay bumps CreatedAt on the replayed row. Best-effort:
// a failure is logged by the caller but never fails the replay.
// Skipped when dryRun is true so dry-runs stay read-only.
func TouchAfterReplay(loader Loader, id int64, dryRun bool) error {
	if loader == nil || id <= 0 || dryRun {
		return nil
	}

	return loader.TouchClonePickCreatedAt(id)
}
