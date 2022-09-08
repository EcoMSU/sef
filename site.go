package sef

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Page interface {
	Build() (io.Reader, error)
	Handle(http.ResponseWriter, *http.Request)
	Watch() func()
}

type SiteData struct {
	Title string
	Desc  string
	URL   string
}

func (sd *SiteData) AddTitle(s string) {
	sd.Title += s
}

type Site interface {
	Run(int) error
	Build(string) error
	SetStatic([][3]string)
	GetData() SiteData
	AddPage(string, Page)
	AddAlias(string, string)
	ServeTo(*http.ServeMux)
}

type site struct {
	server   *http.ServeMux
	pages    map[string]Page
	patterns []string
	alias    map[string]string
	static   [][3]string
	SiteData
}

func NewSite(sd SiteData) (s *site) {
	s = new(site)
	s.server = http.DefaultServeMux
	s.pages = make(map[string]Page)
	s.alias = make(map[string]string)
	s.SiteData = sd
	return s
}

func (s site) GetData() SiteData {
	return s.SiteData
}

func (s *site) Run(port int) error {
	if port == 0 {
		port = 8080
	}
	return http.ListenAndServe(":"+strconv.Itoa(port), s.server)
}

func (s *site) Build(buildpath string) (err error) {
	var html *os.File
	var data io.Reader
	for _, p := range s.patterns {
		if data, err = s.pages[p].Build(); err != nil {
			return
		}
		if html, err = os.Create(path.Join(buildpath, p)); err != nil {
			return
		}
		if _, err = io.Copy(html, data); err != nil {
			html.Close()
			return
		}
		html.Close()
	}
	return
}

// patterns [3]string = pattern string, cut string, realpath string
func (s *site) SetStatic(patterns [][3]string) {
	s.static = patterns
}

func (s *site) AddPage(pattern string, p Page) {
	s.patterns = append(s.patterns, pattern)
	s.pages[pattern] = p
}

func (s *site) AddAlias(alias string, pattern string) {
	s.alias[alias] = pattern
}

func (s *site) ServeTo(sm *http.ServeMux) {
	s.server = sm
	for _, p := range s.static {
		sm.Handle(p[0], http.StripPrefix(p[1],
			http.FileServer(http.Dir(p[2])),
		))
	}
	for _, p := range s.patterns {
		sm.HandleFunc("/"+p, s.pages[p].Handle)
	}
	for a, p := range s.alias {
		sm.HandleFunc("/"+a, s.pages[p].Handle)
	}
}
