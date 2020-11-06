//go:generate sh -c "go-bindata -pkg goswagger -o resources.go static/... tpl/..."

package goswagger

import (
	"html/template"
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
)

var indexTpl = parseTpl("tpl/index.html.tpl")

type Config struct {
	Enabled      bool   `json:"enabled"`
	Name         string `json:"name"`
	Url          string `json:"url"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Urls         []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"urls"`
}

type Swagger struct {
	config *Config
}

func New(config *Config) *Swagger {
	return &Swagger{
		config: config,
	}
}

type mux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

func (p Swagger) Install(mux mux) {
	staticHandler := http.FileServer(&assetfs.AssetFS{
		Asset:    Asset,
		AssetDir: AssetDir,
	})

	mux.HandleFunc("/api", redirectTo("/api/"))
	mux.HandleFunc("/api/", p.apidocsHandler)
	mux.HandleFunc("/oauth2-redirect.html", fileHandler("static/oauth2-redirect.html"))
	mux.Handle("/static/", staticHandler)
}

// redirectTo redirects request to a certain destination.
func redirectTo(to string) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, to, http.StatusFound)
	}
}

func (p Swagger) apidocsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexTpl.Execute(w, p.config)
}

func fileHandler(filename string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		b, _ := Asset(filename)
		w.Write(b)
	}
}

func parseTpl(file string) *template.Template {
	b, _ := Asset(file)
	return template.Must(template.New(file).Parse(string(b)))
}
