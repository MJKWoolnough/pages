package pages // import "vimagination.zapto.org/pages"

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
)

type Pages struct {
	baseTemplate    string
	isFile          bool
	defaultTemplate *template.Template

	mu        sync.RWMutex
	templates map[string]dataFile
	hook      HookFn
}

func New(baseTemplateFilename string) (*Pages, error) {
	templateSrc, err := os.ReadFile(baseTemplateFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading base template (%q): %w", baseTemplateFilename, err)
	}
	p, err := NewString(string(templateSrc))
	if err != nil {
		return nil, err
	}
	p.baseTemplate = baseTemplateFilename
	p.isFile = true
	return p, nil
}

func NewString(baseTemplate string) (*Pages, error) {
	defaultTemplate, err := template.New("").Parse(string(baseTemplate))
	if err != nil {
		return nil, fmt.Errorf("error initialising base template: %w", err)
	}
	return &Pages{
		baseTemplate:    baseTemplate,
		defaultTemplate: defaultTemplate,
		templates:       make(map[string]dataFile),
		hook:            PassthroughHook,
	}, nil
}

func (p *Pages) StaticFile(static string) error {
	return p.RegisterFile(Static, static)
}

func (p *Pages) StaticString(static string) error {
	return p.RegisterString(Static, static)
}

type dataFile struct {
	*template.Template
	data   string
	isFile bool
}

func (p *Pages) RegisterFile(name, filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.templates[name]; ok {
		return ErrTemplateExists
	}
	templateSrc, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error loading template (%q): %w", filename, err)
	}
	dtc, err := p.defaultTemplate.Clone()
	if err != nil {
		return fmt.Errorf("error cloning template (%q): %w", filename, err)
	}
	t, err := dtc.Parse(string(templateSrc))
	if err != nil {
		return fmt.Errorf("error initialising template (%q): %w", filename, err)
	}
	p.templates[name] = dataFile{
		Template: t,
		data:     filename,
		isFile:   true,
	}
	return nil
}

func (p *Pages) RegisterString(name, contents string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.templates[name]; ok {
		return ErrTemplateExists
	}
	dtc, err := p.defaultTemplate.Clone()
	if err != nil {
		return fmt.Errorf("error cloning template (%q): %w", name, err)
	}
	t, err := dtc.Parse(contents)
	if err != nil {
		return fmt.Errorf("error initialising template (%q): %w", name, err)
	}
	p.templates[name] = dataFile{
		Template: t,
		data:     contents,
	}
	return nil
}

func (p *Pages) Rebuild() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var (
		np  *Pages
		err error
	)
	if p.isFile {
		np, err = New(p.baseTemplate)
	} else {
		np = &Pages{
			defaultTemplate: p.defaultTemplate,
			templates:       make(map[string]dataFile),
		}
	}
	if err != nil {
		return fmt.Errorf("error reloading templates: %w", err)
	}
	for name, data := range p.templates {
		if data.isFile {
			if err = np.RegisterFile(name, data.data); err != nil {
				return fmt.Errorf("error reloading templates: %w", err)
			}
		} else {
			if p.isFile {
				if err = np.RegisterString(name, data.data); err != nil {
					return fmt.Errorf("error reloading templates: %w", err)
				}
			} else {
				np.templates[name] = data
			}
		}
	}
	p.defaultTemplate = np.defaultTemplate
	p.templates = np.templates
	return nil
}

func (p *Pages) Write(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) error {
	p.mu.RLock()
	tmpl, ok := p.templates[templateName]
	p.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", templateName, ErrUnknownTemplate)
	}
	if err := tmpl.Execute(w, p.hook(w, r, data)); err != nil {
		return fmt.Errorf("error writing template: %w", err)
	}
	return nil
}

func (p *Pages) Hook(hook HookFn) {
	p.mu.Lock()
	p.hook = hook
	p.mu.Unlock()
}

type HookFn func(http.ResponseWriter, *http.Request, interface{}) interface{}

var PassthroughHook HookFn = func(_ http.ResponseWriter, _ *http.Request, data interface{}) interface{} {
	return data
}

// Errors
var (
	ErrTemplateExists  = errors.New("template already exists")
	ErrUnknownTemplate = errors.New("unknown template")
)
