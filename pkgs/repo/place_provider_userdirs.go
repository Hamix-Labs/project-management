package repo

// UserDirsPlaceProvider returns OS-resolved profile folders (Documents, Desktop, …)
// as first-class picker roots. Skipped when HAMIX_BROWSE_ROOTS is set.
type UserDirsPlaceProvider struct{}

// Places implements PlaceProvider.
//
//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveBrowseRoots."
func (UserDirsPlaceProvider) Places(_ BrowseEnvironment, _ string) ([]Place, error) {
	return resolveUserDirPlaces()
}
