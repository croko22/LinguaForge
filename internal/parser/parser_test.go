package parser

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/croko/language-app/internal/model"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// writeOrFatal writes a string to w, failing the test on error.
func writeOrFatal(t *testing.T, w io.Writer, s string) {
	t.Helper()
	if _, err := io.WriteString(w, s); err != nil {
		t.Fatalf("write error: %v", err)
	}
}

// writeZipEntry creates a file inside zw and writes content to it.
func writeZipEntry(t *testing.T, zw *zip.Writer, name, content string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("zip create %q: %v", name, err)
	}
	writeOrFatal(t, w, content)
}

// buildTestEPUB creates an in-memory valid EPUB 2 file as a byte slice.
//
//	numChapters — number of spine-referenced XHTML files (> 0).
//	    If 0, one XHTML file is placed in the manifest but NOT referenced
//	    by the spine, and no NCX is written (extreme-fallback scenario).
//	includeNCX — if true, an NCX with navPoint entries is added for each chapter.
//	    Only meaningful when numChapters > 0.
func buildTestEPUB(t *testing.T, numChapters int, includeNCX bool) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// ── 1. mimetype — MUST be first entry, stored (not compressed) ────
	mt, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("zip create mimetype header: %v", err)
	}
	writeOrFatal(t, mt, "application/epub+zip")

	// ── 2. META-INF/container.xml ──────────────────────────────────────
	writeZipEntry(t, zw, "META-INF/container.xml", `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/package.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)

	// ── Determine manifest / spine layout ────────────────────────────────
	manifestChapterCount := numChapters
	if manifestChapterCount == 0 {
		manifestChapterCount = 1 // include one XHTML even with empty spine
	}

	// Collect manifest entries.
	type manifestItem struct {
		id, href, mime string
	}
	var items []manifestItem

	for i := 1; i <= manifestChapterCount; i++ {
		items = append(items, manifestItem{
			id:   fmt.Sprintf("chapter%d", i),
			href: fmt.Sprintf("chapter%d.xhtml", i),
			mime: "application/xhtml+xml",
		})
	}

	if includeNCX && numChapters > 0 {
		items = append(items, manifestItem{
			id:   "ncx",
			href: "toc.ncx",
			mime: "application/x-dtbncx+xml",
		})
	}

	// ── 3. OEBPS/package.opf ───────────────────────────────────────────
	var manifestXML strings.Builder
	manifestXML.WriteString("<manifest>\n")
	for _, it := range items {
		fmt.Fprintf(&manifestXML, `<item id="%s" href="%s" media-type="%s"/>`+"\n", it.id, it.href, it.mime)
	}
	manifestXML.WriteString("</manifest>")

	var spineXML strings.Builder
	if includeNCX && numChapters > 0 {
		spineXML.WriteString(`<spine toc="ncx">` + "\n")
	} else {
		spineXML.WriteString("<spine>\n")
	}
	for i := 1; i <= numChapters; i++ {
		spineXML.WriteString(fmt.Sprintf(`<itemref idref="chapter%d"/>`+"\n", i))
	}
	spineXML.WriteString("</spine>")

	opf := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="bookid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="bookid">urn:uuid:test-book</dc:identifier>
    <dc:title>Test Book</dc:title>
    <dc:creator opf:role="aut">Test Author</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  %s
  %s
</package>`, manifestXML.String(), spineXML.String())

	writeZipEntry(t, zw, "OEBPS/package.opf", opf)

	// ── 4. OEBPS/toc.ncx (optional) ────────────────────────────────────
	if includeNCX && numChapters > 0 {
		var ncx strings.Builder
		ncx.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx version="2005-1" xmlns="http://www.daisy.org/z3986/2005/ncx/">
<head>
  <meta name="dtb:uid" content="test-book"/>
  <meta name="dtb:depth" content="1"/>
  <meta name="dtb:totalPageCount" content="0"/>
  <meta name="dtb:maxPageNumber" content="0"/>
</head>
<docTitle><text>Test Book</text></docTitle>
<navMap>
`)
		for i := 1; i <= numChapters; i++ {
			fmt.Fprintf(&ncx, `<navPoint id="navpoint-%d" playOrder="%d">
    <navLabel><text>Chapter %d: Content</text></navLabel>
    <content src="chapter%d.xhtml"/>
  </navPoint>
`, i, i, i, i)
		}
		ncx.WriteString(`</navMap>
</ncx>`)

		writeZipEntry(t, zw, "OEBPS/toc.ncx", ncx.String())
	}

	// ── 5. XHTML chapter files ──────────────────────────────────────────
	for i := 1; i <= manifestChapterCount; i++ {
		ch := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter %d</title></head>
<body>
  <h1>Chapter %d</h1>
  <p>This is the content of chapter %d. It contains meaningful text for testing purposes.</p>
  <p>Multiple paragraphs help test the whitespace collapsing and text extraction logic.</p>
</body>
</html>`, i, i, i)
		writeZipEntry(t, zw, fmt.Sprintf("OEBPS/chapter%d.xhtml", i), ch)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// buildEmptyContentEPUB creates an in-memory EPUB whose XHTML files have empty
