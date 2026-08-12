package repo

import (
	"context"
	"sync"
	"time"
)

const fileListCacheTTL = 45 * time.Second

type fileListCacheEntry struct {
	listing   FileListing
	expiresAt time.Time
}

var (
	fileListCacheMu sync.Mutex
	fileListCache   = map[string]fileListCacheEntry{}
)

// cachedFiles returns the full sorted listing for r, reusing a short-lived
// process-local entry so paginated /repo/files warm requests do not re-run
// git ls-files on every page.
//funclogmeasure:skip category=hot-path reason="Cache helper; Files/ListFiles emit the operation-level trace."
func (r *Root) cachedFiles(ctx context.Context) (FileListing, error) {
	key := r.abs
	now := time.Now()

	fileListCacheMu.Lock()
	if ent, ok := fileListCache[key]; ok && now.Before(ent.expiresAt) {
		listing := ent.listing
		fileListCacheMu.Unlock()
		return listing, nil
	}
	fileListCacheMu.Unlock()

	listing, err := r.Files(ctx)
	if err != nil {
		return FileListing{}, err
	}

	fileListCacheMu.Lock()
	fileListCache[key] = fileListCacheEntry{
		listing:   listing,
		expiresAt: now.Add(fileListCacheTTL),
	}
	fileListCacheMu.Unlock()
	return listing, nil
}

// resetFileListCacheForTest clears the TTL cache (unit tests only).
//funclogmeasure:skip category=hot-path reason="Test-only cache reset without production I/O path."
func resetFileListCacheForTest() {
	fileListCacheMu.Lock()
	fileListCache = map[string]fileListCacheEntry{}
	fileListCacheMu.Unlock()
}
