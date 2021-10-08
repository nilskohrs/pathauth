// Package pathauth a plugin to use headers to authenticate.
package pathauth

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
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
	Path     string   `json:"path,omitempty"`
	Host     string   `json:"host,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Allowed  []string `json:"allowed,omitempty"`
	Method   []string `json:"method,omitempty"`
}

type rule struct {
	path     *regexp.Regexp
	host     *regexp.Regexp
	allowed  map[string]struct{}
	priority int
	method   map[string]struct{}
}

type sourceType int8

const (
	header sourceType = iota
)

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// PathAuthorization a Traefik Authorization plugin.
type PathAuthorization struct {
	next       http.Handler
	name       string
	sourceType sourceType
	sourceName string
	delimiter  string
	rules      []rule
}

// New creates a new CookieStrip plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Source.Type != "header" && config.Source.Type != "" {
		return nil, fmt.Errorf("unknown source type")
	}
	if config.Source.Name == "" {
		return nil, fmt.Errorf("missing source name")
	}

	plugin := &PathAuthorization{
		sourceType: header,
		sourceName: config.Source.Name,
		delimiter:  config.Source.Delimiter,
		next:       next,
		name:       name,
		rules:      []rule{},
	}

	for _, authorization := range config.Authorization {
		if authorization.Path == "" {
			return nil, fmt.Errorf("a authorization rule is missing a path")
		}
		if len(authorization.Allowed) == 0 {
			return nil, fmt.Errorf("a authorization rule has not specified who is allowed")
		}
		var host *regexp.Regexp
		if authorization.Host != "" {
			host = regexp.MustCompile(authorization.Host)
		}
		plugin.rules = append(plugin.rules, rule{
			path:     regexp.MustCompile(authorization.Path),
			host:     host,
			allowed:  asMapStruct(authorization.Allowed, false),
			priority: authorization.Priority,
			method:   asMapStruct(authorization.Method, true),
		})
	}

	sort.SliceStable(plugin.rules, func(i, j int) bool {
		return plugin.rules[i].priority < plugin.rules[j].priority
	})

	return plugin, nil
}

func (c *PathAuthorization) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	roles := c.getRolesFromHeader(req.Header)
	for _, rule := range c.rules {
		if _, ok := rule.method[req.Method]; (len(rule.method) == 0 || ok) && rule.path.MatchString(req.URL.Path) && (rule.host == nil || rule.host.MatchString(req.URL.Hostname())) {
			if !anyIn(roles, rule.allowed) {
				reject(rw)
				return
			}
			break
		}
	}
	c.next.ServeHTTP(rw, req)
}

func reject(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusForbidden)
	_, err := rw.Write([]byte(http.StatusText(http.StatusForbidden)))
	if err != nil {
		fmt.Printf("unexpected error while writing statuscode: %v", err)
	}
}

func anyIn(roles []string, allowed map[string]struct{}) bool {
	for _, role := range roles {
		if _, ok := allowed[role]; ok {
			return true
		}
	}
	return false
}

func (c *PathAuthorization) getRolesFromHeader(headers http.Header) []string {
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
