package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/signintech/gopdf"
)

type upstreamBillingPDFAllocationRow struct {
	row      UpstreamBillingChannelUserRow
	identity string
	height   float64
}

// User display names can contain characters outside the embedded billing font.
// Preserve those identities as explicit Unicode escapes in the PDF rather than
// making an otherwise valid financial archive impossible to generate.
func paginateUpstreamBillingAllocationsPDF(pdf *gopdf.GoPdf, missingGlyph *rune, rows []UpstreamBillingChannelUserRow) ([][]upstreamBillingPDFAllocationRow, error) {
	if err := pdf.SetFont("billing", "", 8.5); err != nil {
		return nil, err
	}
	pages := [][]upstreamBillingPDFAllocationRow{{}}
	height := 0.0
	for _, row := range rows {
		name := fmt.Sprintf("#%d %s / 用户 #%d %s", row.ChannelID, row.ChannelName, row.UserID, row.Username)
		if row.DisplayName != "" && row.DisplayName != row.Username {
			name += " / " + row.DisplayName
		}
		var identity strings.Builder
		for _, character := range name {
			*missingGlyph = 0
			if _, err := pdf.MeasureTextWidth(string(character)); err != nil {
				return nil, err
			}
			if *missingGlyph != 0 {
				fmt.Fprintf(&identity, "[U+%04X]", character)
			} else {
				identity.WriteRune(character)
			}
		}
		*missingGlyph = 0
		lines, err := pdf.SplitText(identity.String(), 150)
		if err != nil {
			return nil, err
		}
		rowHeight := math.Max(66, float64(len(lines))*8.5*1.45+47)
		if rowHeight > 490 {
			return nil, errors.New("upstream user identity exceeds PDF page space")
		}
		if height+rowHeight > 490 && len(pages[len(pages)-1]) > 0 {
			pages = append(pages, []upstreamBillingPDFAllocationRow{})
			height = 0
		}
		pages[len(pages)-1] = append(pages[len(pages)-1], upstreamBillingPDFAllocationRow{row: row, identity: identity.String(), height: rowHeight})
		height += rowHeight
	}
	return pages, nil
}
