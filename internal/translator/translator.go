package translator

import "context"

type TranslateRequest struct {
	Word       string `json:"word"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type TranslateResponse struct {
	Translation string `json:"translation"`
}

type Translator interface {
	Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error)
}
