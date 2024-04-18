package assets

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed files/*.tpl
var tmplFS embed.FS

func Render(w io.Writer, name string, data any) error {
	if !strings.HasSuffix(name, ".tpl") {
		name = fmt.Sprintf("%s.tpl", name)
	}

	eng := template.Must(template.New("").ParseFS(tmplFS, "files/*.tpl"))

	tmpl := template.Must(eng.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, "files/"+name))
	return tmpl.ExecuteTemplate(w, name, data)
}

func Export(target string, dat []byte) error {
	err := os.MkdirAll(filepath.Dir(target), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(target, dat, 0666)
}
