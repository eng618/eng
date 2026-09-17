package config

import (
	"fmt"
	"strings"
)

// WizardStep is one resumable section of the init walkthrough.
type WizardStep string

const (
	WizardProfile   WizardStep = "profile"
	WizardGit       WizardStep = "git"
	WizardDotfiles  WizardStep = "dotfiles"
	WizardProxy     WizardStep = "proxy"
	WizardTelemetry WizardStep = "telemetry"
)

// wizardOrder is the fixed walkthrough order.
var wizardOrder = []WizardStep{WizardProfile, WizardGit, WizardDotfiles, WizardProxy, WizardTelemetry}

// WizardAnswers holds one walkthrough run. Defaults are preloaded from the
// current resolved config; empty means "leave unset".
type WizardAnswers struct {
	Email            string
	GitDevPath       string
	GitEditor        string
	Verbose          bool
	DotfilesRepoURL  string
	DotfilesBranch   string
	SkipDotfiles     bool
	ProxyAddress     string
	ProxyNoProxy     string
	SkipProxy        bool
	TelemetryEnabled bool
}

// ParseWizardStep parses --from into a WizardStep.
func ParseWizardStep(s string) (WizardStep, error) {
	if s == "" {
		return WizardProfile, nil
	}
	for _, st := range wizardOrder {
		if string(st) == strings.ToLower(s) {
			return st, nil
		}
	}
	return "", fmt.Errorf("unknown step %q (choose profile, git, dotfiles, proxy, telemetry)", s)
}

// RunWizard walks the prompt hooks from the given step and returns the
// planned key/value changes. It never writes; callers validate the result
// and persist once via ApplyWizardChanges.
func RunWizard(defaults WizardAnswers, from WizardStep) (map[string]any, error) {
	ans := defaults
	changes := map[string]any{}
	started := false
	for _, step := range wizardOrder {
		if step == from {
			started = true
		}
		if !started {
			continue
		}
		if err := runWizardStep(step, &ans, changes); err != nil {
			return nil, err
		}
	}
	return changes, nil
}

func runWizardStep(step WizardStep, ans *WizardAnswers, changes map[string]any) error {
	var err error
	switch step {
	case WizardProfile:
		ans.Email, err = InputPrompt("Email address (git + notifications)", ans.Email)
		if err != nil {
			return err
		}
		changes["email"] = ans.Email
	case WizardGit:
		ans.GitDevPath, err = InputPrompt("Git dev path (where repos live)", ans.GitDevPath)
		if err != nil {
			return err
		}
		changes["git.dev_path"] = ans.GitDevPath
		ans.GitEditor, err = InputPrompt("Default editor (blank = auto-detect)", ans.GitEditor)
		if err != nil {
			return err
		}
		changes["git.editor"] = ans.GitEditor
		ans.Verbose, err = ConfirmPrompt("Enable verbose output by default?", ans.Verbose)
		if err != nil {
			return err
		}
		changes["verbose"] = ans.Verbose
	case WizardDotfiles:
		ans.SkipDotfiles, err = ConfirmPrompt("Manage dotfiles with a bare repo?", !ans.SkipDotfiles)
		if err != nil {
			return err
		}
		ans.SkipDotfiles = !ans.SkipDotfiles
		if ans.SkipDotfiles {
			break
		}
		ans.DotfilesRepoURL, err = InputPrompt("Dotfiles repo URL (blank = skip)", ans.DotfilesRepoURL)
		if err != nil {
			return err
		}
		changes["dotfiles.repo_url"] = ans.DotfilesRepoURL
		ans.DotfilesBranch, err = InputPrompt("Dotfiles branch", ans.DotfilesBranch)
		if err != nil {
			return err
		}
		changes["dotfiles.branch"] = ans.DotfilesBranch
	case WizardProxy:
		var useProxy bool
		useProxy, err = ConfirmPrompt("Route traffic through a proxy?", !ans.SkipProxy)
		if err != nil {
			return err
		}
		ans.SkipProxy = !useProxy
		if ans.SkipProxy {
			break
		}
		ans.ProxyAddress, err = InputPrompt("Proxy address (host:port, blank = skip)", ans.ProxyAddress)
		if err != nil {
			return err
		}
		if ans.ProxyAddress == "" {
			ans.SkipProxy = true
			break
		}
		if err := ValidateProxyURLString(ans.ProxyAddress); err != nil {
			return err
		}
		changes["proxies"] = []any{map[string]any{
			"title":   "Default",
			"value":   NormalizeProxyURLString(ans.ProxyAddress),
			"enabled": true,
		}}
		ans.ProxyNoProxy, err = InputPrompt("No-proxy hosts (blank = defaults)", ans.ProxyNoProxy)
		if err != nil {
			return err
		}
		if ans.ProxyNoProxy != "" {
			entry := changes["proxies"].([]any)[0].(map[string]any)
			entry["noProxy"] = ans.ProxyNoProxy
			changes["proxies"] = []any{entry}
		}
	case WizardTelemetry:
		ans.TelemetryEnabled, err = ConfirmPrompt("Enable anonymous telemetry?", ans.TelemetryEnabled)
		if err != nil {
			return err
		}
		changes["telemetry.enabled"] = ans.TelemetryEnabled
	}
	return nil
}

// DefaultsFromResolved preloads wizard defaults from the current config.
func DefaultsFromResolved(rc *ResolvedConfig) WizardAnswers {
	if rc == nil {
		return WizardAnswers{GitDevPath: "~/Development", DotfilesBranch: "main"}
	}
	branch := rc.DotfilesBranch
	if branch == "" {
		branch = "main"
	}
	return WizardAnswers{
		Email:            rc.Email,
		GitDevPath:       rc.GitDevPath,
		GitEditor:        rc.GitEditor,
		Verbose:          rc.Verbose,
		DotfilesRepoURL:  rc.DotfilesRepo,
		DotfilesBranch:   branch,
		SkipProxy:        len(rc.Proxies) == 0,
		TelemetryEnabled: IsTelemetryEnabled(),
	}
}
