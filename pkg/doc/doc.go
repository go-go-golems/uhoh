package doc

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}

// GetUhohDSLDocumentation returns the markdown content of the uhoh DSL guide.
func GetUhohDSLDocumentation() (string, error) {
	hs := help.NewHelpSystem()
	if err := AddDocToHelpSystem(hs); err != nil {
		return "", err
	}
	sec, err := hs.GetSectionWithSlug("uhoh-dsl")
	if err != nil {
		return "", err
	}
	return sec.Content, nil
}
