package gfont

import (
	"fmt"
	"strings"
	"net/http"
	"net/url"
	"io/ioutil"
)

const (
	apiBaseURL = "https://fonts.googleapis.com/css2"
	gstaticBaseURL = "https://fonts.gstatic.com/s/"
	svgFontBaseURL = "https://fonts.gstatic.com/l/font?"
)

// FontProfile is a font profile for specific devices and format
type FontProfile int
const (
	// WOFF2 font format supported by most modern browsers
	WOFF2 FontProfile = iota
	// AppleWOFF2 is WOFF2 font format optimized for Apple devices
	AppleWOFF2
	// LegacyWOFF2 is WOFF2 font format without support for Unicode range feature
	LegacyWOFF2
	// AppleLegacyWOFF2 is WOFF2 font format optimized for Apple devices, without support for Unicode range feature
	AppleLegacyWOFF2
	// WOFF font format supported by most modern browsers
	WOFF
	// AppleWOFF is WOFF font format optimized for Apple devices
	AppleWOFF
	// LegacyWOFF is WOFF font format without support for Unicode range feature
	LegacyWOFF
	// AppleLegacyWOFF is WOFF font format optimized for Apple devices, without support for Unicode range feature
	AppleLegacyWOFF
	// TTF is TTF font format
	TTF
	// AppleTTF is TTF font format optimized for Apple devices
	AppleTTF
	// SVG is SVG font format
	SVG
	// EOT is EOT font format
	EOT
)

var useragent = map[FontProfile]string{
	WOFF2:            "Mozilla/5.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36",
    AppleWOFF2:       "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.10; rv:62.0) Gecko/20100101 Firefox/62.0",
    LegacyWOFF2:      "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1",
    AppleLegacyWOFF2: "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; en) Gecko/20100101 Firefox/40.1",
    WOFF:             "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
    AppleWOFF:        "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; en) AppleWebKit/418 (KHTML, like Gecko) Safari/417.9.2",
    LegacyWOFF:       "Mozilla/4.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1667.0 Safari/537.36",
    AppleLegacyWOFF:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.124 Safari/537.36",
    TTF:              "Mozilla/5.0",
    AppleTTF:         "Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_4_11; en) AppleWebKit/528.4+ (KHTML, like Gecko) Version/4.0dp1 Safari/526.11.2",
    SVG:              "(iPad) AppleWebKit/534",
    EOT:              "MSIE 8.0",
}

// GetURL returns the URL for font-face CSS from Google API
func GetURL(ua FontProfile, fontFamily, fontStyle string, mirror *url.URL) string {
	_, ok := useragent[ua]
	if !ok {
		return ""
	}

	apiURL := apiBaseURL
	if mirror != nil {
		apiURL = mirror.String()
	}
	return fmt.Sprintf("%s?family=%s:%s", apiURL, strings.Replace(fontFamily, " ", "+", -1), fontStyle)
}

// DownloadCSS downloads font-face CSS from Google API
func DownloadCSS(ua FontProfile, fontFamily, fontStyle string, mirror *url.URL) ([]byte, error) {
	uaString, ok := useragent[ua]
	if !ok {
		return nil, fmt.Errorf("ua unsupported")
	}

	apiURL := apiBaseURL
	if mirror != nil {
		apiURL = mirror.String()
	}

	cssURL := fmt.Sprintf("%s?family=%s:%s", apiURL, strings.Replace(fontFamily, " ", "+", -1), fontStyle)
	req, errReq := http.NewRequest("GET", cssURL, nil)
	if errReq != nil {
		return nil, errReq
	}
	req.Header.Set("user-agent", uaString)
	response, err := http.DefaultClient.Do(req)
	if err != nil { 
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	_ = response.Body.Close()

	return body, nil
}