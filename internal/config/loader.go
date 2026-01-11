package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadError represents a file loading error.
type LoadError struct {
	Path    string
	Message string
	Err     error
}

func (e *LoadError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

func (e *LoadError) Unwrap() error {
	return e.Err
}

// LoadFromFile loads Principles from a file path.
func LoadFromFile(path string) (*Principles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &LoadError{
				Path:    path,
				Message: "file not found",
				Err:     err,
			}
		}
		return nil, &LoadError{
			Path:    path,
			Message: "failed to read file",
			Err:     err,
		}
	}

	return LoadFromBytes(data, path)
}

// LoadFromBytes parses Principles from a byte slice.
func LoadFromBytes(data []byte, sourcePath string) (*Principles, error) {
	var p Principles
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, &LoadError{
			Path:    sourcePath,
			Message: "invalid YAML syntax",
			Err:     err,
		}
	}

	return &p, nil
}

// LoadOrDefault loads Principles from a file, or returns defaults if not found.
func LoadOrDefault(path string, preset Preset) (*Principles, error) {
	p, err := LoadFromFile(path)
	if err != nil {
		if le, ok := err.(*LoadError); ok && os.IsNotExist(le.Err) {
			return DefaultPrinciples(preset), nil
		}
		return nil, err
	}
	return p, nil
}

// IsLoadError checks if an error is a LoadError.
func IsLoadError(err error) bool {
	var le *LoadError
	return errors.As(err, &le)
}
