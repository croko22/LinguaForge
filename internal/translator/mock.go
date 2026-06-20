package translator

import (
	"context"
	"fmt"
	"strings"
)

// MockTranslator is a Translator backed by a small embedded dictionary.
// Unknown words get a "[word]" fallback so it's obvious when the mock is active.
type MockTranslator struct {
	dict map[string]string
}

// NewMockTranslator creates a Translator with a built-in en→es dictionary.
func NewMockTranslator() Translator {
	return &MockTranslator{
		dict: map[string]string{
			"hello":      "hola",
			"world":      "mundo",
			"book":       "libro",
			"chapter":    "capítulo",
			"word":       "palabra",
			"translate":  "traducir",
			"language":   "idioma",
			"read":       "leer",
			"cat":        "gato",
			"dog":        "perro",
			"house":      "casa",
			"water":      "agua",
			"food":       "comida",
			"good":       "bueno",
			"bad":        "malo",
			"big":        "grande",
			"small":      "pequeño",
			"one":        "uno",
			"two":        "dos",
			"three":      "tres",
			"economics":  "economía",
			"university": "universidad",
			"student":    "estudiante",
			"teacher":    "profesor",
			"class":      "clase",
			"study":      "estudiar",
		},
	}
}

// Translate looks up the word in the embedded dictionary or returns a bracketed fallback.
func (m *MockTranslator) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	if req.Word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	word := strings.ToLower(strings.TrimSpace(req.Word))

	// Check dictionary
	if trans, ok := m.dict[word]; ok {
		return &TranslateResponse{Translation: trans}, nil
	}

	// Fallback: echo with [translated] prefix so it's obvious it's a mock
	return &TranslateResponse{Translation: fmt.Sprintf("[%s]", word)}, nil
}
