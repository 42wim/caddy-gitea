package gitea

import (
	"io"
	"net/http"
	"strings"

	"github.com/42wim/caddy-gitea/pkg/gitea"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("gitea", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Middleware implements gitea plugin.
type Middleware struct {
	Client     *gitea.Client `json:"-"`
	Server     string        `json:"server,omitempty"`
	Token      string        `json:"token,omitempty"`
	GiteaPages string        `json:"gitea_pages,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.gitea",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision provisions gitea client.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.Client = gitea.NewClient(m.Server, m.Token, m.GiteaPages)

	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	return nil
}

// UnmarshalCaddyfile unmarshals a Caddyfile.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	for d.Next() {
		for n := d.Nesting(); d.NextBlock(n); {
			switch d.Val() {
			case "server":
				d.Args(&m.Server)
			case "token":
				d.Args(&m.Token)
			case "gitea_pages":
				d.Args(&m.GiteaPages)
			}
		}
	}

	return nil
}

// ServeHTTP performs gitea content fetcher.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, _ caddyhttp.Handler) error {
	h := strings.Split(r.Host, ".")
	fp := h[0] + r.URL.Path

	f, err := m.Client.Open(fp, r.URL.Query().Get("ref"))
	if err != nil {
		return caddyhttp.Error(http.StatusNotFound, err)
	}

	_, err = io.Copy(w, f)
	return err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
