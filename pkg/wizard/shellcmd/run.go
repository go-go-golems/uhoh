package shellcmd

import (
	"context"

	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
)

// RunActionCallback executes the shell viewer and returns an ActionCallbackResult suitable for
// wiring into wizard action steps. The returned result marks UIHandled so the step can skip its
// default completion notes. The Data payload is the underlying *Result.
func RunActionCallback(ctx context.Context, cmdStr string, opts Options) (interface{}, error) {
	res, err := Run(ctx, cmdStr, opts)
	if res == nil {
		return res, err
	}

	callbackResult := steps.ActionCallbackResult{
		Data:      res,
		UIHandled: true,
	}

	if err != nil {
		return callbackResult, err
	}

	if res.Err != nil {
		return callbackResult, res.Err
	}

	return callbackResult, nil
}
