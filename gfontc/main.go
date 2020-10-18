package main

import (
	"net/url"
	"fmt"
	"flag"
	"os"
	"strings"
	"io/ioutil"
	"encoding/json"

	"github.com/imacks/gfont"
)

var (
	cmdlet string
	infile string
	outfile string
	fontFamily string
	fontStyle string
	fontProfile string
	filterField string
	mirrorProxy string
	pretty bool
	verbose bool
	compatMode bool
)

const (
	appName = "gfontc"
	appVer = "1.0.0"
	appDesc = "Get useful info from Google Fonts"
)

var fpArgmap = map[string]gfont.FontProfile{
	"woff2": gfont.WOFF2,
	"apple_woff2": gfont.AppleWOFF2,
	"legacy_woff2": gfont.LegacyWOFF2,
	"woff": gfont.WOFF,
	"apple_woff": gfont.AppleWOFF,
	"legacy_woff": gfont.LegacyWOFF,
	"ttf": gfont.TTF,
	"apple_ttf": gfont.AppleTTF,
	"svg": gfont.SVG,
	"eot": gfont.EOT,
}

func init() {
	dlFlagSet := flag.NewFlagSet("download", flag.ExitOnError)
	dlFlagSet.StringVar(&outfile, "o", "-", "Output to file or stdout")
	dlFlagSet.StringVar(&fontFamily, "t", "", "Font name (mandatory)")
	dlFlagSet.StringVar(&fontStyle, "s", "", "Font style params")
	dlFlagSet.StringVar(&fontProfile, "p", "woff2", "Font profile (see notes)")
	dlFlagSet.StringVar(&mirrorProxy, "m", "", "Mirror proxy")
	dlFlagSet.BoolVar(&verbose, "v", false, "Verbose mode")
	dlFlagSet.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintf(os.Stdout, "download font-face CSS from Google API\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Usage: %s download -t <family> -s <style> [-p <profile>] [-o <file.css>] [-m <url>] [-v]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		dlFlagSet.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Supported font profiles:\n")
		fmt.Fprintf(os.Stdout, "    woff2 | apple_woff2 | legacy_woff2\n")
		fmt.Fprintf(os.Stdout, "    woff  | apple_woff  | legacy_woff\n")
		fmt.Fprintf(os.Stdout, "    ttf   | apple_ttf\n")
		fmt.Fprintf(os.Stdout, "    svg   | eot\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Example:\n")
		fmt.Fprintf(os.Stdout, "    %s download -t Domine -s 'wght@400;500;600;700' -p woff2 -o font.css\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
	}

	parseFlagSet := flag.NewFlagSet("parse", flag.ExitOnError)
	parseFlagSet.StringVar(&outfile, "o", "-", "Output to file or stdout")
	parseFlagSet.StringVar(&infile, "i", "", "Input file (mandatory)")
	parseFlagSet.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintf(os.Stdout, "parse font-face CSS served by Google API into JSON\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Usage: %s parse -i <file.css> [-o <file.json>]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		parseFlagSet.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Example:\n")
		fmt.Fprintf(os.Stdout, "    %s parse -i font.css -o font.json\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
	}

	filterFlagSet := flag.NewFlagSet("filter", flag.ExitOnError)
	filterFlagSet.StringVar(&infile, "i", "", "Input file (mandatory)")
	filterFlagSet.StringVar(&filterField, "q", "", "Filter field (mandatory)")
	filterFlagSet.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintf(os.Stdout, "query fonts data in JSON format by font properties\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Usage: %s filter -i <file.json> -q <field>\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		filterFlagSet.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Supported query fields:\n")
		fmt.Fprintf(os.Stdout, "    url | family | format\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Example:\n")
		fmt.Fprintf(os.Stdout, "    %s filter -i font.json -q url\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
	}

	mergeFlagSet := flag.NewFlagSet("merge", flag.ExitOnError)
	mergeFlagSet.StringVar(&outfile, "o", "-", "Output to file or stdout")
	mergeFlagSet.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintf(os.Stdout, "combine several fonts data in JSON format\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Usage: %s merge [-o <file.css>] <file1.json> [<file2.json>...]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		mergeFlagSet.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Example:\n")
		fmt.Fprintf(os.Stdout, "    %s merge -o all.json font1.json font2.json\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
	}

	renderFlagSet := flag.NewFlagSet("css", flag.ExitOnError)
	renderFlagSet.StringVar(&infile, "i", "", "Input file (mandatory)")
	renderFlagSet.StringVar(&outfile, "o", "-", "Output to file or stdout")
	renderFlagSet.BoolVar(&pretty, "H", false, "Human readable")
	renderFlagSet.BoolVar(&compatMode, "c", false, "Max legacy compatibility")
	renderFlagSet.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintf(os.Stdout, "create CSS from fonts data in JSON format\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Usage: %s css -i <file.json> [-o <file.css>] [-c] [-H]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		renderFlagSet.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "Example:\n")
		fmt.Fprintf(os.Stdout, "    %s css -i all.json -o all.css -c -H\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
	}

	// gfont download -t Domine -s 'wght@400;500;600;700' | gfont parse -i -
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", appName, appVer, appDesc)
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintf(os.Stdout, "Usage: %s download -t <family> -s <style> [-p <format>] [-o <file.css>] [-v]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s parse -i <file.css> [-o <file.json>]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s filter -i <file.json> -q <field>\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s merge [-o <file.css>] <file1.json> [<file2.json>...]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s css -i <file.json> [-o <file.css>] [-c] [-H]\n", os.Args[0])
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "To view parameters for each subcommand:")
		fmt.Fprintf(os.Stdout, "    %s -h <subcommand>\n", os.Args[0])
		fmt.Fprintln(os.Stdout, "")
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "expect subcommand!\n")
		os.Exit(1)
	}

	cmdlet = os.Args[1]
	switch os.Args[1] {
	case "download":
		if err := dlFlagSet.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "subcommand %s: %v\n", cmdlet, err)
			os.Exit(1)
		}
		if _, ok := fpArgmap[fontProfile]; !ok {
			fmt.Fprintf(os.Stderr, "subcommand %s: unsupported font profile\n", cmdlet)
			os.Exit(1)
		}
		if fontFamily == "" {
			fmt.Fprintf(os.Stderr, "subcommand %s: -t <font> mandatory\n", cmdlet)
			os.Exit(1)
		}
	case "parse":
		if err := parseFlagSet.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "subcommand %s: %v\n", cmdlet, err)
			os.Exit(1)
		}
		if infile == "" {
			fmt.Fprintf(os.Stderr, "subcommand %s: -i <file> mandatory\n", cmdlet)
			os.Exit(1)
		}
	case "filter":
		if err := filterFlagSet.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "subcommand %s: %v\n", cmdlet, err)
			os.Exit(1)
		}
		if infile == "" {
			fmt.Fprintf(os.Stderr, "subcommand %s: -i <file> mandatory\n", cmdlet)
			os.Exit(1)
		}
		if filterField == "" {
			fmt.Fprintf(os.Stderr, "subcommand %s: -q <field> mandatory\n", cmdlet)
			os.Exit(1)
		}
	case "merge":
		if err := mergeFlagSet.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "subcommand %s: %v\n", cmdlet, err)
			os.Exit(1)
		}
		infile = strings.Join(mergeFlagSet.Args(), ",")
	case "css":
		if err := renderFlagSet.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "subcommand %s: %v\n", cmdlet, err)
			os.Exit(1)
		}
		if infile == "" {
			fmt.Fprintf(os.Stderr, "subcommand %s: -i <file> mandatory\n", cmdlet)
			os.Exit(1)
		}
	default:
		if cmdlet == "-h" || cmdlet == "--help" {
			if len(os.Args) < 3 {
				flag.Usage()
				os.Exit(0)
			}

			subtopic := os.Args[2]
			switch subtopic {
			case "download": dlFlagSet.Usage()
			case "parse":    parseFlagSet.Usage()
			case "filter":   filterFlagSet.Usage()
			case "merge":    mergeFlagSet.Usage()
			case "css":      renderFlagSet.Usage()
			default:
				fmt.Fprintf(os.Stderr, "invalid help topic: %s\n", subtopic)
				flag.Usage()
				os.Exit(1)
			}
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "invalid subcommand: %s\n", cmdlet)
		os.Exit(1)
	}
}

func readFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path not specified")
	}

	if path == "-" {
		result, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	result, errRead := ioutil.ReadFile(path)
	if errRead != nil {
		return nil, errRead
	}

	return result, nil
}

func writeFile(content []byte, outPath string) error {
	if outPath == "-" || outPath == "" {
		fmt.Printf(string(content))
		return nil
	}

	errWrite := ioutil.WriteFile(outPath, content, 0644)
	if errWrite != nil {
		return errWrite
	}

	return nil
}

func main() {
	switch cmdlet {
	case "download":
		var mir *url.URL
		if mirrorProxy != "" {
			var errURL error
			mir, errURL = url.Parse(mirrorProxy)
			if errURL != nil {
				panic(errURL)
			}
		}
		cssURL := gfont.GetURL(fpArgmap[fontProfile], fontFamily, fontStyle, mir)
		if verbose {
			fmt.Printf("[INFO] HTTP:GET %s\n", cssURL)
		}
		cssBytes, err := gfont.DownloadCSS(fpArgmap[fontProfile], fontFamily, fontStyle, mir)
		if err != nil {
			panic(err)
		}
		err = writeFile(cssBytes, outfile)
		if err != nil {
			panic(err)
		}
	case "parse":
		cssBytes, err := readFile(infile)
		if err != nil {
			panic(err)
		}

		typefaces := gfont.Typefaces{}
		err = gfont.UnmarshalCSS(cssBytes, &typefaces)
		if err != nil {
			panic(err)
		}

		jsonBytes, errJSON := json.Marshal(typefaces)
		if errJSON != nil {
			panic(errJSON)
		}

		err = writeFile(jsonBytes, outfile)
		if err != nil {
			panic(err)
		}
	case "filter":
		jsonBytes, err := readFile(infile)
		if err != nil {
			panic(err)
		}

		var typefaces gfont.Typefaces
		err = json.Unmarshal(jsonBytes, &typefaces)
		if err != nil {
			panic(err)
		}

		if filterField == "url" {
			filtered := typefaces.URL()
			for _, v := range filtered {
				fmt.Printf("%s\n", v.String())
			}
		} else if filterField == "family" {
			filtered := typefaces.Family()
			for _, v := range filtered {
				fmt.Printf("%s\n", v)
			}
		} else if filterField == "format" {
			filtered := typefaces.Format()
			for _, v := range filtered {
				fmt.Printf("%s\n", v)
			}
		} else {
			panic(fmt.Errorf("unsupported field %s", filterField))
		}
	case "merge":
		allInFiles := strings.Split(infile, ",")
		allTypefaces := gfont.Typefaces{}

		for _, v := range allInFiles {
			jsonBytes, err := readFile(v)
			if err != nil {
				panic(err)
			}

			var typefaces gfont.Typefaces
			err = json.Unmarshal(jsonBytes, &typefaces)
			if err != nil {
				panic(err)
			}

			allTypefaces.Fonts = append(allTypefaces.Fonts, typefaces.Fonts...)
		}

		jsonBytes, errJSON := json.Marshal(allTypefaces)
		if errJSON != nil {
			panic(errJSON)
		}

		err := writeFile(jsonBytes, outfile)
		if err != nil {
			panic(err)
		}
	case "css":
		jsonBytes, err := readFile(infile)
		if err != nil {
			panic(err)
		}

		var typefaces gfont.Typefaces
		err = json.Unmarshal(jsonBytes, &typefaces)
		if err != nil {
			panic(err)
		}

		var result string
		if compatMode {
			if pretty {
				result = typefaces.PrettyCSS()
			} else {
				result = typefaces.CSS()
			}
		} else {
			r := make([]string, len(typefaces.Fonts))
			for i, v := range typefaces.Fonts {
				if pretty {
					r[i] = v.PrettyCSS()
				} else {
					r[i] = v.CSS()
				}
			}

			sep := ""
			if pretty {
				sep = "\n"
			}
			result = strings.Join(r, sep)
		}

		err = writeFile([]byte(result), outfile)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unexpected fallthrough"))
	}
}