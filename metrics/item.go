package metrics

// Item metrics.
//
// Source: shadcn new-york-v4 item.tsx (itemVariants + slot classes):
//
//	Item:        "group/item flex items-center border border-transparent
//	              text-sm rounded-md transition-colors [a&]:hover:bg-accent/50"
//	  variant default: (transparent border)
//	  variant outline: "border-border"
//	  variant muted:   "bg-muted/50"
//	  size default:    "gap-4 p-4"
//	  size sm:         "py-3 px-4 gap-2.5"
//	ItemGroup:    "group/item-group flex flex-col"
//	ItemSeparator:"my-0 ... bg-border h-px" (reuses Separator)
//	ItemMedia:    "flex shrink-0 items-center justify-center gap-2
//	               [&_svg]:pointer-events-none"
//	  variant default: (plain)
//	  variant icon:    "size-8 rounded-sm border bg-muted [&_svg]:size-4"
//	  variant image:   "size-10 rounded-sm overflow-hidden [&_img]:size-full"
//	ItemContent:  "flex flex-1 flex-col gap-1 [&+[data-slot=item-actions]]:flex-none"
//	ItemTitle:    "flex w-fit items-center gap-2 text-sm leading-snug
//	               font-medium"
//	ItemDescription: "text-muted-foreground line-clamp-2 text-sm
//	               leading-normal font-normal"
//	ItemActions:  "flex items-center gap-2"
//
// The Item radius routes through the theme (rounded-md → t.RadiusMD(); the
// icon/image media chip rounded-sm → t.RadiusSM()); only the literal px
// values live here. The muted variant fills bg-muted/50 (ItemMutedBgAlpha).
const (
	// ItemPadDefault is the default-size padding on all sides in px (p-4).
	ItemPadDefault float32 = 16

	// ItemPadXSm is the small-size horizontal padding in px (sm: px-4).
	ItemPadXSm float32 = 16

	// ItemPadYSm is the small-size vertical padding in px (sm: py-3).
	ItemPadYSm float32 = 12

	// ItemGapDefault is the default-size gap between media/content/actions
	// in px (gap-4).
	ItemGapDefault float32 = 16

	// ItemGapSm is the small-size gap between media/content/actions in px
	// (sm: gap-2.5).
	ItemGapSm float32 = 10

	// ItemContentGap is the vertical gap between title and description in
	// px (ItemContent gap-1).
	ItemContentGap float32 = 4

	// ItemActionsGap is the horizontal gap between trailing action widgets
	// in px (ItemActions gap-2).
	ItemActionsGap float32 = 8

	// ItemMediaGap is the gap between media-slot children in px
	// (ItemMedia gap-2).
	ItemMediaGap float32 = 8

	// ItemBorderWidth is the Item border width in px (border = 1px).
	ItemBorderWidth float32 = 1

	// ItemIconChipSize is the icon-media chip side length in px
	// (ItemMedia variant icon: size-8 = 32px).
	ItemIconChipSize float32 = 32

	// ItemMediaIconSize is the lucide icon side length inside the icon chip
	// in px ([&_svg]:size-4 = 16px).
	ItemMediaIconSize float32 = 16

	// ItemImageSize is the image-media side length in px
	// (ItemMedia variant image: size-10 = 40px).
	ItemImageSize float32 = 40

	// ItemTitleFontSize is the title font size in px (text-sm = 14px).
	ItemTitleFontSize float32 = 14

	// ItemTitleFontWeight is the title font weight (font-medium = 500).
	ItemTitleFontWeight = 500

	// ItemTitleLineHeight is the title line box height in px (leading-snug
	// ≈ 1.375 × 14 ≈ 20px; rounded to the text-sm 20px line).
	ItemTitleLineHeight float32 = 20

	// ItemDescriptionFontSize is the description font size in px
	// (text-sm = 14px).
	ItemDescriptionFontSize float32 = 14

	// ItemDescriptionFontWeight is the description font weight
	// (font-normal = 400).
	ItemDescriptionFontWeight = 400

	// ItemDescriptionLineHeight is the description line box height in px
	// (leading-normal = 1.5 × 14 ≈ 20px text-sm line).
	ItemDescriptionLineHeight float32 = 20
)

// ItemMutedBgAlpha is the alpha applied to the Muted token for the muted
// variant background (bg-muted/50 = 50%).
const ItemMutedBgAlpha float32 = 0.5
