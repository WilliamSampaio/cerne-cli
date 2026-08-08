package localization

import "fmt"

type Language string

const (
	English          Language = "en"
	PortugueseBrazil Language = "pt-BR"
	Default                   = PortugueseBrazil
)

type InvalidLanguageError struct{ Value string }

func (err InvalidLanguageError) Error() string {
	return fmt.Sprintf("unsupported language %q", err.Value)
}

func Parse(value string) (Language, error) {
	switch Language(value) {
	case English, PortugueseBrazil:
		return Language(value), nil
	default:
		return "", InvalidLanguageError{Value: value}
	}
}

func Resolve(flag, environment string, environmentSet bool, home string) (Language, error) {
	if flag != "" {
		return Parse(flag)
	}
	if environmentSet {
		return Parse(environment)
	}
	if home == "" {
		return Default, nil
	}
	config, present, err := Load(home)
	if err != nil {
		return "", err
	}
	if present {
		return config.Language, nil
	}
	return Default, nil
}
