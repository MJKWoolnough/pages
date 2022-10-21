# pages
--
    import "vimagination.zapto.org/pages"


## Usage

```go
const (
	Static         = "-static-"
	StaticTemplate = "{{define \"title\"}}{{.Title}}{{end}}{{define \"style\"}}{{.Style}}{{end}}{{define \"body\"}}\n{{.Body}}{{end}}"
)
```

```go
var (
	ErrTemplateExists  = errors.New("template already exists")
	ErrUnknownTemplate = errors.New("unknown template")
)
```
Errors

#### type Bytes

```go
type Bytes struct {
}
```


#### func (*Bytes) ServeHTTP

```go
func (b *Bytes) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

#### type File

```go
type File struct {
	Filename string

	LastModified time.Time
}
```


#### func (*File) ServeHTTP

```go
func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

#### type HookFn

```go
type HookFn func(http.ResponseWriter, *http.Request, interface{}) interface{}
```


```go
var PassthroughHook HookFn = func(_ http.ResponseWriter, _ *http.Request, data interface{}) interface{} {
	return data
}
```

#### type Pages

```go
type Pages struct {
}
```


#### func  New

```go
func New(baseTemplateFilename string) (*Pages, error)
```

#### func  NewString

```go
func NewString(baseTemplate string) (*Pages, error)
```

#### func (*Pages) Bytes

```go
func (p *Pages) Bytes(title, style string, body template.HTML) *Bytes
```

#### func (*Pages) File

```go
func (p *Pages) File(title, style, filename string) *File
```

#### func (*Pages) Hook

```go
func (p *Pages) Hook(hook HookFn)
```

#### func (*Pages) Rebuild

```go
func (p *Pages) Rebuild() error
```

#### func (*Pages) RegisterFile

```go
func (p *Pages) RegisterFile(name, filename string) error
```

#### func (*Pages) RegisterString

```go
func (p *Pages) RegisterString(name, contents string) error
```

#### func (*Pages) StaticFile

```go
func (p *Pages) StaticFile(static string) error
```

#### func (*Pages) StaticString

```go
func (p *Pages) StaticString(static string) error
```

#### func (*Pages) Write

```go
func (p *Pages) Write(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) error
```
