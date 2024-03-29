package helm

import (
	"fmt"
	"io"
	"strings"

	"helm.sh/helm/v3/pkg/release"
)

// PrintRelease prints info about a release
func PrintRelease(out io.Writer, rel *release.Release) {
	if rel == nil {
		return
	}
	fmt.Fprintf(out, "NAME: %s\n", rel.Name)
	if !rel.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", rel.Info.LastDeployed)
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", rel.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", rel.Info.Status.String())

	executions := executionsByHookEvent(rel)
	if tests, ok := executions[release.HookTest]; ok {
		for _, h := range tests {
			// Don't print anything if hook has not been initiated
			if h.LastRun.StartedAt.IsZero() {
				continue
			}
			fmt.Fprintf(out, "TEST SUITE:     %s\n%s\n%s\n%s\n\n",
				h.Name,
				fmt.Sprintf("Last Started:   %s", h.LastRun.StartedAt),
				fmt.Sprintf("Last Completed: %s", h.LastRun.CompletedAt),
				fmt.Sprintf("Phase:          %s", h.LastRun.Phase),
			)
		}
	}

	if strings.EqualFold(rel.Info.Description, "Dry run complete") {
		fmt.Fprintf(out, "MANIFEST:\n%s\n", rel.Manifest)
	}

	if len(rel.Info.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", strings.TrimSpace(rel.Info.Notes))
	}
}

func executionsByHookEvent(rel *release.Release) map[release.HookEvent][]*release.Hook {
	result := make(map[release.HookEvent][]*release.Hook)
	for _, h := range rel.Hooks {
		for _, e := range h.Events {
			executions, ok := result[e]
			if !ok {
				executions = []*release.Hook{}
			}
			result[e] = append(executions, h)
		}
	}
	return result
}
