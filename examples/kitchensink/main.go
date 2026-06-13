// Command kitchensink renders every graft Tier 1 component in one scrollable
// window — the "reference sheet" for visual comparison against ui.shadcn.com.
//
// Run it as a window:
//
//	go run ./examples/kitchensink
//
// Or dump a deterministic PNG of the whole sheet without a window (handy in
// CI or over SSH):
//
//	go run ./examples/kitchensink -png sheet.png        # light
//	go run ./examples/kitchensink -png sheet.png -dark   # dark
package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/graftapp"
	"github.com/TimLai666/graft/icons"
)

func main() {
	pngPath := flag.String("png", "", "render the sheet to this PNG file and exit (no window)")
	dark := flag.Bool("dark", false, "use dark mode")
	flag.Parse()

	th := graft.NewTheme(graft.BaseColor(graft.Neutral), graft.Radius(10))
	if *dark {
		th.SetMode(graft.ModeDark)
	} else {
		th.SetMode(graft.ModeLight)
	}

	if *pngPath != "" {
		renderPNG(th, *pngPath)
		return
	}
	runWindow(th)
}

// renderPNG dumps the sheet headlessly via the offscreen renderer.
func renderPNG(th *graft.Theme, path string) {
	if err := graft.LoadAssets(); err != nil {
		log.Fatal(err)
	}
	// This binary links the GPU SDF accelerator (via graftapp, for the live
	// window). The offscreen renderer is CPU-only and produces a blank image
	// while a GPU accelerator is registered but has no surface, so close it
	// before rendering headlessly. Harmless here: the -png path exits after.
	gg.CloseAccelerator()
	// Components snapshot graft.CurrentTheme() at construction, so make th
	// current before building the sheet (the window path does this via
	// graft.Install).
	graft.SetTheme(th)
	r := offscreen.NewRenderer(0, 0,
		offscreen.WithFitSize(),
		offscreen.WithMaxSize(1100, 6000),
		offscreen.WithTheme(th),
		offscreen.WithBackground(th.Active().Background),
		offscreen.WithScale(2),
	)
	r.Render(sheet())
	img := r.Image()
	if img == nil {
		log.Fatal("kitchensink: renderer produced no image")
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
}

// runWindow opens the live, interactive gallery via the graftapp launcher.
func runWindow(th *graft.Theme) {
	if err := graftapp.New().
		Title("graft — Kitchen Sink").
		Size(1100, 900).
		Theme(th).
		Run(interactiveSheet()); err != nil {
		log.Fatal(err)
	}
}

// interactiveSheet is the window-mode root: real, clickable components
// (triggers, signals, handlers) rather than the static preview widgets used
// for golden images. Returned wrapped in a ScrollArea that fills the window.
func interactiveSheet() widget.Widget {
	selected := state.NewSignal("next")
	dialogOpen := state.NewSignal(false)
	cmdOpen := state.NewSignal(false)

	body := primitives.VBox(
		graft.H1("graft"),
		graft.Lead("Interactive demo — click, type, drag, toggle."),

		section("Buttons (hover + click)",
			row(
				graft.Button("Default").OnClick(func() { dialogOpen.Set(true) }),
				graft.Button("Secondary").Secondary(),
				graft.Button("Destructive").Destructive(),
				graft.Button("Outline").Outline(),
				graft.Button("Ghost").Ghost(),
			),
		),

		section("Form controls (uncontrolled — just click/type/drag)",
			graft.Label("Email"),
			graft.Input().Placeholder("m@example.com").W(320),
			graft.Textarea().Placeholder("Type a message...").W(320),
			row(
				graft.Checkbox().Label("Accept terms"),
				graft.Checkbox().Label("Subscribe").Checked(true),
			),
			row(graft.Switch(), graft.Switch().Checked(true), graft.Switch().Sm()),
			graft.RadioGroup(
				graft.RadioGroupItem("comfortable", "Comfortable"),
				graft.RadioGroupItem("compact", "Compact"),
			).Value("comfortable"),
			graft.Slider().Value(40).W(320),
			row(
				graft.Toggle("Bold"),
				graft.Toggle("Italic"),
				graft.ToggleGroup(
					graft.ToggleGroupItem("left", "Left"),
					graft.ToggleGroupItem("center", "Center"),
					graft.ToggleGroupItem("right", "Right"),
				).Value("center"),
			),
		),

		section("Input OTP",
			graft.InputOTP().Length(6).Groups(3, 3),
		),

		section("Tabs",
			graft.Tabs(
				graft.TabsList(
					graft.TabsTrigger("account", "Account"),
					graft.TabsTrigger("password", "Password"),
				),
				graft.TabsContent("account", graft.MutedText("Make changes to your account here.")),
				graft.TabsContent("password", graft.MutedText("Change your password here.")),
			).Value("account"),
		),

		section("Accordion",
			graft.Accordion(
				graft.AccordionItem("a", "Is it accessible?",
					graft.MutedText("Yes. It adheres to the WAI-ARIA design pattern.")),
				graft.AccordionItem("b", "Is it styled?",
					graft.MutedText("Yes. It comes with default styles.")),
				graft.AccordionItem("c", "Is it animated?",
					graft.MutedText("Yes. Animated by default.")),
			),
		),

		section("Carousel",
			graft.Carousel(
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H2("Slide 1"))).W(300)),
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H2("Slide 2"))).W(300)),
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H2("Slide 3"))).W(300)),
			),
		),

		section("Resizable panels (drag the divider)",
			graft.ResizablePanelGroup(graft.ResizableHorizontal,
				graft.ResizablePanel(
					graft.Card(graft.CardContent(graft.MutedText("Panel A"))).W(200),
				).DefaultSize(0.5),
				graft.ResizableHandle().WithHandle(),
				graft.ResizablePanel(
					graft.Card(graft.CardContent(graft.MutedText("Panel B"))).W(200),
				),
			),
		),

		section("Charts",
			graft.LineChart(
				graft.LineSeries("desktop", 186, 305, 237, 273, 209, 214).Dots(true),
				graft.LineSeries("mobile", 80, 200, 120, 190, 130, 140),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").
				Legend(true).W(500).H(250),
		),

		section("Overlays (click to open)",
			row(
				graft.Select(
					graft.SelectItem("next", "Next.js"),
					graft.SelectItem("svelte", "SvelteKit"),
					graft.SelectItem("astro", "Astro"),
				).Bind(selected).W(200),

				graft.DropdownMenu(
					graft.DropdownMenuTrigger(graft.Button("Open menu").Outline()),
					graft.DropdownMenuContent(
						graft.DropdownMenuLabel("My Account"),
						graft.DropdownMenuItem("Profile").Icon(icons.Circle),
						graft.DropdownMenuItem("Settings").Icon(icons.Search),
						graft.DropdownMenuSeparator(),
						graft.DropdownMenuItem("Log out").Destructive(),
					),
				),

				graft.Popover(
					graft.Button("Open popover").Outline(),
					graft.PopoverContent(
						graft.H4("Dimensions"),
						graft.MutedText("Set the dimensions for the layer."),
						graft.Input().Placeholder("Width").W(200),
					),
				),

				graft.DialogTrigger(graft.Button("Edit profile"), dialogOpen),
				graft.Dialog(
					graft.DialogContent(
						graft.DialogHeader(
							graft.DialogTitle("Edit profile"),
							graft.DialogDescription("Make changes to your profile here."),
						),
						graft.Input().Placeholder("Name").W(360),
						graft.DialogFooter(
							graft.Button("Cancel").Outline().OnClick(func() { dialogOpen.Set(false) }),
							graft.Button("Save changes").OnClick(func() {
								dialogOpen.Set(false)
							}),
						),
					),
				).Bind(dialogOpen),

				graft.Button("Command palette").Outline().OnClick(func() { cmdOpen.Set(true) }),
				graft.CommandDialog(
					graft.CommandInput().Placeholder("Type a command or search..."),
					graft.CommandList(
						graft.CommandGroup("Suggestions",
							graft.CommandItem("Calendar").Icon(icons.Calendar),
							graft.CommandItem("Search").Icon(icons.Search),
						),
						graft.CommandSeparator(),
						graft.CommandGroup("Settings",
							graft.CommandItem("Profile").Icon(icons.Circle),
							graft.CommandItem("Billing"),
						),
					),
				).Bind(cmdOpen),
			),
		),
	).Padding(40).Gap(40)

	return body
}

