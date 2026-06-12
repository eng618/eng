# Technical Roadmap & Action Plan (UI/UX Focus)

This document serves as the authoritative roadmap for elevating the user experience (UX) and user interface (UI) of the `eng` CLI. With foundational architecture and basic TUIs in place, the next evolution focuses on premium aesthetics, cohesive theming, and advanced interactive workflows using the Charm ecosystem (Bubble Tea, Lipgloss, Huh).

**Aesthetic Goal**: Follow a minimalist aesthetic leveraging the `gv-tech` brand tokens and design system ([GV Tech Theming Reference](https://design.garciaericn.com/docs/theming#token-reference)).

## Status Tracking Mechanism

- `[ ]` Not Started
- `[/]` In Progress
- `[!]` Blocked
- `[x]` Complete

---

## 1. Phase 6: Foundational Theming & Styling

Create a consistent visual language across the entire application that prioritizes general usability and consistency.

### 1.1 Global Design System (Lipgloss)
- **Status**: `[ ]`
- **Task**: Introduce an `internal/ui/theme` package defining a global, branded color palette, typography (bolding/italics), and border styles using `charmbracelet/lipgloss`. Utilize the `gv-tech` brand tokens to ensure a minimalist, professional aesthetic.
- **Impact**: High (Ensures the CLI looks consistent and recognizable).

### 1.2 Rich Command Output
- **Status**: `[ ]`
- **Task**: Replace standard `fmt.Printf` and plain text logging with Lipgloss-styled components (e.g., success banners, info boxes, warning callouts) aligned with the global theme.
- **Impact**: Medium (Makes reading command output easier and more pleasant).

---

## 2. Phase 7: Error UX & Actionable Feedback

Transform how the CLI communicates failures to the user.

### 2.1 Smart, Styled Error Handling
- **Status**: `[ ]`
- **Task**: Create a custom error formatter. When an error occurs, output a styled error box that not only explains *what* failed, but provides *actionable suggestions* on how to fix it (e.g., pointing to the exact `eng system setup` command needed).
- **Impact**: High (Drastically reduces user frustration during setup or edge cases).

### 2.2 Progressive Disclosure for Logs
- **Status**: `[ ]`
- **Task**: For verbose commands (like git syncs or system updates), show a clean summary by default. Log the verbose output to a file, and provide a command to view the logs, preventing the terminal from being overwhelmed with text.
- **Impact**: Medium (Keeps the terminal clean while preserving debuggability).

---

## 3. Phase 8: Interactive Dashboards (Advanced TUIs)

Introduce full-screen, interactive applications for complex workflows.

### 3.1 Project & Git Dashboard
- **Status**: `[ ]`
- **Task**: Build an `eng dashboard` or interactive `eng project view` using `charmbracelet/bubbletea`. Show real-time statuses of repositories (branches, uncommitted changes, sync state) in a navigable, split-pane view.
- **Impact**: High (Provides a "mission control" feel for managing multiple repositories).

### 3.2 Interactive Configuration Editor
- **Status**: `[ ]`
- **Task**: Create `eng config edit --interactive`. Instead of manually editing `.eng.yaml`, provide a beautiful TUI form (via `huh`) to toggle settings, set paths, and validate input dynamically.
- **Impact**: Medium (Reduces misconfiguration errors).

---

## 4. Phase 9: Onboarding & Discoverability

Ensure new users (and existing users discovering new features) have a frictionless experience.

### 4.1 First-Run Onboarding Wizard
- **Status**: `[ ]`
- **Task**: Detect if `.eng.yaml` is missing or incomplete on first run. Instead of failing, automatically launch a welcoming, step-by-step TUI wizard to configure essential paths and preferences.
- **Impact**: High (Creates a magical first impression).

### 4.2 Overhaul Cobra Help & Usage
- **Status**: `[ ]`
- **Task**: Customize Cobra's help templates to use colored headers, aligned columns, and examples. Consider integrating `charmbracelet/glamour` to render markdown-based command documentation directly in the terminal.
- **Impact**: Medium (Improves discoverability of flags and subcommands).

### 4.3 Dynamic Auto-Completion
- **Status**: `[ ]`
- **Task**: Implement Cobra's `ValidArgsFunction` across commands to provide dynamic shell auto-completion for project names, dotfile templates, and git branches.
- **Impact**: High (Massively speeds up daily usage).
