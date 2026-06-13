package metrics

// DatePicker holds the pixel metrics for the shadcn DatePicker composition
// (docs/research — DatePicker is the canonical shadcn recipe combining
// Popover + Button + Calendar). Source markup:
//
//	<Popover>
//	  <PopoverTrigger asChild>
//	    <Button variant="outline"
//	      className="w-[240px] justify-start text-left font-normal
//	                 data-[empty=true]:text-muted-foreground">
//	      <CalendarIcon />     // mr-2 h-4 w-4
//	      {date ? format(date, "PPP") : <span>Pick a date</span>}
//	    </Button>
//	  </PopoverTrigger>
//	  <PopoverContent className="w-auto p-0">
//	    <Calendar mode="single" selected={date} onSelect={setDate} />
//	  </PopoverContent>
//	</Popover>
//
// The trigger is an outline Button widened to 240px, left-aligned, with a
// leading 16px calendar icon; the placeholder text uses --muted-foreground.
// The popover content has no padding (p-0, w-auto) so the calendar sizes the
// surface; it keeps the popover border, radius (rounded-md), and shadow-md.
var DatePicker = struct {
	// TriggerWidth is the default trigger width (w-[240px] = 240px).
	TriggerWidth float32
	// IconSize is the leading calendar icon size (h-4 w-4 = 16px).
	IconSize float32
	// ContentPadding is the popover content padding (p-0 = 0px; the
	// calendar carries its own internal padding via the wrapper below).
	ContentPadding float32
	// ContentInset is the breathing room added around the calendar inside
	// the otherwise-padding-free popover (matches react-day-picker's p-3).
	ContentInset float32
	// DefaultFormat is the Go time layout used to render the selected date
	// when no custom format is set ("PPP" ≈ "January 2, 2006").
	DefaultFormat string
	// Placeholder is the empty-state label.
	Placeholder string
}{
	TriggerWidth:   240,
	IconSize:       16,
	ContentPadding: 0,
	ContentInset:   12, // react-day-picker root p-3
	DefaultFormat:  "January 2, 2006",
	Placeholder:    "Pick a date",
}
