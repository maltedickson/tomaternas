package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

// iconSprite is the hidden SVG block injected once per page.
var iconSprite template.HTML

// iconVariants defines each icon family: where the SVG files live,
// the ID prefix to use in the sprite, and the SVG presentation attributes
// to apply to each <symbol>. Attributes are hardcoded here rather than
// parsed from individual files because Tabler is consistent across all icons.
//
// stroke-width uses a CSS custom property (var(--icon-stroke-width,2)) so
// it can be overridden from CSS
var iconVariants = []struct {
	dir   string
	id    string
	attrs string
}{
	{
		dir:   filepath.Join("icons", "outline"),
		id:    "icon-outline",
		attrs: `fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="var(--icon-stroke-width,2)"`,
	},
	{
		dir:   filepath.Join("icons", "filled"),
		id:    "icon-filled",
		attrs: `fill="currentColor"`,
	},
}

func init() {
	iconSprite = buildIconSprite(IconFiles)
}

func buildIconSprite(fsys fs.ReadFileFS) template.HTML {
	var buf bytes.Buffer
	buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" style="display:none" aria-hidden="true">`)

	for _, v := range iconVariants {
		entries, err := fs.ReadDir(fsys.(fs.ReadDirFS), v.dir)
		if err != nil {
			continue // folder doesn't exist yet — not an error
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".svg") {
				continue
			}
			data, err := fsys.ReadFile(filepath.Join(v.dir, e.Name()))
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".svg")
			fmt.Fprintf(&buf,
				`<symbol id="%s-%s" viewBox="0 0 24 24" %s>%s</symbol>`,
				v.id, name, v.attrs, svgInner(string(data)),
			)
		}
	}

	buf.WriteString(`</svg>`)
	return template.HTML(buf.String())
}

// svgInner extracts the child elements from a Tabler SVG source file,
// stripping the outer <svg> wrapper and any leading HTML comments.
func svgInner(src string) string {
	svgStart := strings.Index(src, "<svg")
	if svgStart < 0 {
		return src
	}
	tagEnd := strings.Index(src[svgStart:], ">")
	if tagEnd < 0 {
		return src
	}
	contentStart := svgStart + tagEnd + 1
	end := strings.LastIndex(src, "</svg>")
	if end < 0 || end < contentStart {
		return src
	}
	return strings.TrimSpace(src[contentStart:end])
}

// IconFuncMap returns the template functions for icon rendering.
func IconFuncMap() template.FuncMap {
	return template.FuncMap{
		// icons-sprite outputs the hidden SVG definitions block.
		// Call it once at the top of <body> in your base layout.
		"iconsSprite": func() template.HTML {
			return iconSprite
		},

		// icon renders a single icon as a <use> reference.
		//
		//   variant  "outline" or "filled"
		//   name     icon filename without .svg  (e.g. "clock", "star-filled")
		//   extra    optional additional CSS classes
		//
		// Examples:
		//   {{ icon "outline" "clock" }}
		//   {{ icon "filled"  "star"  "icon--lg" }}
		"icon": func(variant, name string, extra ...string) template.HTML {
			classAttribute := "icon"
			if len(extra) > 0 {
				classAttribute += " " + strings.Join(extra, " ")
			}
			return template.HTML(fmt.Sprintf(
				`<svg class="%s" aria-hidden="true" focusable="false"><use href="#icon-%s-%s"></use></svg>`,
				classAttribute, variant, name,
			))
		},
	}
}
