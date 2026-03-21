package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"

	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

var reportChartPalette = []drawing.Color{
	drawing.ColorFromHex("#0F766E"),
	drawing.ColorFromHex("#2563EB"),
	drawing.ColorFromHex("#F97316"),
	drawing.ColorFromHex("#9333EA"),
	drawing.ColorFromHex("#D946EF"),
}

func buildHTMLReportDocument(document reportDocument) (reportDocument, error) {
	htmlDocument := document
	htmlDocument.Meta = append([]reportMetaItem(nil), document.Meta...)
	htmlDocument.Sections = make([]reportSection, 0, len(document.Sections))

	for _, section := range document.Sections {
		htmlSection := section
		htmlSection.Paragraphs = append([]string(nil), section.Paragraphs...)
		htmlSection.Bullets = append([]string(nil), section.Bullets...)
		htmlSection.TableHeader = append([]string(nil), section.TableHeader...)
		htmlSection.ChartBars = append([]reportChartBar(nil), section.ChartBars...)
		if len(section.TableRows) > 0 {
			htmlSection.TableRows = make([][]string, 0, len(section.TableRows))
			for _, row := range section.TableRows {
				htmlSection.TableRows = append(htmlSection.TableRows, append([]string(nil), row...))
			}
		}

		if len(section.ChartBars) > 0 {
			chartPNG, err := renderScaleChartPNG(section.Title, section.ChartBars)
			if err != nil {
				return reportDocument{}, err
			}
			htmlSection.ChartImageDataURI = template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(chartPNG))
			htmlSection.ChartAltText = section.Title
			if strings.TrimSpace(htmlSection.ChartCaption) == "" {
				htmlSection.ChartCaption = "Нормализованные проценты по карьерным шкалам. Более высокий столбец означает более выраженную склонность по этой шкале."
			}
		}

		htmlDocument.Sections = append(htmlDocument.Sections, htmlSection)
	}

	return htmlDocument, nil
}

func renderScaleChartPNG(title string, bars []reportChartBar) ([]byte, error) {
	if len(bars) == 0 {
		return nil, nil
	}

	chartBars := make([]chart.Value, 0, len(bars))
	for index, bar := range bars {
		fillColor := reportChartPalette[index%len(reportChartPalette)]
		chartBars = append(chartBars, chart.Value{
			Label: bar.Label,
			Value: bar.Value,
			Style: chart.Style{
				FillColor:   fillColor,
				StrokeColor: darkenColor(fillColor, 0.75),
				StrokeWidth: 2,
			},
		})
	}

	graph := chart.BarChart{
		Title:      title,
		Width:      920,
		Height:     560,
		BarWidth:   112,
		BarSpacing: 28,
		Background: chart.Style{
			FillColor: drawing.ColorFromHex("#FFFFFF"),
			Padding: chart.Box{
				Top:    52,
				Left:   24,
				Right:  24,
				Bottom: 76,
			},
		},
		Canvas: chart.Style{
			FillColor:   drawing.ColorFromHex("#F7FBFB"),
			StrokeColor: drawing.ColorFromHex("#D8E6E7"),
			StrokeWidth: 1,
		},
		TitleStyle: chart.Style{
			FontColor: drawing.ColorFromHex("#14323C"),
			FontSize:  16,
		},
		XAxis: chart.Style{
			FontColor: drawing.ColorFromHex("#35505B"),
			FontSize:  11,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				StrokeColor: drawing.ColorFromHex("#9CB4BA"),
				StrokeWidth: 1,
				FontColor:   drawing.ColorFromHex("#54707A"),
				FontSize:    10,
			},
			TickStyle: chart.Style{
				FontColor: drawing.ColorFromHex("#54707A"),
				FontSize:  10,
			},
			GridMajorStyle: chart.Style{
				StrokeColor: drawing.ColorFromHex("#DFEAEC"),
				StrokeWidth: 1,
			},
			ValueFormatter: formatChartPercentValue,
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: 100,
			},
		},
		Bars: chartBars,
	}

	var content bytes.Buffer
	if err := graph.Render(chart.PNG, &content); err != nil {
		return nil, err
	}
	return content.Bytes(), nil
}

func formatChartPercentValue(value interface{}) string {
	switch typed := value.(type) {
	case float64:
		return fmt.Sprintf("%.0f%%", typed)
	case float32:
		return fmt.Sprintf("%.0f%%", typed)
	case int:
		return fmt.Sprintf("%d%%", typed)
	case int64:
		return fmt.Sprintf("%d%%", typed)
	default:
		return fmt.Sprintf("%v%%", value)
	}
}

func darkenColor(color drawing.Color, factor float64) drawing.Color {
	if factor <= 0 {
		factor = 1
	}
	if factor > 1 {
		factor = 1
	}

	return drawing.Color{
		R: uint8(float64(color.R) * factor),
		G: uint8(float64(color.G) * factor),
		B: uint8(float64(color.B) * factor),
		A: color.A,
	}
}
