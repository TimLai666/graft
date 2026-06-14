package metrics

// Card metrics.
//
// Source: current ui.shadcn.com live card.tsx (radix era, newer than
// new-york-v4):
//
//	Card:        "flex flex-col gap-4 rounded-xl border bg-card py-4
//	              text-card-foreground shadow-sm"
//	CardHeader:  "@container/card-header grid auto-rows-min
//	              grid-rows-[auto_auto] items-start gap-1 px-4
//	              has-data-[slot=card-action]:grid-cols-[1fr_auto]
//	              [.border-b]:pb-4"
//	CardTitle:   "leading-snug font-medium"
//	CardDescription: "text-sm text-muted-foreground"
//	CardAction:  "col-start-2 row-span-2 row-start-1 self-start
//	              justify-self-end"
//	CardContent: "px-4"
//	CardFooter:  "flex items-center px-4 [.border-t]:pt-4"
//
// The card radius is rounded-xl, drawn with the theme RadiusXL (14px at
// the default --radius of 10). The card surface uses shadow-sm (ShadowSM).
const (
	// CardPadY is the card's vertical padding in px (py-4).
	CardPadY float32 = 16

	// CardGap is the vertical gap between card sections in px (gap-4).
	CardGap float32 = 16

	// CardBorderWidth is the card border width in px (border).
	CardBorderWidth float32 = 1

	// CardSectionPadX is the horizontal padding of every card section in
	// px (px-4 on CardHeader/CardContent/CardFooter).
	CardSectionPadX float32 = 16

	// CardHeaderGap is the gap between title and description in px
	// (gap-1).
	CardHeaderGap float32 = 4

	// CardTitleFontSize is the title font size in px (text-base 16px).
	CardTitleFontSize float32 = 16

	// CardTitleFontWeight is the title font weight (font-medium).
	CardTitleFontWeight = 500

	// CardTitleLineHeight is the title line height in px (leading-snug =
	// 1.375 × 16px ≈ 22px).
	CardTitleLineHeight float32 = 22

	// CardDescriptionFontSize is the description font size in px
	// (text-sm).
	CardDescriptionFontSize float32 = 14

	// CardDescriptionLineHeight is the description line height in px
	// (text-sm = 14px/20px).
	CardDescriptionLineHeight float32 = 20
)
