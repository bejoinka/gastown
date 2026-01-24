package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var tapGuardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Block forbidden operations (PreToolUse hook)",
	Long: `Block forbidden operations via Claude Code PreToolUse hooks.

Guard commands exit with code 2 to BLOCK tool execution when a policy
is violated. They're called before the tool runs, preventing the
forbidden operation entirely.

Available guards:
  pr-workflow   - Block PR creation and feature branches (all agents)
  polecat       - Block git push and PR workflow (polecats only)

Example hook configuration:
  {
    "PreToolUse": [{
      "matcher": "Bash(gh pr create*)",
      "hooks": [{"command": "gt tap guard pr-workflow"}]
    }]
  }`,
}

var tapGuardPRWorkflowCmd = &cobra.Command{
	Use:   "pr-workflow",
	Short: "Block PR creation and feature branches",
	Long: `Block PR workflow operations in Gas Town.

Gas Town workers push directly to main. PRs add friction that breaks
the autonomous execution model (GUPP principle).

This guard blocks:
  - gh pr create
  - git checkout -b (feature branches)
  - git switch -c (feature branches)

Exit codes:
  0 - Operation allowed (not in Gas Town agent context)
  2 - Operation BLOCKED (in agent context)

The guard only blocks when running as a Gas Town agent (crew, polecat,
witness, etc.). Humans running outside Gas Town can still use PRs.`,
	RunE: runTapGuardPRWorkflow,
}

var tapGuardPolecatCmd = &cobra.Command{
	Use:   "polecat",
	Short: "Block polecat-forbidden git operations",
	Long: `Block git operations that polecats should not use directly.

Polecats must use 'gt done' to submit work to the merge queue.
Direct git push or PR workflows break the Refinery merge model.

This guard blocks:
  - git push (any push operation)
  - gh pr create
  - git checkout -b (feature branches)
  - git switch -c (feature branches)

Exit codes:
  0 - Operation allowed (not a polecat)
  2 - Operation BLOCKED (polecat context)

The guard only blocks when running as a polecat. Other roles and
humans are unaffected.`,
	RunE: runTapGuardPolecat,
}

func init() {
	tapCmd.AddCommand(tapGuardCmd)
	tapGuardCmd.AddCommand(tapGuardPRWorkflowCmd)
	tapGuardCmd.AddCommand(tapGuardPolecatCmd)
}

func runTapGuardPRWorkflow(cmd *cobra.Command, args []string) error {
	// Check if we're in a Gas Town agent context
	if !isGasTownAgentContext() {
		// Not in a Gas Town managed context - allow the operation
		return nil
	}

	// We're in a Gas Town context - block PR operations
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔══════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  ❌ PR WORKFLOW BLOCKED                                          ║")
	fmt.Fprintln(os.Stderr, "╠══════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(os.Stderr, "║  Gas Town workers push directly to main. PRs are forbidden.     ║")
	fmt.Fprintln(os.Stderr, "║                                                                  ║")
	fmt.Fprintln(os.Stderr, "║  Instead of:  gh pr create / git checkout -b / git switch -c    ║")
	fmt.Fprintln(os.Stderr, "║  Do this:     git add . && git commit && git push origin main   ║")
	fmt.Fprintln(os.Stderr, "║                                                                  ║")
	fmt.Fprintln(os.Stderr, "║  Why? PRs add friction that breaks autonomous execution.        ║")
	fmt.Fprintln(os.Stderr, "║  See: ~/gt/docs/PRIMING.md (GUPP principle)                     ║")
	fmt.Fprintln(os.Stderr, "╚══════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(2) // Exit 2 = BLOCK in Claude Code hooks

	return nil
}

// isGasTownAgentContext returns true if we're running as a Gas Town managed agent.
func isGasTownAgentContext() bool {
	// Check environment variables set by Gas Town session management
	envVars := []string{
		"GT_POLECAT",
		"GT_CREW",
		"GT_WITNESS",
		"GT_REFINERY",
		"GT_MAYOR",
		"GT_DEACON",
	}
	for _, env := range envVars {
		if os.Getenv(env) != "" {
			return true
		}
	}

	// Also check if we're in a crew or polecat worktree by path
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	agentPaths := []string{"/crew/", "/polecats/"}
	for _, path := range agentPaths {
		if strings.Contains(cwd, path) {
			return true
		}
	}

	return false
}

// isPolecatContext returns true if we're running as a polecat.
func isPolecatContext() bool {
	// Check environment variable
	if os.Getenv("GT_POLECAT") != "" {
		return true
	}

	// Check if we're in a polecat worktree by path
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	return strings.Contains(cwd, "/polecats/")
}

func runTapGuardPolecat(cmd *cobra.Command, args []string) error {
	// Check if we're in a polecat context
	if !isPolecatContext() {
		// Not a polecat - allow the operation
		return nil
	}

	// We're a polecat - block the operation
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔══════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  ❌ POLECAT RESTRICTION                                          ║")
	fmt.Fprintln(os.Stderr, "╠══════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(os.Stderr, "║  Polecats use gt done to submit work. Direct git push forbidden. ║")
	fmt.Fprintln(os.Stderr, "║                                                                  ║")
	fmt.Fprintln(os.Stderr, "║  Instead of:  git push / gh pr create / feature branches         ║")
	fmt.Fprintln(os.Stderr, "║  Do this:     gt done                                            ║")
	fmt.Fprintln(os.Stderr, "╚══════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(2) // Exit 2 = BLOCK in Claude Code hooks

	return nil
}
