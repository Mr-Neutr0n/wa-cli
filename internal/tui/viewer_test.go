package tui

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codebyoketch/wa-cli/internal/termimg"
)

func writePNG(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0x80
	}
	path := filepath.Join(t.TempDir(), "photo.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

// Every protocol termimg knows has a drawer, and "m" visits each once.
func TestEveryProtocolHasADrawer(t *testing.T) {
	seen := map[termimg.Protocol]bool{}
	p := termimg.Blocks
	for range photoDrawers {
		if seen[p] {
			t.Fatalf("cycle returned to %s early", p)
		}
		seen[p] = true
		p = nextProtocol(p)
	}
	if p != termimg.Blocks {
		t.Errorf("cycle doesn't return to blocks")
	}
	for _, name := range []string{"blocks", "sixel", "iterm", "kitty"} {
		p, ok := termimg.ParseProtocol(name)
		if !ok || int(p) >= len(photoDrawers) || photoDrawers[p] == nil || !seen[p] {
			t.Errorf("%s has no drawer", name)
		}
	}
}

func TestBlockDrawerReturnsText(t *testing.T) {
	msg := blockDrawer{}.draw(photoJob{path: writePNG(t, 40, 20), cols: 20, rows: 5})
	if msg.err != nil || msg.info.Width != 40 {
		t.Fatalf("draw: %+v", msg)
	}
	if lines := strings.Count(msg.text, "\n") + 1; lines > 5 || msg.text == "" {
		t.Errorf("drew %d lines for 5 rows", lines)
	}
}

// A pixel drawer scales the photo to the screen (never up), centres it,
// and leaves the cursor where it was.
func TestPixelDrawerCentresAndRestoresCursor(t *testing.T) {
	var out bytes.Buffer
	var got image.Rectangle
	d := pixelDrawer{proto: termimg.Sixel, encode: func(w io.Writer, img image.Image) error {
		got = img.Bounds()
		_, err := io.WriteString(w, "PIXELS")
		return err
	}}
	job := photoJob{out: &out, path: writePNG(t, 400, 200), cols: 80, rows: 24, started: time.Now().Add(-time.Second)}
	if msg := d.draw(job); msg.err != nil || msg.text != "" {
		t.Fatalf("draw: %+v", msg)
	}
	if got.Dx() != 400 || got.Dy() != 200 {
		t.Errorf("encoded %v, want the photo's own 400x200", got)
	}
	// 400x200 px is 40x10 cells of 10x20 px: centred in 80x24 at column 21, row 8.
	if want := "\x1b7\x1b[8;21HPIXELS\x1b8"; out.String() != want {
		t.Errorf("wrote %q, want %q", out.String(), want)
	}
}

func TestOnlyKittyNeedsClearing(t *testing.T) {
	for p, d := range photoDrawers {
		var out bytes.Buffer
		d.clear(&out)
		if wantClear := termimg.Protocol(p) == termimg.Kitty; (out.Len() > 0) != wantClear {
			t.Errorf("%s clear wrote %q", termimg.Protocol(p), out.String())
		}
	}
}
