//go:generate sh -c "go-bindata -pkg goswagger -o resources.go static/... tpl/..."

package goswagger

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/go-openapi/spec"
	"github.com/yubo/golib/openapi"
)

var indexTpl = parseTpl("tpl/index.html.tpl")

type SchemeConfig struct {
	Name             string               `json:"name"`
	Type             openapi.SecurityType `json:"type" description:"base|apiKey|implicit|password|application|accessCode"`
	FieldName        string               `json:"fieldName" description:"used for apiKey"`
	ValueSource      string               `json:"valueSource" description:"used for apiKey, header|query|cookie"`
	AuthorizationURL string               `json:"authorizationURL" description:"used for OAuth2"`
	TokenURL         string               `json:"tokenURL" description:"used for OAuth2"`
	scheme           *spec.SecurityScheme `json:"-"`
}

func (p *SchemeConfig) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name must be set")
	}
	switch p.Type {
	case openapi.SecurityTypeBase:
		p.scheme = spec.BasicAuth()
	case openapi.SecurityTypeApiKey:
		if p.FieldName == "" {
			return fmt.Errorf("fieldName must be set for %s", p.Type)
		}
		if p.ValueSource == "" {
			return fmt.Errorf("valueSource must be set for %s", p.Type)
		}
		p.scheme = spec.APIKeyAuth(p.FieldName, p.ValueSource)
	case openapi.SecurityTypeImplicit:
		if p.AuthorizationURL == "" {
			return fmt.Errorf("authorizationURL must be set for %s", p.Type)
		}
		p.scheme = spec.OAuth2Implicit(p.AuthorizationURL)
	case openapi.SecurityTypePassword:
		if p.TokenURL == "" {
			return fmt.Errorf("tokenURL must be set for %s", p.Type)
		}
		p.scheme = spec.OAuth2Password(p.TokenURL)
	case openapi.SecurityTypeApplication:
		if p.TokenURL == "" {
			return fmt.Errorf("tokenURL must be set for %s", p.Type)
		}
		p.scheme = spec.OAuth2Application(p.TokenURL)
	case openapi.SecurityTypeAccessCode:
		if p.TokenURL == "" {
			return fmt.Errorf("tokenURL must be set for %s", p.Type)
		}
		if p.AuthorizationURL == "" {
			return fmt.Errorf("authorizationURL must be set for %s", p.Type)
		}
		p.scheme = spec.OAuth2AccessToken(p.AuthorizationURL, p.TokenURL)
	default:
		return fmt.Errorf("scheme.type %s is invalid, should be one of %s", p.Type,
			strings.Join([]string{
				string(openapi.SecurityTypeBase),
				string(openapi.SecurityTypeApiKey),
				string(openapi.SecurityTypeImplicit),
				string(openapi.SecurityTypePassword),
				string(openapi.SecurityTypeApplication),
				string(openapi.SecurityTypeAccessCode),
			}, ", "))
	}
	return nil
}

type Config struct {
	Enabled      bool           `json:"enabled"`
	Name         string         `json:"name"`
	Url          string         `json:"url"`
	ClientId     string         `json:"clientId"`
	ClientSecret string         `json:"clientSecret"`
	Schemes      []SchemeConfig `json:"schemes"`
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

func (p Swagger) Install(mux mux) error {
	staticHandler := http.FileServer(&assetfs.AssetFS{
		Asset:    Asset,
		AssetDir: AssetDir,
	})

	mux.HandleFunc("/api", redirectTo("/api/"))
	mux.HandleFunc("/api/", p.apidocsHandler)
	mux.HandleFunc("/oauth2-redirect.html", fileHandler("static/oauth2-redirect.html"))
	mux.Handle("/static/", staticHandler)

	// register scheme to openapi
	for _, v := range p.config.Schemes {
		if err := v.Validate(); err != nil {
			return err
		}
		if err := openapi.SecuritySchemeRegister(v.Name, v.scheme); err != nil {
			return err
		}
	}
	return nil
}

func (p Swagger) Schemes() []SchemeConfig {
	return p.config.Schemes
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