// <body> elements — no extractable text. Used to test the empty-content code path.
func buildEmptyContentEPUB(t *testing.T, numChapters int) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// mimetype
	mt, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("zip create mimetype header: %v", err)
	}
	writeOrFatal(t, mt, "application/epub+zip")

	// container.xml
	writeZipEntry(t, zw, "META-INF/container.xml", `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/package.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)

	// manifest + spine
	var manifestXML, spineXML strings.Builder
	manifestXML.WriteString("<manifest>\n")
	spineXML.WriteString("<spine>\n")
	for i := 1; i <= numChapters; i++ {
		fmt.Fprintf(&manifestXML, `<item id="chapter%d" href="chapter%d.xhtml" media-type="application/xhtml+xml"/>`+"\n", i, i)
		fmt.Fprintf(&spineXML, `<itemref idref="chapter%d"/>`+"\n", i)
	}
	manifestXML.WriteString("</manifest>")
	spineXML.WriteString("</spine>")

	opf := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="bookid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="bookid">urn:uuid:test-book</dc:identifier>
    <dc:title>Test Book</dc:title>
    <dc:creator opf:role="aut">Test Author</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  %s
  %s
</package>`, manifestXML.String(), spineXML.String())

	writeZipEntry(t, zw, "OEBPS/package.opf", opf)

	// XHTML files with empty <body>
	for i := 1; i <= numChapters; i++ {
		ch := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Empty Chapter</title></head>
<body>
</body>
</html>`
		writeZipEntry(t, zw, fmt.Sprintf("OEBPS/chapter%d.xhtml", i), ch)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// parseTestEPUB is a small convenience wrapper around EpubParser.Parse.
func parseTestEPUB(t *testing.T, data []byte) (*model.ParsedDocument, error) {
	t.Helper()
	p := NewEpubParser()
	return p.Parse(bytes.NewReader(data), int64(len(data)))
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestParseStandardEPUB(t *testing.T) {
	data := buildTestEPUB(t, 3, true) // 3 chapters, with NCX
	doc, err := parseTestEPUB(t, data)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	// ── Metadata checks ────────────────────────────────────────────────
	if doc.Title != "Test Book" {
		t.Errorf("doc.Title = %q, want %q", doc.Title, "Test Book")
	}
	if doc.Author != "Test Author" {
		t.Errorf("doc.Author = %q, want %q", doc.Author, "Test Author")
	}
	if doc.Language != "en" {
		t.Errorf("doc.Language = %q, want %q", doc.Language, "en")
	}

	// ── Chapter count ──────────────────────────────────────────────────
	if len(doc.Chapters) != 3 {
		t.Fatalf("len(doc.Chapters) = %d, want 3", len(doc.Chapters))
	}

	// ── Per-chapter checks ─────────────────────────────────────────────
	expectedTitles := []string{
		"Chapter 1: Content",
		"Chapter 2: Content",
		"Chapter 3: Content",
	}
	for i, ch := range doc.Chapters {
		if ch.Title != expectedTitles[i] {
			t.Errorf("chapter[%d].Title = %q, want %q", i, ch.Title, expectedTitles[i])
		}
		if ch.Index != i {
			t.Errorf("chapter[%d].Index = %d, want %d", i, ch.Index, i)
		}
		// Content must be clean plain text (no HTML tags).
		if strings.Contains(ch.Content, "<") || strings.Contains(ch.Content, ">") {
			t.Errorf("chapter[%d].Content contains HTML markup", i)
		}
		// Content must be non-empty.
		if len(ch.Content) == 0 {
			t.Errorf("chapter[%d].Content is empty", i)
		}
		// Content should contain meaningful text.
		if !strings.Contains(ch.Content, fmt.Sprintf("content of chapter %d", i+1)) {
			t.Errorf("chapter[%d].Content missing expected text:\n%s", i, ch.Content)
		}

		// NOTE: ParsedChapter does not carry TokenCount, so we use
		// Content length as a proxy for "content was extracted".
	}
}

func TestParseEPUBWithoutNCX(t *testing.T) {
	// 3 spine items, no NCX → auto-titles "Chapter 1" …
	data := buildTestEPUB(t, 3, false)
	doc, err := parseTestEPUB(t, data)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(doc.Chapters) != 3 {
		t.Fatalf("len(doc.Chapters) = %d, want 3", len(doc.Chapters))
	}

	for i, ch := range doc.Chapters {
		wantTitle := fmt.Sprintf("Chapter %d", i+1)
		if ch.Title != wantTitle {
			t.Errorf("chapter[%d].Title = %q, want %q", i, ch.Title, wantTitle)
		}
		if ch.Index != i {
			t.Errorf("chapter[%d].Index = %d, want %d", i, ch.Index, i)
		}
		if len(ch.Content) == 0 {
			t.Errorf("chapter[%d].Content is empty", i)
		}
	}
}

func TestParseSingleChapterEPUB(t *testing.T) {
	// numChapters=0 and includeNCX=false produces:
	//   - 1 XHTML file in the manifest
	//   - empty spine (no itemrefs)
	//   - no NCX
	// This triggers the fallback branch in Parse: len(chapters)==0
	// → a single anonymous chapter with doc.Title.
	data := buildTestEPUB(t, 0, false)
	doc, err := parseTestEPUB(t, data)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(doc.Chapters) != 1 {
		t.Fatalf("len(doc.Chapters) = %d, want 1 (fallback)", len(doc.Chapters))
	}

	if doc.Chapters[0].Title != doc.Title {
		t.Errorf("chapter[0].Title = %q, want %q (document title)", doc.Chapters[0].Title, doc.Title)
	}
	if doc.Chapters[0].Index != 0 {
		t.Errorf("chapter[0].Index = %d, want 0", doc.Chapters[0].Index)
	}
}

func TestCanParse(t *testing.T) {
	p := NewEpubParser()

	tests := []struct {
		filename string
		want     bool
	}{
		{"book.epub", true},
		{"book.EPUB", true},
		{"book.Epub", true},
		{"book.PUB", false},
		{"book.pdf", false},
		{"book.txt", false},
		{"noextension", false},
		{"", false},
	}

	for _, tc := range tests {
		got := p.CanParse(tc.filename)
		if got != tc.want {
			t.Errorf("CanParse(%q) = %v, want %v", tc.filename, got, tc.want)
		}
	}
}

func TestParseCorruptEPUB(t *testing.T) {
	p := NewEpubParser()
	data := []byte("this is not a valid zip file at all!!!")
	_, err := p.Parse(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("Parse() should have returned an error for corrupt data")
	}
}

func TestParseEmptyContent(t *testing.T) {
	data := buildEmptyContentEPUB(t, 2)
	doc, err := parseTestEPUB(t, data)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(doc.Chapters) != 2 {
		t.Fatalf("len(doc.Chapters) = %d, want 2", len(doc.Chapters))
	}

	for i, ch := range doc.Chapters {
		if ch.Content != "" {
			t.Errorf("chapter[%d].Content = %q, want empty string", i, ch.Content)
		}
	}
}
