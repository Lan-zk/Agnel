// Package font embeds and registers the bilingual reading typeface
// ("Bible Literata": Literata + Source Han Serif SC + STIX Two IPA,
// licensed under SIL Open Font License 1.1) into the current process.
package font

import (
	_ "embed"
	"sync"

	"github.com/vivaui/viva/app"
)

// Family is the typographic family name registered with the platform text engine.
const Family = "Bible Literata"

var (
	//go:embed bible-literata-regular.ttf
	regularTTF []byte

	//go:embed bible-literata-bold.ttf
	boldTTF []byte

	//go:embed bible-literata-italic.ttf
	italicTTF []byte

	once       sync.Once
	registered = Family
)

// Register makes the embedded Regular, Bold, and Italic faces available to the
// current process without installing them into the user's system font library,
// and returns the family name to pass to SetFontFamily.
func Register() string {
	once.Do(func() {
		for _, face := range [][]byte{regularTTF, boldTTF, italicTTF} {
			if fam, _, err := app.RegisterFont(face); err == nil && fam != "" {
				registered = fam
			}
		}
		app.SetInterfaceFontFamily(registered)
	})
	return registered
}
