package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/go-go-golems/uhoh/pkg/wizard"
	"github.com/go-go-golems/uhoh/pkg/wizard/shellcmd"
)

func main() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("unable to determine caller info")
	}

	wizardPath := filepath.Join(filepath.Dir(thisFile), "..", "wizard", "shell_viewer.yaml")
	w, err := wizard.LoadWizard(
		wizardPath,
		wizard.WithActionCallback("runShell", runShellAction),
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := w.Run(context.Background(), map[string]interface{}{}); err != nil {
		log.Fatal(err)
	}
}

func runShellAction(ctx context.Context, _ map[string]interface{}, args map[string]interface{}) (interface{}, error) {
	cmd, _ := args["cmd"].(string)
	if cmd == "" {
		return nil, fmt.Errorf("missing cmd argument")
	}

	return shellcmd.RunActionCallback(ctx, cmd, shellcmd.Options{
		Title:    "Shell Viewer Demo",
		KeepOpen: true,
	})
}
