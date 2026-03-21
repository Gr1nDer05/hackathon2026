package service

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image/png"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type docxBlock struct {
	Paragraph *docxParagraph
	Table     *docxTable
	Image     *docxImage
}

type docxParagraph struct {
	Text   string
	Size   int
	Bold   bool
	Bullet bool
}

type docxTable struct {
	Headers []string
	Rows    [][]string
}

type docxImage struct {
	Filename       string
	RelationshipID string
	AltText        string
	Data           []byte
	WidthEMU       int64
	HeightEMU      int64
}

type parsedHTMLReport struct {
	Blocks []docxBlock
	Images []docxImage
}

func renderReportDOCX(htmlContent []byte) ([]byte, error) {
	report, err := parseHTMLReport(htmlContent)
	if err != nil {
		return nil, err
	}

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)

	files := []struct {
		Name    string
		Content []byte
	}{
		{Name: "[Content_Types].xml", Content: []byte(buildContentTypesXML())},
		{Name: "_rels/.rels", Content: []byte(relsXML)},
		{Name: "word/document.xml", Content: []byte(buildWordDocumentXML(report.Blocks))},
		{Name: "word/_rels/document.xml.rels", Content: []byte(buildDocumentRelationshipsXML(report.Images))},
	}

	for _, image := range report.Images {
		files = append(files, struct {
			Name    string
			Content []byte
		}{
			Name:    "word/media/" + image.Filename,
			Content: image.Data,
		})
	}

	for _, file := range files {
		writer, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(file.Content); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return archive.Bytes(), nil
}

func parseHTMLReport(htmlContent []byte) (parsedHTMLReport, error) {
	root, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return parsedHTMLReport{}, err
	}

	container := findFirstElement(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "main" && hasClass(node, "page")
	})
	if container == nil {
		container = findFirstElement(root, func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.Data == "body"
		})
	}
	if container == nil {
		return parsedHTMLReport{}, fmt.Errorf("report html body not found")
	}

	var report parsedHTMLReport
	if err := appendDocxBlocksFromHTML(container, &report); err != nil {
		return parsedHTMLReport{}, err
	}
	return report, nil
}

func appendDocxBlocksFromHTML(node *html.Node, report *parsedHTMLReport) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		switch child.Data {
		case "style", "script", "head", "title", "meta":
			continue
		case "h1":
			appendParagraphBlock(report, innerText(child), 32, true, false)
		case "h2":
			appendParagraphBlock(report, innerText(child), 26, true, false)
		case "p":
			appendParagraphBlock(report, innerText(child), 22, false, false)
		case "ul":
			for listItem := child.FirstChild; listItem != nil; listItem = listItem.NextSibling {
				if listItem.Type == html.ElementNode && listItem.Data == "li" {
					appendParagraphBlock(report, innerText(listItem), 22, false, true)
				}
			}
		case "table":
			table := parseHTMLTable(child)
			if len(table.Headers) > 0 {
				report.Blocks = append(report.Blocks, docxBlock{Table: &table})
			}
		case "img":
			image, err := parseHTMLImage(child, len(report.Images)+1)
			if err != nil {
				return err
			}
			report.Images = append(report.Images, image)
			report.Blocks = append(report.Blocks, docxBlock{Image: &image})
		case "div":
			if hasClass(child, "meta") {
				metaTable := parseMetaTable(child)
				if len(metaTable.Rows) > 0 {
					report.Blocks = append(report.Blocks, docxBlock{Table: &metaTable})
				}
				continue
			}
			if err := appendDocxBlocksFromHTML(child, report); err != nil {
				return err
			}
		case "main", "body", "section", "figure":
			if err := appendDocxBlocksFromHTML(child, report); err != nil {
				return err
			}
		default:
			if err := appendDocxBlocksFromHTML(child, report); err != nil {
				return err
			}
		}
	}

	return nil
}

func appendParagraphBlock(report *parsedHTMLReport, text string, size int, bold bool, bullet bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	report.Blocks = append(report.Blocks, docxBlock{
		Paragraph: &docxParagraph{
			Text:   text,
			Size:   size,
			Bold:   bold,
			Bullet: bullet,
		},
	})
}

