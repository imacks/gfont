gfont
=====
Get information about any Google Font. No API key required.


Quick start
-----------
Google serves different CSS based on your useragent string. WOFF2, WOFF and TTF have Apple variants. WOFF2 and WOFF also have 
variants that support unicode range.

Use `gfont.DownloadCSS` to get the CSS for any variant:

```golang
cssBytes, _ := gfont.DownloadCSS(gfont.AppleWOFF2, "Domine", "wght@400;500;600;700", nil)
fmt.Printf("%s\n", string(cssBytes))
```

You can get all the styles by going to Google Fonts website, and select all the styles. Note the embed URL.

Next, parse the CSS into a collection of font objects:

```golang
var typefaces gfont.Typefaces
_ := gfont.UnmarshalCSS(cssBytes, &typefaces)
```

Each font object has its properties all parsed out. It can also be converted back to CSS:

```golang
fmt.Printf("format: %s\n", typefaces.Fonts[0].Format)
fmt.Printf("%s\n", typefaces.Fonts[0].PrettyCSS())
```

You can select fonts from a collection:

```golang
ts := typefaces.Select("", "Domine", "normal", 400)
for _, v := range ts {
    fmt.Printf("%s: %s\n", v.Format, v.URL.String())
}
```

Generate CSS from a collection of fonts:

```golang
fmt.Printf("%s\n", typefaces.PrettyCSS())
```

CLI utility
-----------
If you need to use the above functionality on the commandline, check out the `gfontc` subfolder.
