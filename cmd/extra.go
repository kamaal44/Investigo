package cmd

import (
	"github.com/lucmski/Investigo/model"
	//"github.com/lucmski/Investigo/service"
)

// Initialize sites not included in Sherlock
func initializeExtraSiteData() {
	siteData["Pornhub"] = model.SiteData{
		ErrorType: "status_code",
		URLMain:   "https://www.pornhub.com/",
		URL:       "https://www.pornhub.com/users/{}",
	}
	siteData["NAVER"] = model.SiteData{
		ErrorType: "status_code",
		URLMain:   "https://www.naver.com/",
		URL:       "https://blog.naver.com/{}",
	}
	siteData["xvideos"] = model.SiteData{
		ErrorType: "status_code",
		URLMain:   "https://xvideos.com/",
		URL:       "https://xvideos.com/profiles/{}",
	}
}

// "https://api.ipify.org"
// https://www.ipify.org/
//
