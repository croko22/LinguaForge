package parser

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/croko/language-app/internal/model"
)

// PdfParser parses PDF files using the ledongthuc/pdf library.
type PdfParser struct{}

// NewPdfParser creates a new PdfParser.
func NewPdfParser() *PdfParser {
	return &PdfParser{}
}

// CanParse returns true if the filename has a .pdf extension (case-insensitive).
func (p *PdfParser) CanParse(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".pdf")
}

// Parse reads a PDF file and returns a ParsedDocument with chapters.
//
// Chapter detection strategy:
//  1. If the PDF has an outline (bookmark tree), outline titles are used as
//     chapter names and mapped 1:1 to pages by index.
//  2. Otherwise, one chapter per page is created with "Page N" titles.
//  3. Metadata (title, author) is extracted from the PDF trailer Info dict.
func (p *PdfParser) Parse(readerAt ReaderAt, size int64) (*model.ParsedDocument, error) {
	r, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("pdf: open: %w", err)
	}

	doc := &model.ParsedDocument{}
	extractPDFMetadata(r, doc)

	numPages := r.NumPage()

	// Collect outline titles if available.
	outlineTitles := collectOutlineTitles(r.Outline())

	// Build chapters — one page per chapter.
	chapters := make([]model.ParsedChapter, 0, numPages)
	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			text = ""
		}

		title := fmt.Sprintf("Page %d", i)
		if i-1 < len(outlineTitles) {
			title = outlineTitles[i-1]
		}

		chapters = append(chapters, model.ParsedChapter{
			Index:   len(chapters),
			Title:   title,
			Content: strings.TrimSpace(text),
		})
	}

	// Fallback: ensure at least one chapter.
	if len(chapters) == 0 {
		chapters = append(chapters, model.ParsedChapter{
			Index:   0,
			Title:   doc.Title,
			Content: "",
		})
	}

	doc.Chapters = chapters

	// Fallback title if empty.
	if doc.Title == "" {
		doc.Title = "Untitled PDF"
	}

	return doc, nil
}

// extractPDFMetadata pulls title and author from the PDF trailer Info dictionary.
func extractPDFMetadata(r *pdf.Reader, doc *model.ParsedDocument) {
	info := r.Trailer().Key("Info")
	if info.IsNull() {
		return
	}

	if title := info.Key("Title"); !title.IsNull() {
		doc.Title = strings.TrimSpace(title.Text())
	}
	if author := info.Key("Author"); !author.IsNull() {
		doc.Author = strings.TrimSpace(author.Text())
	}
}

// collectOutlineTitles flattens the PDF outline tree into a list of titles,
// traversing depth-first.
func collectOutlineTitles(outline pdf.Outline) []string {
	var titles []string
	var walk func(o pdf.Outline)
	walk = func(o pdf.Outline) {
		if o.Title != "" {
			titles = append(titles, o.Title)
		}
		for _, child := range o.Child {
			walk(child)
		}
	}
	for _, child := range outline.Child {
		walk(child)
	}
	return titles
}
