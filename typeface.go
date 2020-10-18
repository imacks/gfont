package gfont

import (
	"fmt"
	"strings"
	"strconv"
	"net/url"
	"encoding/json"

	"github.com/gorilla/css/scanner"
)

// Typeface represents a font face
type Typeface struct {
	Format string          `json:"format"`
	Weight int             `json:"weight"`
	Family string          `json:"family"`
	Style string           `json:"style"`
	URL *url.URL           `json:"url"`
	UnicodeRange []string  `json:"unicodeRange,omitempty"`
}

// Typefaces represents a collection of Typeface
type Typefaces struct {
	Fonts []Typeface  `json:"fonts"`
}

// UnmarshalCSS unmarshals a byte slice to Typefaces
func UnmarshalCSS(cssBytes []byte, typefaces *Typefaces) error {
	if typefaces == nil {
		return fmt.Errorf("typefaces cannot be nil")
	}
	result := []Typeface{}

	s := scanner.New(string(cssBytes))
	for {
		ptoken := nextSig(s)
		if ptoken.Type == scanner.TokenEOF || ptoken.Type == scanner.TokenError {
			break
		}
		if ptoken.Type != scanner.TokenAtKeyword || ptoken.Value != "@font-face" {
			continue
		}
 		lastGoodToken := ptoken.String()

		ptoken = nextSig(s)
		if ptoken.Type != scanner.TokenChar || ptoken.Value != "{" {
			return fmt.Errorf("expect { after %s", lastGoodToken)
		}
		lastGoodToken = ptoken.String()

		fface := Typeface{}
		for {
			token := nextSig(s)
			if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
				return fmt.Errorf("unexpected EOF after %s", lastGoodToken)
			}

			if token.Type == scanner.TokenChar && token.Value == "}" {
				break
			}

			if token.Type == scanner.TokenIdent && token.Value == "font-family" {
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ":" {
					return fmt.Errorf("expect : after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenString {
					return fmt.Errorf("expect <string> after %s", lastGoodToken)
				}
				fface.Family = strings.Trim(token.Value, "'")

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ";" {
					return fmt.Errorf("expect ; after %s", lastGoodToken)
				}
				continue
			}

			if token.Type == scanner.TokenIdent && token.Value == "font-style" {
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ":" {
					return fmt.Errorf("expect : after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenIdent {
					return fmt.Errorf("expect <ident> after %s", lastGoodToken)
				}
				fface.Style = token.Value

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ";" {
					return fmt.Errorf("expect ; after %s", lastGoodToken)
				}
				continue
			}

			if token.Type == scanner.TokenIdent && token.Value == "font-weight" {
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ":" {
					return fmt.Errorf("expect : after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenNumber {
					return fmt.Errorf("expect <number> after %s", lastGoodToken)
				}
				var errNum error
				fface.Weight, errNum = strconv.Atoi(token.Value)
				if errNum != nil {
					return fmt.Errorf("convert token to number failed after %s: %v", lastGoodToken, errNum)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ";" {
					return fmt.Errorf("expect ; after %s", lastGoodToken)
				}
				continue
			}

			// #todo support multiple urls
			if token.Type == scanner.TokenIdent && token.Value == "src" {
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ":" {
					return fmt.Errorf("expect : after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenURI {
					return fmt.Errorf("expect <uri> after %s", lastGoodToken)
				}			
				urlString := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(token.Value, "url("), ")"), "'")
				var errURL error
				fface.URL, errURL = url.Parse(urlString)
				if errURL != nil {
					return fmt.Errorf("parse url failed after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				// EOT format
				//   src: url(https://xxx.eot);
				if token.Type == scanner.TokenChar && token.Value == ";" {
					fface.Format = "eot"
					continue
				}

				if token.Type != scanner.TokenFunction || token.Value != "format(" {
					return fmt.Errorf("expect format( after %s but got %s", lastGoodToken, token.String())
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenString {
					return fmt.Errorf("expect <string> after %s", lastGoodToken)
				}
				fface.Format = strings.Trim(token.Value, "'")
				
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ")" {
					return fmt.Errorf("expect ) after %s", lastGoodToken)
				}

				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ";" {
					return fmt.Errorf("expect ; after %s", lastGoodToken)
				}
				continue
			}

			if token.Type == scanner.TokenIdent && token.Value == "unicode-range" {
				lastGoodToken = token.String()
				token = nextSig(s)
				if token.Type != scanner.TokenChar || token.Value != ":" {
					return fmt.Errorf("expect : after %s", lastGoodToken)
				}

				unicodeRange := []string{}
				for {
					subtoken := nextSig(s)
					if subtoken.Type != scanner.TokenUnicodeRange {
						if subtoken.Type == scanner.TokenEOF || subtoken.Type == scanner.TokenError {
							return fmt.Errorf("unexpected EOF after %s", lastGoodToken)
						} else if subtoken.Type == scanner.TokenChar && subtoken.Value == "," {
							continue
						} else if subtoken.Type == scanner.TokenChar && subtoken.Value == ";" {
							break
						}
					}

					lastGoodToken = subtoken.String()
					unicodeRange = append(unicodeRange, subtoken.Value)
				}
				fface.UnicodeRange = unicodeRange
				continue
			}
		}

		result = append(result, fface)
	}

	typefaces.Fonts = result
	return nil
}

// Select returns a slice of Typeface based on the selection criteria
func (ts *Typefaces) Select(format, family, style string, weight int) []Typeface {
	filtered := []Typeface{}
	for _, v := range ts.Fonts {
		if format != "" && v.Format != format {
			continue
		}
		if family != "" && v.Family != family {
			continue
		}
		if style != "" && v.Style != style {
			continue
		}
		if weight != -1 && v.Weight != weight {
			continue
		}
		filtered = append(filtered, v)
	}

	return filtered
}

// CSS returns the CSS for all fonts in the collection, in the most legacy compatible manner
func (ts *Typefaces) CSS() string {
	cssTpl := `@font-face{font-family:%s;font-style:%s;font-weight:%d;`

	families := ts.Family()
	weights := ts.Weight()
	styles := ts.Style()
	formats := ts.Format()

	result := []string{}
	for _, fam := range families {
		fontFamily := fam
		if strings.Contains(fam, " ") {
			fontFamily = "'" + fam + "'"
		}

		for _, w := range weights {
			for _, s := range styles {
				cssPart := fmt.Sprintf(cssTpl, fontFamily, s, w)

				selected := map[string][]Typeface{}
				for _, m := range formats {
					filtered := ts.Select(m, fam, s, w)
					if len(filtered) > 0 {
						selected[m] = filtered
					}
				}

				if v, ok := selected["eot"]; ok {
					cssPart = cssPart + fmt.Sprintf("src:url('%s');", v[0].URL.String())
					cssPart = cssPart + fmt.Sprintf("src:url('%s') format('embedded-opentype')", v[0].URL.String()+"?#iefix")
					if len(selected) == 1 {
						cssPart = cssPart + ";"
					} else {
						cssPart = cssPart + ","
					}
				} else {
					cssPart = cssPart + "src:"
				}

				if v, ok := selected["woff2"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('woff2'),", v[0].URL.String())
				}
				if v, ok := selected["woff"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('woff'),", v[0].URL.String())
				}
				if v, ok := selected["ttf"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('ttf'),", v[0].URL.String())
				}
				if v, ok := selected["svg"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('svg'),", v[0].URL.String())
				}

				cssPart = strings.TrimSpace(cssPart)
				cssPart = strings.TrimSuffix(cssPart, ",")
				cssPart = strings.TrimSuffix(cssPart, ";")
				cssPart = cssPart + "}"

				result = append(result, cssPart)
			}
		}
	}

	return strings.Join(result, "")
}

// PrettyCSS is the human readable version of method CSS
func (ts *Typefaces) PrettyCSS() string {
	cssTpl := `@font-face {
	font-family: %s;
	font-style: %s; 
	font-weight: %d;`

	families := ts.Family()
	weights := ts.Weight()
	styles := ts.Style()
	formats := ts.Format()

	result := []string{}
	for _, fam := range families {
		fontFamily := fam
		if strings.Contains(fam, " ") {
			fontFamily = "'" + fam + "'"
		}

		for _, w := range weights {
			for _, s := range styles {
				cssPart := fmt.Sprintf(cssTpl, fontFamily, s, w)

				selected := map[string][]Typeface{}
				for _, m := range formats {
					filtered := ts.Select(m, fam, s, w)
					if len(filtered) > 0 {
						selected[m] = filtered
					}
				}

				if v, ok := selected["eot"]; ok {
					cssPart = cssPart + fmt.Sprintf("\n\tsrc: url('%s');", v[0].URL.String())
					cssPart = cssPart + fmt.Sprintf("\n\tsrc: url('%s') format('embedded-opentype')", v[0].URL.String()+"?#iefix")
					if len(selected) == 1 {
						cssPart = cssPart + ";"
					} else {
						cssPart = cssPart + ",\n\t\t"
					}
				} else {
					cssPart = cssPart + "\n\tsrc: "
				}

				if v, ok := selected["woff2"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('woff2'),\n\t\t", v[0].URL.String())
				}
				if v, ok := selected["woff"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('woff'),\n\t\t", v[0].URL.String())
				}
				if v, ok := selected["ttf"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('ttf'),\n\t\t", v[0].URL.String())
				}
				if v, ok := selected["svg"]; ok {
					cssPart = cssPart + fmt.Sprintf("url('%s') format('svg'),\n\t\t", v[0].URL.String())
				}

				cssPart = strings.TrimSpace(cssPart)
				cssPart = strings.TrimSuffix(cssPart, ",")
				if !strings.HasSuffix(cssPart, ";") {
					cssPart = cssPart + ";"
				}
				cssPart = cssPart + "\n}"

				result = append(result, cssPart)
			}
		}
	}

	return strings.Join(result, "\n") + "\n"
}

// Format returns a unique list of font formats
func (ts *Typefaces) Format() []string {
	result := []string{}
	for _, v := range ts.Fonts {
		if v.Format == "" {
			continue
		}

		if !isUniqueString(result, v.Format) {
			continue
		}

		result = append(result, v.Format)
	}
	return result
}

// Family returns a unique list of font families
func (ts *Typefaces) Family() []string {
	result := []string{}
	for _, v := range ts.Fonts {
		if v.Family == "" {
			continue
		}

		if !isUniqueString(result, v.Family) {
			continue
		}

		result = append(result, v.Family)
	}
	return result
}

// Style returns a unique list of font style
func (ts *Typefaces) Style() []string {
	result := []string{}
	for _, v := range ts.Fonts {
		if v.Style == "" {
			continue
		}

		if !isUniqueString(result, v.Style) {
			continue
		}

		result = append(result, v.Style)
	}
	return result
}

// Weight returns a unique list of font weights
func (ts *Typefaces) Weight() []int {
	result := []int{}
	for _, v := range ts.Fonts {
		if v.Weight < 1 {
			continue
		}

		if !isUniqueInt(result, v.Weight) {
			continue
		}

		result = append(result, v.Weight)
	}
	return result
}

// URL returns a unique list of font URLs
func (ts *Typefaces) URL() []*url.URL {
	result := []*url.URL{}
	resultString := []string{}

	for _, v := range ts.Fonts {
		if v.Style == "" {
			continue
		}

		if !isUniqueString(resultString, v.URL.String()) {
			continue
		}

		result = append(result, v.URL)
	}
	return result
}

// MarshalJSON returns a JSON representation of Typeface
func (t *Typeface) MarshalJSON() ([]byte, error) {
	type tfAlias Typeface
	return json.Marshal(&struct {
		URL string      `json:"url"`
		Version  string `json:"version"`
		FileName string `json:"filename"`
		*tfAlias
	}{
		URL: t.URL.String(),
		Version: t.Version(),
		FileName: t.FileName(),
		tfAlias: (*tfAlias)(t),
	})
}

// UnmarshalJSON converts a JSON representation of Typeface to Typeface
func (t *Typeface) UnmarshalJSON(data []byte) error {
	type tfAlias Typeface
	aux := &struct {
		URL string   `json:"url"`
		*tfAlias
	}{
		tfAlias: (*tfAlias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	u, errURL := url.Parse(aux.URL)
	if errURL != nil {
		return errURL
	}
	t.URL = u

	return nil
}

func (t *Typeface) String() string {
	return fmt.Sprintf("%s %s %d", t.Family, t.Style, t.Weight)
}

// Version returns the font version by parsing the URL
func (t *Typeface) Version() string {
	if t.URL == nil {
		return ""
	}

	if t.URL.RawQuery != "" {
		// /l/font?kit=L0xhDFMnlVwD4h3Lt9JWnbX3jG-2X3LAI18&skey=ea73fc1e1d1dfd9a&v=v10#Domine
		return t.URL.Query().Get("v")
	}

	pathParts := strings.Split(t.URL.Path, "/")
	if !strings.HasPrefix(pathParts[len(pathParts)-2], "v") {
		return ""
	}
	return pathParts[len(pathParts)-2]
}

// FileName returns the font filename by parsing the URL
func (t *Typeface) FileName() string {
	if t.URL == nil {
		return ""
	}

	if t.URL.RawQuery != "" {
		// /l/font?kit=L0xhDFMnlVwD4h3Lt9JWnbX3jG-2X3LAI18&skey=ea73fc1e1d1dfd9a&v=v10#Domine
		kitUID := t.URL.Query().Get("kit")
		if kitUID == "" {
			return ""
		}
		return kitUID + ".svg"
	}

	pathParts := strings.Split(t.URL.Path, "/")
	return pathParts[len(pathParts)-1]
}

// CSS returns the CSS representation of a TypeFace
func (t *Typeface) CSS() string {
	cssTpl := `@font-face{font-family:%s;font-style:%s;font-weight:%d;src:url('%s') format('%s');`

	if len(t.UnicodeRange) > 0 {
		cssTpl = cssTpl + "unicode-range:%s;"
	}
	cssTpl = strings.TrimSuffix(cssTpl, ";")
	cssTpl = cssTpl + "}"

	fontFamily := t.Family
	if strings.Contains(t.Family, " ") {
		fontFamily = "'" + t.Family + "'"
	}

	if len(t.UnicodeRange) > 0 {
		return fmt.Sprintf(cssTpl, fontFamily, t.Style, t.Weight, t.URL, t.Format, strings.Join(t.UnicodeRange, ","))
	}

	return fmt.Sprintf(cssTpl, fontFamily, t.Style, t.Weight, t.URL, t.Format)
}

// PrettyCSS is the human readable version of method CSS
func (t *Typeface) PrettyCSS() string {
	cssTpl := `@font-face {
	font-family: %s;
	font-style: %s;
	font-weight: %d;
	src: url('%s') format('%s');`

	if len(t.UnicodeRange) > 0 {
		cssTpl = cssTpl + "\n\tunicode-range: %s;"
	}
	cssTpl = cssTpl + "\n}"

	fontFamily := t.Family
	if strings.Contains(t.Family, " ") {
		fontFamily = "'" + t.Family + "'"
	}

	if len(t.UnicodeRange) > 0 {
		return fmt.Sprintf(cssTpl, fontFamily, t.Style, t.Weight, t.URL, t.Format, strings.Join(t.UnicodeRange, ", "))
	}

	return fmt.Sprintf(cssTpl, fontFamily, t.Style, t.Weight, t.URL, t.Format)
}

// --- helpers ---

func nextSig(s *scanner.Scanner) *scanner.Token {
	for {
		t := s.Next()
		if t.Type == scanner.TokenS || t.Type == scanner.TokenComment {
			continue
		}

		return t
	}
}

func isUniqueString(sl []string, s string) bool {
	found := false
	for _, r := range sl {
		if s == r {
			found = true
			break
		}
	}
	return !found
}

func isUniqueInt(sl []int, s int) bool {
	found := false
	for _, r := range sl {
		if s == r {
			found = true
			break
		}
	}
	return !found
}