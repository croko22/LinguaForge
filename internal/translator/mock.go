package translator

import (
	"context"
	"fmt"
	"strings"
)

type mockTranslator struct {
	// Simple embedded dictionary for realistic-looking translations
	dict map[string]string
}

func NewMockTranslator() Translator {
	return &mockTranslator{
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

func (m *mockTranslator) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
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
