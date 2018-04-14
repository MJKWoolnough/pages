package pages

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/MJKWoolnough/errors"
)

type Pages struct {
	baseTemplate    string
	isFile          bool
	defaultTemplate *template.Template

	mu        sync.RWMutex
	templates map[string]dataFile
}

func New(baseTemplateFilename string) (*Pages, error) {
	templateSrc, err := ioutil.ReadFile(baseTemplateFilename)
	if err != nil {
		return nil, errors.WithContext(fmt.Sprintf("error loading base template (%q): ", baseTemplateFilename), err)
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
		return nil, errors.WithContext("error initialising base template: ", err)
	}
	return &Pages{
		baseTemplate:    baseTemplate,
		defaultTemplate: defaultTemplate,
		templates:       make(map[string]dataFile),
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
	templateSrc, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading template (%q): ", filename), err)
	}
	dtc, err := p.defaultTemplate.Clone()
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error cloning template (%q): ", filename), err)
	}
	t, err := dtc.Parse(string(templateSrc))
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error initialising template (%q): ", filename), err)
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
		return errors.WithContext(fmt.Sprintf("error cloning template (%q): ", name), err)
	}
	t, err := dtc.Parse(contents)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error initialising template (%q): ", name), err)
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
		return errors.WithContext("error reloading templates: ", err)
	}
	for name, data := range p.templates {
		if data.isFile {
			if err = np.RegisterFile(name, data.data); err != nil {
				return errors.WithContext("error reloading templates: ", err)
			}
		} else {
			if p.isFile {
				if err = np.RegisterString(name, data.data); err != nil {
					return errors.WithContext("error reloading templates: ", err)
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
		return errors.WithContext(templateName, ErrUnknownTemplate)
	}
	if err := tmpl.Execute(w, data); err != nil {
		return errors.WithContext("error writing template: ", err)
	}
	return nil
}

const (
	ErrTemplateExists  errors.Error = "template already exists"
	ErrUnknownTemplate errors.Error = "unknown template"
)
