package goswagger

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed tpl/index.html swagger/*.png swagger/*.css swagger/*.js
var fs embed.FS

var indexTpl = parseTpl("tpl/index.html")

type Config struct {
	Name         string `json:"name" flag:"swagger-name" description:"swagger name"`
	Url          string `json:"url" flag:"swagger-url"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Urls         []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"urls"`
}

func (p *Config) Validate() error {
	if p == nil {
		return nil
	}
	return nil
}

type Swagger struct {
	config *Config
}

func New(config *Config) *Swagger {
	return &Swagger{
		config: config,
	}
}

func (p Swagger) Handler() http.Handler {
	if p.config == nil {
		panic("swagger.config is nil")
	}
	return &swaggerServe{
		indexHandler:  p.apidocsHandler,
		staticHandler: http.FileServer(http.FS(fs)),
	}
}

func (p Swagger) apidocsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexTpl.Execute(w, p.config)
}

func parseTpl(file string) *template.Template {
	b, _ := fs.ReadFile(file)
	return template.Must(template.New(file).Parse(string(b)))
}

type swaggerServe struct {
	staticHandler http.Handler
	indexHandler  func(http.ResponseWriter, *http.Request)
}

func (s swaggerServe) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	switch path {
	case "/swagger", "/swagger/", "/swagger/index.html":
		s.indexHandler(w, req)
	default:
		s.staticHandler.ServeHTTP(w, req)
	}
}
