package filesystem

// Visibility constants to mimic the Visibility class in the original PHP code.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// PortableVisibilityConverter provides an implementation of VisibilityConverter.
type PortableVisibilityConverter struct {
	filePublic            int
	filePrivate           int
	directoryPublic       int
	directoryPrivate      int
	defaultForDirectories string
}

func NewDefault() *PortableVisibilityConverter {
	filePublic := 0644
	filePrivate := 0600
	dirPublic := 0755
	dirPrivate := 0700

	return NewPortableVisibilityConverter(
		filePublic,
		filePrivate,
		dirPublic,
		dirPrivate,
		VisibilityPrivate,
	)
}

// NewPortableVisibilityConverter creates a new PortableVisibilityConverter with the given settings.
func NewPortableVisibilityConverter(
	filePublic int,
	filePrivate int,
	directoryPublic int,
	directoryPrivate int,
	defaultForDirectories string,
) *PortableVisibilityConverter {
	return &PortableVisibilityConverter{
		filePublic:            filePublic,
		filePrivate:           filePrivate,
		directoryPublic:       directoryPublic,
		directoryPrivate:      directoryPrivate,
		defaultForDirectories: defaultForDirectories,
	}
}

// ForFile converts the visibility string to an integer representation for files.
func (p *PortableVisibilityConverter) ForFile(visibility string) int {
	if !p.isValidVisibility(visibility) {
		panic("invalid visibility value")
	}
	if visibility == VisibilityPublic {
		return p.filePublic
	}
	return p.filePrivate
}

// ForDirectory converts the visibility string to an integer representation for directories.
func (p *PortableVisibilityConverter) ForDirectory(visibility string) int {
	if !p.isValidVisibility(visibility) {
		panic("invalid visibility value")
	}
	if visibility == VisibilityPublic {
		return p.directoryPublic
	}
	return p.directoryPrivate
}

// InverseForFile converts the integer visibility back to string representation for files.
func (p *PortableVisibilityConverter) InverseForFile(visibility int) string {
	if visibility == p.filePublic {
		return VisibilityPublic
	} else if visibility == p.filePrivate {
		return VisibilityPrivate
	}
	return VisibilityPublic // default
}

// InverseForDirectory converts the integer visibility back to string representation for directories.
func (p *PortableVisibilityConverter) InverseForDirectory(visibility int) string {
	if visibility == p.directoryPublic {
		return VisibilityPublic
	} else if visibility == p.directoryPrivate {
		return VisibilityPrivate
	}
	return VisibilityPublic // default
}

// DefaultForDirectories returns the default visibility setting for directories as an integer.
func (p *PortableVisibilityConverter) DefaultForDirectories() int {
	if p.defaultForDirectories == VisibilityPublic {
		return p.directoryPublic
	}
	return p.directoryPrivate
}

// isValidVisibility checks if the provided visibility value is valid.
func (p *PortableVisibilityConverter) isValidVisibility(visibility string) bool {
	return visibility == VisibilityPublic || visibility == VisibilityPrivate
}

// FromArray creates a PortableVisibilityConverter from a permission map.
func FromArray(permissionMap map[string]map[string]int, defaultForDirectories string) *PortableVisibilityConverter {
	filePublic := permissionMap["file"]["public"]
	filePrivate := permissionMap["file"]["private"]
	dirPublic := permissionMap["dir"]["public"]
	dirPrivate := permissionMap["dir"]["private"]

	return NewPortableVisibilityConverter(
		filePublic,
		filePrivate,
		dirPublic,
		dirPrivate,
		defaultForDirectories,
	)
}
