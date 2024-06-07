package pages

import (
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	Static         = "-static-"
	StaticTemplate = "{{define \"title\"}}{{.Title}}{{end}}{{define \"style\"}}{{.Style}}{{end}}{{define \"body\"}}\n{{.Body}}{{end}}"
)

type staticData struct {
	Title, Style string
	Body         template.HTML
}

type Bytes struct {
	pages      *Pages
	staticData staticData
}

func (p *Pages) Bytes(title, style string, body template.HTML) *Bytes {
	if _, ok := p.templates[Static]; !ok {
		p.StaticString(StaticTemplate)
	}

	return &Bytes{
		pages: p,
		staticData: staticData{
			Title: title,
			Style: style,
			Body:  body,
		},
	}
}

func (b *Bytes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.pages.Write(w, r, Static, &b.staticData)
}

type File struct {
	pages    *Pages
	Filename string

	mu           sync.RWMutex
	staticData   *staticData
	LastModified time.Time
}

func (p *Pages) File(title, style, filename string) *File {
	if _, ok := p.templates[Static]; !ok {
		p.StaticString(StaticTemplate)
	}

	return &File{
		pages:    p,
		Filename: filename,
		staticData: &staticData{
			Title: title,
			Style: style,
		},
	}
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats, err := os.Stat(f.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if modtime := stats.ModTime(); modtime.After(f.LastModified) {
		f.mu.Lock()

		stats, err = os.Stat(f.Filename)
		if err != nil {
			f.mu.Unlock()
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if modtime = stats.ModTime(); modtime.After(f.LastModified) { // in case another goroutine has changed it already
			if data, err := os.ReadFile(f.Filename); err != nil {
				f.mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)

				return
			} else {
				f.staticData = &staticData{
					Title: f.staticData.Title,
					Style: f.staticData.Style,
					Body:  template.HTML(data),
				}
				f.LastModified = modtime
			}
		}

		f.mu.Unlock()
	}

	f.mu.RLock()
	staticData := f.staticData
	f.mu.RUnlock()

	f.pages.Write(w, r, Static, staticData)
}