// section wraps a titled group of demos.
func section(title string, children ...widget.Widget) widget.Widget {
	items := append([]widget.Widget{graft.H3(title)}, children...)
	return primitives.VBox(items...).Gap(16)
}

// row lays demos out horizontally.
func row(children ...widget.Widget) widget.Widget {
	return primitives.HBox(children...).Gap(12).CrossAlign(primitives.CrossAxisCenter)
}

// col lays demos out vertically.
func col(children ...widget.Widget) widget.Widget {
	return primitives.VBox(children...).Gap(12)
}

// sheet builds the full reference sheet of every Tier 1 component.
func sheet() widget.Widget {
	return primitives.VBox(
		graft.H1("graft"),
		graft.Lead("A pixel-faithful shadcn/ui for Go."),

		section("Typography",
			graft.H2("The quick brown fox"),
			graft.P("Body copy renders in Geist at 16px with comfortable leading."),
			graft.MutedText("Muted secondary text."),
			graft.InlineCode("go get github.com/TimLai666/graft"),
		),

		section("Buttons",
			row(
				graft.Button("Default"),
				graft.Button("Secondary").Secondary(),
				graft.Button("Destructive").Destructive(),
				graft.Button("Outline").Outline(),
				graft.Button("Ghost").Ghost(),
				graft.Button("Link").Link(),
			),
			row(
				graft.Button("Small").Sm(),
				graft.Button("Default"),
				graft.Button("Large").Lg(),
				graft.Button("With icon").Icon(icons.Search),
				graft.Button("Disabled").Disabled(true),
			),
		),

		section("Badges",
			row(
				graft.Badge("Default"),
				graft.Badge("Secondary").Secondary(),
				graft.Badge("Destructive").Destructive(),
				graft.Badge("Outline").Outline(),
				graft.Badge("Verified").Icon(icons.CircleCheck),
			),
		),

		section("Inputs & selection",
			col(
				graft.Label("Email"),
				graft.Input().Placeholder("m@example.com").W(320),
				row(
					graft.Checkbox().Label("Accept terms").Checked(true),
					graft.Checkbox().Label("Disabled").Disabled(true),
				),
				row(
					graft.Switch().Checked(true),
					graft.Switch(),
					graft.Switch().Sm().Checked(true),
				),
				graft.RadioGroup(
					graft.RadioGroupItem("comfortable", "Comfortable"),
					graft.RadioGroupItem("compact", "Compact"),
				),
				graft.Slider().Value(50).W(320),
				graft.Select(
					graft.SelectItem("apple", "Apple"),
					graft.SelectItem("banana", "Banana"),
				).Placeholder("Select a fruit").W(220),
			),
		),

		section("Feedback",
			col(
				primitives.Box(graft.Progress().Value(66)).Width(320),
				row(graft.Spinner(), graft.Skeleton().W(160).Rounded(6)),
				graft.Alert(
					graft.AlertTitle("Heads up!"),
					graft.AlertDescription("You can add components to your app."),
				).Icon(icons.Info),
				graft.Alert(
					graft.AlertTitle("Unable to process your payment."),
					graft.AlertDescription("Please verify your billing information."),
				).Icon(icons.CircleAlert).Destructive(),
			),
		),

		section("Tabs",
			graft.Tabs(
				graft.TabsList(
					graft.TabsTrigger("account", "Account"),
					graft.TabsTrigger("password", "Password"),
				),
				graft.TabsContent("account", graft.MutedText("Make changes to your account here.")),
				graft.TabsContent("password", graft.MutedText("Change your password here.")),
			).Value("account"),
		),

		section("Card",
			graft.Card(
				graft.CardHeader(
					graft.CardTitle("Create project"),
					graft.CardDescription("Deploy your new project in one click."),
				),
				graft.CardContent(
					graft.Input().Placeholder("Project name").W(300),
				),
				graft.CardFooter(
					graft.Button("Cancel").Outline(),
					graft.Button("Deploy"),
				),
			).W(360),
		),

		section("Overlays (content previews)",
			row(
				graft.DropdownMenuPreview(graft.DropdownMenuContent(
					graft.DropdownMenuLabel("My Account"),
					graft.DropdownMenuItem("Profile").Icon(icons.Circle).Shortcut("Ctrl P"),
					graft.DropdownMenuItem("Settings").Icon(icons.Search),
					graft.DropdownMenuSeparator(),
					graft.DropdownMenuItem("Log out").Destructive(),
				)),
				graft.SelectMenuPreview(graft.Select(
					graft.SelectGroup("Fruits",
						graft.SelectItem("apple", "Apple"),
						graft.SelectItem("banana", "Banana"),
					),
				).Value("banana")),
			),
		),

		section("Disclosure",
			graft.Accordion(
				graft.AccordionItem("a", "Is it accessible?",
					graft.MutedText("Yes. It adheres to the WAI-ARIA design pattern.")),
				graft.AccordionItem("b", "Is it styled?",
					graft.MutedText("Yes. It comes with default styles.")),
				graft.AccordionItem("c", "Is it animated?",
					graft.MutedText("Yes. Animated by default.")),
			).Value("a"),
		),

		section("Toggles & keys",
			row(
				graft.Toggle("Bold"),
				graft.Toggle("Italic").On(true),
				graft.Toggle("Underline").Outline(),
			),
			graft.ToggleGroup(
				graft.ToggleGroupItem("left", "Left"),
				graft.ToggleGroupItem("center", "Center"),
				graft.ToggleGroupItem("right", "Right"),
			).Value("center"),
			row(graft.Kbd("Ctrl", "K"), graft.Kbd("Esc")),
		),

		section("Table",
			graft.Table(
				graft.TableHeader(graft.TableRow(
					graft.TableHead("Invoice"), graft.TableHead("Status"), graft.TableHead("Amount"),
				)),
				graft.TableBody(
					graft.TableRow(graft.TableCell("INV001"), graft.TableCell("Paid"), graft.TableCell("$250.00")),
					graft.TableRow(graft.TableCell("INV002"), graft.TableCell("Pending"), graft.TableCell("$150.00")),
					graft.TableRow(graft.TableCell("INV003"), graft.TableCell("Unpaid"), graft.TableCell("$350.00")),
				),
			),
		),

		section("Navigation & composition",
			graft.Breadcrumb(
				graft.BreadcrumbLink("Home"),
				graft.BreadcrumbLink("Components"),
				graft.BreadcrumbPage("Breadcrumb"),
			),
			graft.Pagination().Pages(10, 2),
			graft.ButtonGroup(
				graft.Button("Years").Outline(),
				graft.Button("Months").Outline(),
				graft.Button("Days").Outline(),
			),
			graft.InputGroup().
				Leading(graft.InputGroupAddon(icons.Search)).
				Placeholder("Search...").W(320),
		),

		section("Text input",
			graft.Textarea().Placeholder("Type your message here.").W(360),
		),

		section("Sheet (right, settled open)",
			graft.SheetPreview(
				graft.SheetContent(
					graft.SheetHeader(
						graft.SheetTitle("Edit profile"),
						graft.SheetDescription("Make changes to your profile here."),
					),
					graft.SheetFooter(graft.Button("Save changes")),
				),
				geometry.Sz(560, 320),
			),
		),

		section("Calendar & date picker",
			row(
				graft.Calendar().
					Month(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)).
					Selected(time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)),
				graft.DatePickerContentPreview(
					time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC),
					graft.CurrentTheme(),
				),
			),
		),

		section("Menus & cards (content previews)",
			row(
				graft.ContextMenuPreview(graft.ContextMenuContent(
					graft.ContextMenuItem("Back"),
					graft.ContextMenuItem("Forward").Disabled(true),
					graft.ContextMenuSeparator(),
					graft.ContextMenuItem("Reload").Shortcut("Ctrl R"),
				)),
				graft.MenubarMenuPreview(graft.MenubarContent(
					graft.MenubarItem("New Tab").Shortcut("Ctrl T"),
					graft.MenubarItem("New Window"),
					graft.MenubarSeparator(),
					graft.MenubarItem("Print").Shortcut("Ctrl P"),
				)),
				graft.ComboboxContentPreview(
					graft.Combobox(
						graft.ComboboxItem("next", "Next.js"),
						graft.ComboboxItem("svelte", "SvelteKit"),
						graft.ComboboxItem("astro", "Astro"),
					),
					"", "next", graft.CurrentTheme(),
				),
			),
		),

		section("Toasts",
			graft.ToastStack(
				graft.ToastCard("Your message was sent", graft.ToastSuccessOpt()),
				graft.ToastCard("Could not save changes",
					graft.ToastDescription("There was a problem with your request."),
					graft.ToastErrorOpt()),
			),
		),

		section("Input OTP",
			graft.InputOTP().Length(6).Groups(3, 3).SetValue("123456"),
		),

		section("Carousel",
			graft.Carousel(
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H3("Slide 1"))).W(280)),
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H3("Slide 2"))).W(280)),
				graft.CarouselItem(graft.Card(graft.CardContent(graft.H3("Slide 3"))).W(280)),
			),
		),

		section("Resizable panels",
			graft.ResizablePanelGroup(graft.ResizableHorizontal,
				graft.ResizablePanel(graft.MutedText("Panel A")).DefaultSize(0.5),
				graft.ResizableHandle().WithHandle(),
				graft.ResizablePanel(graft.MutedText("Panel B")),
			),
		),

		section("Charts",
			graft.LineChart(
				graft.LineSeries("desktop", 186, 305, 237, 273, 209, 214).Dots(true),
				graft.LineSeries("mobile", 80, 200, 120, 190, 130, 140),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").
				Legend(true).W(500).H(250),
			graft.BarChart(
				graft.BarSeries("revenue", 450, 380, 520, 490, 610, 580),
			).Categories("Jan", "Feb", "Mar", "Apr", "May", "Jun").W(500).H(200),
		),

		section("Command palette (preview)",
			graft.CommandPreview(
				graft.Command(
					graft.CommandInput().Placeholder("Type a command or search..."),
					graft.CommandList(
						graft.CommandGroup("Suggestions",
							graft.CommandItem("Calendar").Icon(icons.Calendar),
							graft.CommandItem("Search").Icon(icons.Search).Shortcut("Ctrl K"),
						),
						graft.CommandSeparator(),
						graft.CommandGroup("Settings",
							graft.CommandItem("Profile").Icon(icons.Circle),
							graft.CommandItem("Billing"),
						),
					),
				), "", 0, graft.CurrentTheme(),
			),
		),

		section("Field & form",
			graft.Field(
				graft.Label("Username"),
				graft.Input().Placeholder("Enter username").W(300),
				graft.MutedText("This is your public display name."),
			),
		),

		section("Empty state",
			primitives.Box(graft.Empty(
				graft.H4("No results"),
				graft.MutedText("Try adjusting your search or filters."),
			)).Width(360),
		),

		section("Item",
			graft.ItemGroup(
				graft.Item(
					graft.ItemMedia(icons.Circle).Icon(),
					graft.ItemContent(
						graft.ItemTitle("Notifications"),
						graft.ItemDescription("Configure how you receive notifications."),
					),
					graft.ItemActions(graft.Switch().Checked(true)),
				).Outline().W(420),
				graft.ItemSeparator(),
				graft.Item(
					graft.ItemMedia(icons.Search).Icon(),
					graft.ItemContent(
						graft.ItemTitle("Search"),
						graft.ItemDescription("Find anything across your workspace."),
					),
					graft.ItemActions(graft.Kbd("Ctrl", "K")),
				).Muted().W(420),
			),
		),

		section("Aspect ratio",
			graft.AspectRatio(16.0/9.0,
				graft.Card(graft.CardContent(graft.MutedText("16:9 content area"))).W(320),
			),
		),
	).Padding(40).Gap(40)
}
