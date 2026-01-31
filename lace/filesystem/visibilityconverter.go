package filesystem

// VisibilityConverter is an interface for converting visibility settings
// between integer representations and string representations.
type VisibilityConverter interface {
	// ForFile converts the visibility string to an integer representation for files.
	ForFile(visibility string) int

	// ForDirectory converts the visibility string to an integer representation for directories.
	ForDirectory(visibility string) int

	// InverseForFile converts the integer visibility back to string representation for files.
	InverseForFile(visibility int) string

	// InverseForDirectory converts the integer visibility back to string representation for directories.
	InverseForDirectory(visibility int) string

	// DefaultForDirectories returns the default visibility setting for directories as an integer.
	DefaultForDirectories() int
}
