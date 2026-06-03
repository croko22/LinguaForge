package parser

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"testing"
)

func createTestCoverImage() []byte {
	buf := new(bytes.Buffer)
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if err := png.Encode(buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func buildCoverEPUB(coverData []byte) *bytes.Buffer {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	mh := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
	mf, _ := w.CreateHeader(mh)
	mf.Write([]byte("application/epub+zip"))

	f, _ := w.Create("META-INF/container.xml")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	f, _ = w.Create("OEBPS/content.opf")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Cover Test</dc:title>
    <dc:creator xmlns:dc="http://purl.org/dc/elements/1.1/" opf:role="aut" xmlns:opf="http://www.idpf.org/2007/opf">Test Author</dc:creator>
    <dc:language xmlns:dc="http://purl.org/dc/elements/1.1/">en</dc:language>
  </metadata>
  <manifest>
    <item id="cover-img" href="cover.png" media-type="image/png" properties="cover-image"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="chapter1"/>
  </spine>
</package>`))

	f, _ = w.Create("OEBPS/cover.png")
	f.Write(coverData)

	f, _ = w.Create("OEBPS/chapter1.xhtml")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body><p>Content.</p></body>
</html>`))

	w.Close()
	return &buf
}

func buildCoverWithoutCoverEPUB() *bytes.Buffer {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	mh := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
	mf, _ := w.CreateHeader(mh)
	mf.Write([]byte("application/epub+zip"))

	f, _ := w.Create("META-INF/container.xml")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	f, _ = w.Create("OEBPS/content.opf")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">No Cover</dc:title>
    <dc:creator xmlns:dc="http://purl.org/dc/elements/1.1/" opf:role="aut" xmlns:opf="http://www.idpf.org/2007/opf">Author</dc:creator>
    <dc:language xmlns:dc="http://purl.org/dc/elements/1.1/">en</dc:language>
  </metadata>
  <manifest>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="chapter1"/>
  </spine>
</package>`))

	f, _ = w.Create("OEBPS/chapter1.xhtml")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body><p>Content.</p></body>
</html>`))

	w.Close()
	return &buf
}

func TestEPUBCoverExtraction(t *testing.T) {
	coverData := createTestCoverImage()
	data := buildCoverEPUB(coverData)

	p := NewEpubParser()
	doc, err := p.Parse(bytes.NewReader(data.Bytes()), int64(data.Len()))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.CoverImageData) == 0 {
		t.Fatal("expected CoverImageData to be non-empty")
	}
	if !bytes.Equal(doc.CoverImageData, coverData) {
		t.Fatal("CoverImageData does not match original cover data")
	}
}

func TestEPUBCoverWithoutCover(t *testing.T) {
	data := buildCoverWithoutCoverEPUB()

	p := NewEpubParser()
	doc, err := p.Parse(bytes.NewReader(data.Bytes()), int64(data.Len()))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.CoverImageData != nil {
		t.Fatal("expected CoverImageData to be nil for EPUB without cover")
	}
}
