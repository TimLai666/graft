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

	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
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

// runWindow opens the live, interactive gallery.
func runWindow(th *graft.Theme) {
	gpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("graft — Kitchen Sink").
		WithSize(1100, 900).
		WithContinuousRender(false))

	uiApp := app.New(
		app.WithWindowProvider(gpuApp),
		app.WithPlatformProvider(gpuApp),
		app.WithEventSource(gpuApp.EventSource()),
		app.WithTheme(th.AsUITheme()),
	)
	if err := graft.Install(uiApp, th); err != nil {
		log.Fatal(err)
	}
	uiApp.SetRoot(graft.ScrollArea(sheet()).W(1100).H(900))

	if err := desktop.Run(gpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
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
					graft.DropdownMenuItem("Profile").Icon(icons.Circle).Shortcut("⇧⌘P"),
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
	).Padding(40).Gap(40)
}
