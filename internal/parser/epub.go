package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/croko/language-app/internal/model"
	"golang.org/x/net/html/charset"
)

// ── XML parsing structs ──────────────────────────────────────────────────────

// Container maps to META-INF/container.xml.
type Container struct {
	Rootfiles struct {
		Rootfile []struct {
			FullPath  string `xml:"full-path,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"rootfile"`
	} `xml:"rootfiles"`
}

// Package maps to the OPF package document.
type Package struct {
	Metadata struct {
		Titles   []string `xml:"http://purl.org/dc/elements/1.1/ title"`
		Creators []struct {
			Text string `xml:",chardata"`
			Role string `xml:"role,attr"`
		} `xml:"http://purl.org/dc/elements/1.1/ creator"`
		Language  string `xml:"http://purl.org/dc/elements/1.1/ language"`
		MetaItems []struct {
			Name    string `xml:"name,attr"`
			Content string `xml:"content,attr"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			Type       string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Toc   string `xml:"toc,attr"`
		Items []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

// NCX maps to the EPUB 2 NCX table of contents.
type NCX struct {
	NavMap struct {
		NavPoints []struct {
			ID        string `xml:"id,attr"`
			PlayOrder int    `xml:"playOrder,attr"`
			NavLabel  struct {
				Text string `xml:"text"`
			} `xml:"navLabel"`
			Content struct {
				Src string `xml:"src,attr"`
			} `xml:"content"`
		} `xml:"navPoint"`
	} `xml:"navMap"`
}

// chapterDef groups content hrefs under a single chapter title.
// Used internally by Parse and the NCX mapping helpers.
type chapterDef struct {
	title string
	hrefs []string
}

// ── whitespace collapsing regex (compile once) ───────────────────────────────

var whitespaceRe = regexp.MustCompile(`\s+`)

// xmlEncodingRe matches encoding="..." in XML declarations.
var xmlEncodingRe = regexp.MustCompile(`<\?xml\s+[^>]*encoding\s*=\s*["']([^"']+)["']`)

// ── EpubParser ───────────────────────────────────────────────────────────────

// EpubParser parses EPUB files using archive/zip and XML parsing.
type EpubParser struct{}

// NewEpubParser creates a new EpubParser.
func NewEpubParser() *EpubParser {
	return &EpubParser{}
}

// CanParse returns true if the filename has an .epub extension (case-insensitive).
func (p *EpubParser) CanParse(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".epub")
}

// Parse reads an EPUB file and returns a ParsedDocument with chapters.
func (p *EpubParser) Parse(readerAt ReaderAt, size int64) (*model.ParsedDocument, error) {
	zipFiles, err := openEPUB(readerAt, size)
	if err != nil {
		return nil, err
	}

	pkg, opfBaseDir, err := parseOPF(zipFiles)
	if err != nil {
		return nil, err
	}

	manifestHrefs := buildManifestHrefs(pkg, opfBaseDir)

	doc := extractMetadata(pkg)

	doc.CoverImageData = extractCoverImage(&pkg, manifestHrefs, zipFiles, opfBaseDir)

	spineHrefs := buildSpineHrefs(pkg, manifestHrefs)

	chapters := mapChapters(pkg, manifestHrefs, opfBaseDir, zipFiles, spineHrefs)

	// Fallback: if nothing matched, create a single anonymous chapter.
	if len(chapters) == 0 {
		chapters = append(chapters, chapterDef{
			title: doc.Title,
			hrefs: spineHrefs,
		})
	}

	for _, ch := range chapters {
		content := extractChapterContent(ch.hrefs, zipFiles)
		doc.Chapters = append(doc.Chapters, model.ParsedChapter{
			Index:   len(doc.Chapters),
			Title:   ch.title,
			Content: content,
		})
	}

	return doc, nil
}

func openEPUB(readerAt ReaderAt, size int64) (map[string]*zip.File, error) {
	zipReader, err := zip.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("epub: failed to open zip: %w", err)
	}
	zipFiles := make(map[string]*zip.File, len(zipReader.File)*2)
	for _, f := range zipReader.File {
		zipFiles[f.Name] = f
		lower := strings.ToLower(f.Name)
		if lower != f.Name {
			if _, exists := zipFiles[lower]; !exists {
				zipFiles[lower] = f
			}
		}
	}
	return zipFiles, nil
}

func parseOPF(zipFiles map[string]*zip.File) (Package, string, error) {
	containerData, err := readZipFile(zipFiles, "META-INF/container.xml")
	if err != nil {
		return Package{}, "", fmt.Errorf("epub: missing META-INF/container.xml: %w", err)
	}

	var container Container
	if err := xml.Unmarshal(containerData, &container); err != nil {
		return Package{}, "", fmt.Errorf("epub: failed to parse container.xml: %w", err)
	}

	if len(container.Rootfiles.Rootfile) == 0 {
		return Package{}, "", fmt.Errorf("epub: no rootfile found in container.xml")
	}

	opfPath := container.Rootfiles.Rootfile[0].FullPath
	opfBaseDir := path.Dir(opfPath)

	opfData, err := readZipFile(zipFiles, opfPath)
	if err != nil {
		return Package{}, "", fmt.Errorf("epub: failed to read OPF %s: %w", opfPath, err)
	}

	var pkg Package
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return Package{}, "", fmt.Errorf("epub: failed to parse OPF: %w", err)
	}

	return pkg, opfBaseDir, nil
}

func buildManifestHrefs(pkg Package, opfBaseDir string) map[string]string {
	manifestHrefs := make(map[string]string, len(pkg.Manifest.Items))
	for _, item := range pkg.Manifest.Items {
		manifestHrefs[item.ID] = path.Clean(path.Join(opfBaseDir, item.Href))
	}
	return manifestHrefs
}

func extractMetadata(pkg Package) *model.ParsedDocument {
	doc := &model.ParsedDocument{}
	if len(pkg.Metadata.Titles) > 0 {
		doc.Title = strings.TrimSpace(pkg.Metadata.Titles[0])
	}
	for _, c := range pkg.Metadata.Creators {
		if c.Role == "" || c.Role == "aut" {
			doc.Author = strings.TrimSpace(c.Text)
			break
		}
	}
	doc.Language = strings.TrimSpace(pkg.Metadata.Language)
	return doc
}

func buildSpineHrefs(pkg Package, manifestHrefs map[string]string) []string {
	spineHrefs := make([]string, 0, len(pkg.Spine.Items))
	seen := make(map[string]bool, len(pkg.Spine.Items))
	for _, item := range pkg.Spine.Items {
		href, ok := manifestHrefs[item.IDRef]
		if !ok {
			continue
		}
		canonical := path.Clean(href)
		if seen[canonical] {
			continue
		}
		seen[canonical] = true
		spineHrefs = append(spineHrefs, canonical)
	}
	return spineHrefs
}

func mapChapters(pkg Package, manifestHrefs map[string]string, opfBaseDir string, zipFiles map[string]*zip.File, spineHrefs []string) []chapterDef {
	ncxHref := resolveNCXPath(pkg, manifestHrefs, opfBaseDir)

	var ncx *NCX
	if ncxHref != "" {
		if ncxData, err := readZipFile(zipFiles, ncxHref); err == nil {
			var n NCX
			if xml.Unmarshal(ncxData, &n) == nil {
				ncx = &n
			}
		}
	}

	if ncx != nil && len(ncx.NavMap.NavPoints) > 0 {
		return mapChaptersFromNCX(ncx, ncxHref, spineHrefs)
	}

	chapters := make([]chapterDef, 0, len(spineHrefs))
	for i, href := range spineHrefs {
		chapters = append(chapters, chapterDef{
			title: fmt.Sprintf("Chapter %d", i+1),
			hrefs: []string{href},
		})
	}
	return chapters
}

// ── NCX resolution ───────────────────────────────────────────────────────────

// resolveNCXPath finds the NCX file path from the OPF data.
// It first checks the spine's toc attribute, then scans the manifest for
// application/x-dtbncx+xml media-type.
func resolveNCXPath(pkg Package, manifestHrefs map[string]string, opfBaseDir string) string {
	// Prefer the spine's toc attribute.
	if pkg.Spine.Toc != "" {
		if href, ok := manifestHrefs[pkg.Spine.Toc]; ok {
			return href
		}
	}
	// Fallback: scan manifest for NCX media-type.
	for _, item := range pkg.Manifest.Items {
		if item.Type == "application/x-dtbncx+xml" {
			return path.Clean(path.Join(opfBaseDir, item.Href))
		}
	}
	return ""
}

// mapChaptersFromNCX builds chapter definitions from NCX navPoint entries,
// matching each to the spine content hrefs.
func mapChaptersFromNCX(ncx *NCX, ncxHref string, spineHrefs []string) []chapterDef {
	ncxDir := path.Dir(ncxHref)

	// Build a set of spine hrefs for quick matching.
	spineHrefSet := make(map[string]bool, len(spineHrefs))
	for _, h := range spineHrefs {
		spineHrefSet[h] = true
	}

	var chapters []chapterDef
	covered := make(map[string]bool)

	for _, np := range ncx.NavMap.NavPoints {
		resolvedSrc := path.Clean(path.Join(ncxDir, np.Content.Src))

		// Try exact match first.
		matchingHrefs := matchHrefToSpine(resolvedSrc, spineHrefSet)

		title := strings.TrimSpace(np.NavLabel.Text)
		if title == "" {
			title = fmt.Sprintf("Chapter %d", np.PlayOrder)
		}

		for _, h := range matchingHrefs {
			covered[h] = true
		}

		chapters = append(chapters, chapterDef{
			title: title,
			hrefs: matchingHrefs,
		})
	}

	// Collect spine items not covered by any NCX entry.
	var uncovered []string
	for _, h := range spineHrefs {
		if !covered[h] {
			uncovered = append(uncovered, h)
		}
	}
	if len(uncovered) > 0 {
		chapters = append(chapters, chapterDef{
			title: fmt.Sprintf("Chapter %d", len(chapters)+1),
			hrefs: uncovered,
		})
	}

	return chapters
}

// matchHrefToSpine tries to find the spine href(s) matching an NCX src.
// It handles exact matches and fragment-stripped matches.
func matchHrefToSpine(resolvedSrc string, spineHrefSet map[string]bool) []string {
	// Exact match.
	if spineHrefSet[resolvedSrc] {
		return []string{resolvedSrc}
	}

	// Try without fragment (#...).
	baseSrc := strings.SplitN(resolvedSrc, "#", 2)[0]
	if spineHrefSet[baseSrc] {
		return []string{baseSrc}
	}

	// No match found.
	return nil
}

// ── Content extraction ───────────────────────────────────────────────────────

// extractChapterContent reads all XHTML files listed in hrefs and concatenates
// their clean text content.
func extractChapterContent(hrefs []string, zipFiles map[string]*zip.File) string {
	if len(hrefs) == 0 {
		return ""
	}

	var buf strings.Builder
	for _, href := range hrefs {
		f, ok := zipFiles[href]
		if !ok {
			continue
		}
		if !isXHTMLContent(f) {
			continue
		}

		data, err := readZipFileEntry(f)
		if err != nil {
			continue
		}

		text, err := extractTextFromXHTML(data)
		if err != nil || text == "" {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(text)
	}
	return buf.String()
}

// isXHTMLContent returns true if the zip entry is an XHTML/HTML file based on
// its filename extension.
func isXHTMLContent(f *zip.File) bool {
	ext := strings.ToLower(path.Ext(f.Name))
	switch ext {
	case ".xhtml", ".html", ".htm", ".xml":
		return true
	default:
		return false
	}
}

// extractTextFromXHTML parses the given XHTML data using goquery and returns
// clean plain text: script/style elements removed, body text extracted,
// whitespace collapsed.
func extractTextFromXHTML(data []byte) (string, error) {
	reader := decodeReader(data)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", fmt.Errorf("goquery parse error: %w", err)
	}

	doc.Find("script, style").Remove()
	text := doc.Find("body").Text()
	text = strings.TrimSpace(text)
	text = whitespaceRe.ReplaceAllString(text, " ")
	return text, nil
}

// decodeReader wraps data in a charset-aware reader if the XML declaration
// specifies a non-UTF-8 encoding.
func decodeReader(data []byte) io.Reader {
	enc := detectEncoding(data)
	if strings.EqualFold(enc, "utf-8") || strings.EqualFold(enc, "us-ascii") {
		return bytes.NewReader(data)
	}

	// Attempt charset decoding for legacy encodings (e.g. windows-1252, iso-8859-1).
	reader, err := charset.NewReaderLabel(enc, bytes.NewReader(data))
	if err != nil {
		// Fallback: return raw data as-is.
		return bytes.NewReader(data)
	}
	return reader
}

// detectEncoding extracts the encoding from the XML declaration.
// Returns "utf-8" if none is found.
func detectEncoding(data []byte) string {
	// Check for UTF-8 BOM.
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8"
	}

	// Only scan the first line for the XML declaration.
	var firstLine []byte
	if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
		firstLine = data[:idx]
	} else if len(data) > 256 {
		firstLine = data[:256]
	} else {
		firstLine = data
	}

	matches := xmlEncodingRe.FindSubmatch(firstLine)
	if len(matches) >= 2 {
		return string(matches[1])
	}
	return "utf-8"
}

// ── Cover image extraction ────────────────────────────────────────────────────

// extractCoverImage extracts the cover image from an EPUB using three strategies:
//  1. Manifest item with properties="cover-image" (EPUB 3)
//  2. <meta name="cover" content="..."> pointing to a manifest item (EPUB 2/3)
//  3. Item with id="cover" or id="cover-image" convention
func extractCoverImage(pkg *Package, manifestHrefs map[string]string, zipFiles map[string]*zip.File, opfBaseDir string) []byte {
	// Strategy 1: properties="cover-image" (EPUB 3)
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(strings.ToLower(item.Properties), "cover-image") {
			if data := readCoverZipFile(zipFiles, path.Join(opfBaseDir, item.Href)); data != nil {
				return data
			}
		}
	}

	// Strategy 2: <meta name="cover" content="..."> (EPUB 2/3)
	for _, m := range pkg.Metadata.MetaItems {
		if strings.ToLower(m.Name) == "cover" {
			if href, ok := manifestHrefs[m.Content]; ok {
				if data := readCoverZipFile(zipFiles, href); data != nil {
					return data
				}
			}
		}
	}

	// Strategy 3: id="cover" or id="cover-image" convention
	for _, item := range pkg.Manifest.Items {
		id := strings.ToLower(item.ID)
		if id == "cover" || id == "cover-image" {
			if data := readCoverZipFile(zipFiles, path.Join(opfBaseDir, item.Href)); data != nil {
				return data
			}
		}
	}

	return nil
}

// readCoverZipFile tries to read a file from the zip index, normalizing
// path separators. Returns nil on any error (caller treats missing cover
// as non-fatal).
func readCoverZipFile(zipFiles map[string]*zip.File, filePath string) []byte {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	if f, ok := zipFiles[normalized]; ok {
		data, err := readZipFileEntry(f)
		if err == nil {
			return data
		}
	}
	return nil
}

// ── ZIP helpers ──────────────────────────────────────────────────────────────

// readZipFile reads a file from the zip index by name.
func readZipFile(files map[string]*zip.File, name string) ([]byte, error) {
	f, ok := files[name]
	if !ok {
		f, ok = files[strings.ToLower(name)]
	}
	if !ok {
		return nil, fmt.Errorf("file not found in zip: %s", name)
	}
	return readZipFileEntry(f)
}

// readZipFileEntry reads and returns the full contents of a zip entry.
func readZipFileEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
