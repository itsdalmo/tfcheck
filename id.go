package tfcheck

import "sync"

// Global ID used to address tea.Msg to the correct model.
var (
	lastID int
	idMtx  sync.Mutex
)

// nextID returns the next global ID.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()

	lastID++
	return lastID
}
