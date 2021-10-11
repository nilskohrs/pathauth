// Package pathauth a plugin to use headers to authenticate.
package pathauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/rules"
)

// Config the plugin configuration.
type Config struct {
	Source        Source          `json:"headers,omitempty"`
	Authorization []Authorization `json:"authorization,omitempty"`
}

// Source is part of the plugin config.
type Source struct {
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Delimiter string `json:"delimiter,omitempty"`
}

// Authorization is part of the plugin config.
type Authorization struct {
	Match    string   `json:"match,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Allowed  []string `json:"allowed,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// PathAuthorization a Traefik Authorization plugin.
type PathAuthorization struct {
	next   http.Handler
	router *rules.Router
}

type roleHTTPHandler struct {
	next    http.Handler
	allowed map[string]struct{}
	*source
}

type source struct {
	sourceType sourceType
	sourceName string
	delimiter  string
}

type sourceType int8

const (
	header sourceType = iota
)

// New creates a new CookieStrip plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Source.Type != "header" && config.Source.Type != "" {
		return nil, fmt.Errorf("unknown source type")
	}
	if config.Source.Name == "" {
		return nil, fmt.Errorf("missing source name")
	}
	router, err := rules.NewRouter()
	if err != nil {
		return nil, fmt.Errorf("unable to create router")
	}

	s := source{
		sourceType: header,
		sourceName: config.Source.Name,
		delimiter:  config.Source.Delimiter,
	}

	for _, authorization := range config.Authorization {
		if len(authorization.Allowed) == 0 {
			return nil, fmt.Errorf("a authorization rule has not specified who is allowed")
		}
		handler := &roleHTTPHandler{
			next:    next,
			allowed: asMapStruct(authorization.Allowed, false),
			source:  &s,
		}

		err := router.AddRoute(authorization.Match, authorization.Priority, handler)
		if err != nil {
			return nil, fmt.Errorf("failed setting up rule: %w", err)
		}
	}
	router.SortRoutes()

	return &PathAuthorization{
		next:   next,
		router: router,
	}, nil
}

func (c *PathAuthorization) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rm := &mux.RouteMatch{}
	if c.router.Match(req, rm) {
		rm.Handler.ServeHTTP(rw, req)
	} else {
		c.next.ServeHTTP(rw, req)
	}
}

func (c *roleHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	roles := c.getRolesFromHeader(req.Header)
	if anyIn(roles, c.allowed) {
		c.next.ServeHTTP(rw, req)
	} else {
		reject(rw)
	}
}

func (c *roleHTTPHandler) getRolesFromHeader(headers http.Header) []string {
	rawRoles := headers.Values(c.sourceName)
	var roles []string
	for _, rawRole := range rawRoles {
		if c.delimiter != "" {
			roles = append(roles, strings.Split(rawRole, c.delimiter)...)
		} else {
			roles = append(roles, rawRole)
		}
	}
	return roles
}

func asMapStruct(stringSlice []string, toUpper bool) map[string]struct{} {
	set := make(map[string]struct{}, len(stringSlice))
	for _, s := range stringSlice {
		if toUpper {
			s = strings.ToUpper(s)
		}
		set[s] = struct{}{}
	}
	return set
}

func reject(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusForbidden)
	_, err := rw.Write([]byte(http.StatusText(http.StatusForbidden)))
	if err != nil {
		fmt.Printf("unexpected error while writing statuscode: %v", err)
	}
}

func anyIn(roles []string, allowed map[string]struct{}) (ok bool) {
	for _, role := range roles {
		if _, ok = allowed[role]; ok {
			return
		}
	}
	return
}