func parseMetaTable(node *html.Node) docxTable {
	table := docxTable{
		Headers: []string{"Параметр", "Значение"},
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || !hasClass(child, "meta-item") {
			continue
		}

		labelNode := findFirstElement(child, func(node *html.Node) bool {
			return node.Type == html.ElementNode && hasClass(node, "meta-label")
		})
		valueNode := findFirstElement(child, func(node *html.Node) bool {
			return node.Type == html.ElementNode && hasClass(node, "meta-value")
		})

		label := innerText(labelNode)
		value := innerText(valueNode)
		if label == "" && value == "" {
			continue
		}
		table.Rows = append(table.Rows, []string{label, value})
	}

	return table
}

func parseHTMLTable(tableNode *html.Node) docxTable {
	var headers []string
	var rows [][]string

	for child := tableNode.FirstChild; child != nil; child = child.NextSibling {
		collectTableRows(child, &headers, &rows)
	}

	return docxTable{
		Headers: headers,
		Rows:    rows,
	}
}

func collectTableRows(node *html.Node, headers *[]string, rows *[][]string) {
	if node == nil || node.Type != html.ElementNode {
		return
	}

	if node.Data == "tr" {
		var cells []string
		hasHeaderCells := false
		for cell := node.FirstChild; cell != nil; cell = cell.NextSibling {
			if cell.Type != html.ElementNode {
				continue
			}
			if cell.Data != "th" && cell.Data != "td" {
				continue
			}
			if cell.Data == "th" {
				hasHeaderCells = true
			}
			cells = append(cells, innerText(cell))
		}
		if len(cells) == 0 {
			return
		}
		if hasHeaderCells && len(*headers) == 0 {
			*headers = append([]string(nil), cells...)
			return
		}
		*rows = append(*rows, append([]string(nil), cells...))
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectTableRows(child, headers, rows)
	}
}

func parseHTMLImage(node *html.Node, index int) (docxImage, error) {
	src := attrValue(node, "src")
	if !strings.HasPrefix(src, "data:image/png;base64,") {
		return docxImage{}, fmt.Errorf("unsupported image src in report html")
	}

	rawData := strings.TrimPrefix(src, "data:image/png;base64,")
	data, err := base64.StdEncoding.DecodeString(rawData)
	if err != nil {
		return docxImage{}, err
	}

	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return docxImage{}, err
	}

	widthEMU, heightEMU := scaleImageForDocx(config.Width, config.Height)
	return docxImage{
		Filename:       fmt.Sprintf("chart-%d.png", index),
		RelationshipID: fmt.Sprintf("rId%d", index),
		AltText:        attrValue(node, "alt"),
		Data:           data,
		WidthEMU:       widthEMU,
		HeightEMU:      heightEMU,
	}, nil
}

func scaleImageForDocx(widthPx, heightPx int) (int64, int64) {
	const emuPerPixel = int64(9525)
	const maxWidthEMU = int64(5760720)

	widthEMU := int64(widthPx) * emuPerPixel
	heightEMU := int64(heightPx) * emuPerPixel
	if widthEMU <= maxWidthEMU {
		return widthEMU, heightEMU
	}

	scaledHeight := heightEMU * maxWidthEMU / widthEMU
	return maxWidthEMU, scaledHeight
}

func buildWordDocumentXML(blocks []docxBlock) string {
	var body strings.Builder
	imageIndex := 0

	for _, block := range blocks {
		switch {
		case block.Paragraph != nil:
			writeDocxParagraph(&body, block.Paragraph)
		case block.Table != nil:
			writeDocxTable(&body, block.Table.Headers, block.Table.Rows)
		case block.Image != nil:
			imageIndex++
			writeDocxImage(&body, *block.Image, imageIndex)
		}
	}

	body.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" ` +
		`xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" ` +
		`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" ` +
		`xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">` +
		`<w:body>` + body.String() + `</w:body></w:document>`
}

func writeDocxParagraph(builder *strings.Builder, paragraph *docxParagraph) {
	text := strings.TrimSpace(paragraph.Text)
	if text == "" {
		return
	}
	if paragraph.Bullet {
		text = "• " + text
	}

	builder.WriteString(`<w:p><w:r><w:rPr>`)
	if paragraph.Bold {
		builder.WriteString(`<w:b/>`)
	}
	builder.WriteString(`<w:sz w:val="`)
	builder.WriteString(escapeDocxText(strconv.Itoa(paragraph.Size)))
	builder.WriteString(`"/></w:rPr><w:t xml:space="preserve">`)
	builder.WriteString(escapeDocxText(text))
	builder.WriteString(`</w:t></w:r></w:p>`)
}

