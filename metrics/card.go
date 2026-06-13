package metrics

// Card metrics.
//
// Source: shadcn new-york-v4 card.tsx:
//
//	Card:        "flex flex-col gap-6 rounded-xl border bg-card py-6
//	              text-card-foreground shadow-sm"
//	CardHeader:  "@container/card-header grid auto-rows-min
//	              grid-rows-[auto_auto] items-start gap-2 px-6
//	              has-data-[slot=card-action]:grid-cols-[1fr_auto]
//	              [.border-b]:pb-6"
//	CardTitle:   "leading-none font-semibold"
//	CardDescription: "text-sm text-muted-foreground"
//	CardAction:  "col-start-2 row-span-2 row-start-1 self-start
//	              justify-self-end"
//	CardContent: "px-6"
//	CardFooter:  "flex items-center px-6 [.border-t]:pt-6"
//
// The card radius is rounded-xl, drawn with the theme RadiusXL (14px at
// the default --radius of 10). The card surface uses shadow-sm (ShadowSM).
const (
	// CardPadY is the card's vertical padding in px (py-6).
	CardPadY float32 = 24

	// CardGap is the vertical gap between card sections in px (gap-6).
	CardGap float32 = 24

	// CardBorderWidth is the card border width in px (border).
	CardBorderWidth float32 = 1

	// CardSectionPadX is the horizontal padding of every card section in
	// px (px-6 on CardHeader/CardContent/CardFooter).
	CardSectionPadX float32 = 24

	// CardHeaderGap is the gap between title and description in px
	// (gap-2).
	CardHeaderGap float32 = 8

	// CardTitleFontSize is the title font size in px (inherits text-base
	// 16px from the card body).
	CardTitleFontSize float32 = 16

	// CardTitleFontWeight is the title font weight (font-semibold).
	CardTitleFontWeight = 600

	// CardTitleLineHeight is the title line height in px (leading-none =
	// font size).
	CardTitleLineHeight float32 = 16

	// CardDescriptionFontSize is the description font size in px
	// (text-sm).
	CardDescriptionFontSize float32 = 14

	// CardDescriptionLineHeight is the description line height in px
	// (text-sm = 14px/20px).
	CardDescriptionLineHeight float32 = 20
)
