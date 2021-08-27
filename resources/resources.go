package resources

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//go:embed assets/*
//go:embed js/*
//go:embed css/*
var assetsFS embed.FS

//go:embed tmpl/*
var tmplFS embed.FS

func loadTmplFS() *template.Template {
	tmpl := template.New("")

	fs.WalkDir(tmplFS, "tmpl", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if d.IsDir() {
			return nil
		}
		log.Infof("Loading template: %s", path)
		template.Must(tmpl.New(path).ParseFS(tmplFS, path))

		return nil
	})
	return tmpl
}

func Init(r *gin.Engine) {
	// templates files
	tmpl := loadTmplFS()
	r.SetHTMLTemplate(tmpl)

	// static files
	r.StaticFS("/public", http.FS(assetsFS))
}