func writeDocxImage(builder *strings.Builder, image docxImage, imageIndex int) {
	builder.WriteString(`<w:p><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">`)
	builder.WriteString(`<wp:extent cx="`)
	builder.WriteString(strconv.FormatInt(image.WidthEMU, 10))
	builder.WriteString(`" cy="`)
	builder.WriteString(strconv.FormatInt(image.HeightEMU, 10))
	builder.WriteString(`"/>`)
	builder.WriteString(`<wp:effectExtent l="0" t="0" r="0" b="0"/>`)
	builder.WriteString(`<wp:docPr id="`)
	builder.WriteString(strconv.Itoa(imageIndex))
	builder.WriteString(`" name="`)
	builder.WriteString(escapeDocxText(image.Filename))
	builder.WriteString(`" descr="`)
	builder.WriteString(escapeDocxText(image.AltText))
	builder.WriteString(`"/>`)
	builder.WriteString(`<wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr>`)
	builder.WriteString(`<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`)
	builder.WriteString(`<pic:pic><pic:nvPicPr><pic:cNvPr id="`)
	builder.WriteString(strconv.Itoa(imageIndex))
	builder.WriteString(`" name="`)
	builder.WriteString(escapeDocxText(image.Filename))
	builder.WriteString(`" descr="`)
	builder.WriteString(escapeDocxText(image.AltText))
	builder.WriteString(`"/><pic:cNvPicPr/></pic:nvPicPr>`)
	builder.WriteString(`<pic:blipFill><a:blip r:embed="`)
	builder.WriteString(escapeDocxText(image.RelationshipID))
	builder.WriteString(`"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`)
	builder.WriteString(`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="`)
	builder.WriteString(strconv.FormatInt(image.WidthEMU, 10))
	builder.WriteString(`" cy="`)
	builder.WriteString(strconv.FormatInt(image.HeightEMU, 10))
	builder.WriteString(`"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`)
	builder.WriteString(`</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`)
}

func writeDocxTable(builder *strings.Builder, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	builder.WriteString(`<w:tbl><w:tblPr><w:tblBorders>`)
	builder.WriteString(`<w:top w:val="single" w:sz="6" w:space="0" w:color="AAB7C4"/>`)
	builder.WriteString(`<w:left w:val="single" w:sz="6" w:space="0" w:color="AAB7C4"/>`)
	builder.WriteString(`<w:bottom w:val="single" w:sz="6" w:space="0" w:color="AAB7C4"/>`)
	builder.WriteString(`<w:right w:val="single" w:sz="6" w:space="0" w:color="AAB7C4"/>`)
	builder.WriteString(`<w:insideH w:val="single" w:sz="4" w:space="0" w:color="D7DDE8"/>`)
	builder.WriteString(`<w:insideV w:val="single" w:sz="4" w:space="0" w:color="D7DDE8"/>`)
	builder.WriteString(`</w:tblBorders></w:tblPr>`)
	writeDocxTableRow(builder, headers, true)
	for _, row := range rows {
		writeDocxTableRow(builder, row, false)
	}
	builder.WriteString(`</w:tbl>`)
}

func writeDocxTableRow(builder *strings.Builder, cells []string, bold bool) {
	builder.WriteString(`<w:tr>`)
	for _, cell := range cells {
		builder.WriteString(`<w:tc><w:p><w:r><w:rPr>`)
		if bold {
			builder.WriteString(`<w:b/>`)
		}
		builder.WriteString(`<w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">`)
		builder.WriteString(escapeDocxText(cell))
		builder.WriteString(`</w:t></w:r></w:p></w:tc>`)
	}
	builder.WriteString(`</w:tr>`)
}

func buildDocumentRelationshipsXML(images []docxImage) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, image := range images {
		builder.WriteString(`<Relationship Id="`)
		builder.WriteString(escapeDocxText(image.RelationshipID))
		builder.WriteString(`" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/`)
		builder.WriteString(escapeDocxText(image.Filename))
		builder.WriteString(`"/>`)
	}
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func buildContentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
}

func findFirstElement(node *html.Node, predicate func(*html.Node) bool) *html.Node {
	if node == nil {
		return nil
	}
	if predicate(node) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findFirstElement(child, predicate); found != nil {
			return found
		}
	}
	return nil
}

func hasClass(node *html.Node, className string) bool {
	classes := strings.Fields(attrValue(node, "class"))
	for _, class := range classes {
		if class == className {
			return true
		}
	}
	return false
}

func attrValue(node *html.Node, attrName string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func innerText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)

	return strings.Join(strings.Fields(builder.String()), " ")
}

func escapeDocxText(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}

const relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
