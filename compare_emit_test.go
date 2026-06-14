package graft_test

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/icons"
)

// emitEnv, when set to a directory path, makes TestEmitCompareImages render
// each comparison case to <dir>/graft/<name>.png. It is a tooling test (not a
// pass/fail assertion): it exists so the graft-side comparison images are
// produced by the same GPU-free offscreen path the golden harness uses — the
// kitchensink binary cannot render headless PNGs because its GPU SDF
// accelerator (required for the live window) blanks the offscreen renderer.
//
//	$env:GRAFT_EMIT_COMPARE = "compare"; go test -run TestEmitCompareImages .
const emitEnv = "GRAFT_EMIT_COMPARE"

type compareCase struct {
	name  string
	build func() widget.Widget
}

// compareCases is the curated set of visually load-bearing components for
// side-by-side validation against real ui.shadcn.com screenshots, at the
// default Neutral theme / radius 0.625rem (shadcn's default light preview).
var compareCases = []compareCase{
	{"button-default", func() widget.Widget { return graft.Button("Button") }},
	{"button-secondary", func() widget.Widget { return graft.Button("Secondary").Secondary() }},
	{"button-destructive", func() widget.Widget { return graft.Button("Destructive").Destructive() }},
	{"button-outline", func() widget.Widget { return graft.Button("Outline").Outline() }},
	{"button-ghost", func() widget.Widget { return graft.Button("Ghost").Ghost() }},
	{"button-link", func() widget.Widget { return graft.Button("Link").Link() }},
	{"badge-default", func() widget.Widget { return graft.Badge("Badge") }},
	{"badge-secondary", func() widget.Widget { return graft.Badge("Secondary").Secondary() }},
	{"badge-destructive", func() widget.Widget { return graft.Badge("Destructive").Destructive() }},
	{"badge-outline", func() widget.Widget { return graft.Badge("Outline").Outline() }},
	{"input", func() widget.Widget { return graft.Input().Placeholder("Email").W(280) }},
	{"checkbox", func() widget.Widget { return graft.Checkbox().Label("Accept terms and conditions").Checked(true) }},
	{"switch-on", func() widget.Widget { return graft.Switch().Checked(true) }},
	{"switch-off", func() widget.Widget { return graft.Switch() }},
	{"radiogroup", func() widget.Widget {
		return graft.RadioGroup(
			graft.RadioGroupItem("a", "Default"),
			graft.RadioGroupItem("b", "Comfortable"),
		).Value("a")
	}},
	{"slider", func() widget.Widget { return graft.Slider().Value(50).W(280) }},
	{"progress", func() widget.Widget { return primitives.Box(graft.Progress().Value(60)).Width(280) }},
	{"label", func() widget.Widget { return graft.Label("Email address") }},
	{"separator", func() widget.Widget { return primitives.Box(graft.Separator()).Width(280) }},
	{"avatar", func() widget.Widget { return graft.Avatar(graft.AvatarFallback("CN")) }},
	{"alert", func() widget.Widget {
		return primitives.Box(graft.Alert(
			graft.AlertTitle("Heads up!"),
			graft.AlertDescription("You can add components to your app using the cli."),
		).Icon(icons.Info)).Width(420)
	}},
	{"card", func() widget.Widget {
		return graft.Card(
			graft.CardHeader(
				graft.CardTitle("Create project"),
				graft.CardDescription("Deploy your new project in one-click."),
			),
			graft.CardContent(graft.Input().Placeholder("Name of your project").W(300)),
			graft.CardFooter(
				graft.Button("Cancel").Outline(),
				graft.Button("Deploy"),
			),
		).W(360)
	}},
	{"tabs", func() widget.Widget {
		return graft.Tabs(
			graft.TabsList(
				graft.TabsTrigger("account", "Account"),
				graft.TabsTrigger("password", "Password"),
			),
			graft.TabsContent("account", graft.MutedText("Make changes to your account here.")),
			graft.TabsContent("password", graft.MutedText("Change your password here.")),
		).Value("account")
	}},
	{"accordion", func() widget.Widget {
		return primitives.Box(graft.Accordion(
			graft.AccordionItem("a", "Is it accessible?",
				graft.MutedText("Yes. It adheres to the WAI-ARIA design pattern.")),
			graft.AccordionItem("b", "Is it styled?",
				graft.MutedText("Yes. It comes with default styles.")),
		).Value("a")).Width(380)
	}},
	{"select-menu", func() widget.Widget {
		return graft.SelectMenuPreview(graft.Select(
			graft.SelectGroup("Fruits",
				graft.SelectItem("apple", "Apple"),
				graft.SelectItem("banana", "Banana"),
				graft.SelectItem("blueberry", "Blueberry"),
			),
		).Value("apple"))
	}},
	{"dropdown-menu", func() widget.Widget {
		return graft.DropdownMenuPreview(graft.DropdownMenuContent(
			graft.DropdownMenuLabel("My Account"),
			graft.DropdownMenuItem("Profile").Icon(icons.Circle).Shortcut("Ctrl P"),
			graft.DropdownMenuItem("Settings").Icon(icons.Search),
			graft.DropdownMenuSeparator(),
			graft.DropdownMenuItem("Log out").Destructive(),
		))
	}},
	{"badge-icon", func() widget.Widget { return graft.Badge("Verified").Icon(icons.CircleCheck) }},
	{"textarea", func() widget.Widget { return graft.Textarea().Placeholder("Type your message here.").W(360) }},
	{"spinner", func() widget.Widget { return graft.Spinner() }},
	{"skeleton", func() widget.Widget { return graft.Skeleton().W(240).Rounded(6) }},
	{"item", func() widget.Widget {
		return graft.Item(
			graft.ItemMedia(icons.Circle).Icon(),
			graft.ItemContent(
				graft.ItemTitle("Notifications"),
				graft.ItemDescription("Configure how you receive notifications."),
			),
			graft.ItemActions(graft.Switch().Checked(true)),
		).Outline().W(420)
	}},
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func TestEmitCompareImages(t *testing.T) {
	dir := os.Getenv(emitEnv)
	if dir == "" {
		t.Skipf("set %s=<dir> to emit comparison PNGs", emitEnv)
	}

	if err := graft.LoadAssets(); err != nil {
		t.Fatalf("loading assets: %v", err)
	}
	// If a theme CSS is present (e.g. the extracted live ui.shadcn.com
	// "theme-default" palette), render in THAT theme so the side-by-side is an
	// apples-to-apples color comparison. Otherwise use the default Neutral.
	th := graft.CurrentTheme()
	if cssPath := filepath.Join(dir, "shadcn-theme.css"); fileExists(cssPath) {
		css, err := os.ReadFile(cssPath)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := graft.ParseThemeCSS(string(css))
		if err != nil {
			t.Fatalf("parsing %s: %v", cssPath, err)
		}
		th = parsed
		graft.SetTheme(th)
		t.Logf("using theme from %s", cssPath)
	}
	th.SetMode(graft.ModeLight)

	outDir := filepath.Join(dir, "graft")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, c := range compareCases {
		r := offscreen.NewRenderer(0, 0,
			offscreen.WithFitSize(),
			offscreen.WithMaxSize(900, 4000),
			offscreen.WithTheme(th),
			offscreen.WithBackground(th.Active().Background),
			offscreen.WithScale(2),
		)
		r.Render(c.build())
		img := r.Image()
		if img == nil {
			t.Errorf("%s: nil image", c.name)
			continue
		}
		path := filepath.Join(outDir, c.name+".png")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Logf("wrote %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	}
}
