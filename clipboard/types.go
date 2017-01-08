package clipboard

type Clipboarder interface {
	SetString(s string) error
	GetString() (string, error)
}
