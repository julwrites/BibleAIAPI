package bible

// Version represents a Bible version configuration.
type Version struct {
	Code      string            `json:"code" yaml:"code"`
	Name      string            `json:"name" yaml:"name"`
	Language  string            `json:"language" yaml:"language"`
	Providers map[string]string `json:"providers" yaml:"providers"`
}
