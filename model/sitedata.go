package model

// A SiteData struct for json datatype
type SiteData struct {
	Model
	ErrorType string `json:"errorType"`
	ErrorMsg  string `json:"errorMsg"`
	URL       string `json:"url"`
	URLMain   string `json:"urlMain"`
	URLProbe  string `json:"urlProbe"`
	URLError  string `json:"errorUrl"`
	// UsedUsername   string `json:"username_claimed"`
	// UnusedUsername string `json:"username_unclaimed"`
	// RegexCheck string `json:"regexCheck"`
	// Rank int`json:"rank"`
}
