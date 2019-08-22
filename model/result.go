package model

// Result of Investigo function
type Result struct {
	Model
	Usernane string
	Exist    bool
	Proxied  bool
	Site     string
	URL      string
	URLProbe string
	Link     string
	Err      bool
	ErrMsg   string
}
