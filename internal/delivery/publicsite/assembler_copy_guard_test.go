package publicsite

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed assembler.go
var assemblerGoSource string

func TestAssemblerGoDoesNotEmbedLegacyEnglishPublicLabels(t *testing.T) {
	banned := []string{
		`"Not Found"`,
		`"Go home"`,
		`"Latest published`,
		`"Published:"`,
	}
	for _, frag := range banned {
		if strings.Contains(assemblerGoSource, frag) {
			t.Fatalf("assembler.go contains banned legacy user-visible fragment %q — use publicfixtures instead", frag)
		}
	}
}
