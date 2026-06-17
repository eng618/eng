package codemod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
)

var WebCmd = &cobra.Command{
	Use:   "web [projectName]",
	Short: "Bootstrap a new React web project with gvtech-design",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		targetDir := filepath.Join(os.Getenv("PWD"), projectName)

		if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
			return fmt.Errorf("directory %s already exists", projectName)
		}

		log.Info("Bootstrapping Web project: %s...", projectName)

		// Create Vite App
		createCmd := execCommand("bun", "create", "vite", projectName, "--template", "react-ts")
		createCmd.Stdout = log.Writer()
		createCmd.Stderr = log.ErrorWriter()
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create vite app: %w", err)
		}

		log.Info("Installing gvtech-design Web dependencies...")
		installCmd := execCommand(
			"bun",
			"add",
			"tailwindcss@^4.3.0",
			"@tailwindcss/vite@^4.3.0",
			"@gv-tech/ui-web",
			"@gv-tech/ui-core",
			"@gv-tech/design-tokens",
			"clsx",
			"tailwind-merge",
			"lucide-react",
		)
		installCmd.Dir = targetDir
		installCmd.Stdout = log.Writer()
		installCmd.Stderr = log.ErrorWriter()
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		log.Info("Configuring Tailwind...")

		viteConfig := `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    tailwindcss(),
    react()
  ],
})
`
		if err := os.WriteFile(filepath.Join(targetDir, "vite.config.ts"), []byte(viteConfig), 0o644); err != nil {
			return fmt.Errorf("failed to write vite.config.ts: %w", err)
		}

		indexCss := `@import "tailwindcss";
@import "@gv-tech/design-tokens/theme.css";
@source "../node_modules/@gv-tech/ui-web";
`
		if err := os.WriteFile(filepath.Join(targetDir, "src", "index.css"), []byte(indexCss), 0o644); err != nil {
			return fmt.Errorf("failed to write index.css: %w", err)
		}

		componentsDir := filepath.Join(targetDir, "src", "components")
		if err := os.MkdirAll(componentsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create components dir: %w", err)
		}

		// Write src/components/Header.tsx
		headerTsx := `import { ThemeToggle } from "@gv-tech/ui-web/theme-toggle";
import { Text } from "@gv-tech/ui-web/text";

export function Header() {
  return (
    <header className="flex items-center justify-between p-4 border-b border-border bg-background">
      <Text className="text-xl font-bold text-foreground">GV Tech</Text>
      <ThemeToggle />
    </header>
  );
}
`
		if err := os.WriteFile(filepath.Join(componentsDir, "Header.tsx"), []byte(headerTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write src/components/Header.tsx: %w", err)
		}

		// Write src/components/Footer.tsx
		footerTsx := `import { Text } from "@gv-tech/ui-web/text";

export function Footer() {
  return (
    <footer className="p-4 border-t border-border bg-background flex items-center justify-center">
      <Text className="text-sm text-muted-foreground">© 2026 GV Tech</Text>
    </footer>
  );
}
`
		if err := os.WriteFile(filepath.Join(componentsDir, "Footer.tsx"), []byte(footerTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write src/components/Footer.tsx: %w", err)
		}

		appTsx := `import { ThemeProvider } from "@gv-tech/ui-web/theme-provider";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@gv-tech/ui-web/empty";
import { Button } from "@gv-tech/ui-web/button";
import { Text } from "@gv-tech/ui-web/text";
import { Header } from "./components/Header";
import { Footer } from "./components/Footer";
import { LayoutDashboard } from "lucide-react";

function App() {
  return (
    <ThemeProvider>
      <div className="flex flex-col min-h-screen bg-background">
        <Header />
        <main className="flex-1 flex flex-col items-center justify-center p-4">
          <Empty>
            <EmptyHeader>
              <EmptyMedia>
                <LayoutDashboard className="text-primary" size={32} />
              </EmptyMedia>
              <EmptyTitle>Welcome to Your New Web App</EmptyTitle>
              <EmptyDescription>
                This is a fully functional design system boilerplate. Start customizing by editing the <Text className="font-mono text-muted-foreground">src/App.tsx</Text> file.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent className="mt-4 flex justify-center">
              <Button onClick={() => window.open('https://design.garciaericn.com/docs/getting-started', '_blank')}>
                Read the Docs
              </Button>
            </EmptyContent>
          </Empty>
        </main>
        <Footer />
      </div>
    </ThemeProvider>
  )
}

export default App
`
		if err := os.WriteFile(filepath.Join(targetDir, "src", "App.tsx"), []byte(appTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write App.tsx: %w", err)
		}

		// Replace default icon with GV Tech icon
		if iconData, err := AssetsFS.ReadFile("assets/icon.png"); err == nil {
			publicDir := filepath.Join(targetDir, "public")
			if err := os.MkdirAll(publicDir, 0o755); err != nil {
				return fmt.Errorf("failed to create public dir: %w", err)
			}
			if err := os.WriteFile(filepath.Join(publicDir, "favicon.png"), iconData, 0o644); err != nil {
				return fmt.Errorf("failed to write favicon.png: %w", err)
			}

			// Update index.html to use the new icon
			indexPath := filepath.Join(targetDir, "index.html")
			if htmlData, err := os.ReadFile(indexPath); err == nil {
				updatedHtml := strings.ReplaceAll(string(htmlData), "/vite.svg", "/favicon.png")
				updatedHtml = strings.ReplaceAll(updatedHtml, `type="image/svg+xml"`, `type="image/png"`)
				if err := os.WriteFile(indexPath, []byte(updatedHtml), 0o644); err != nil {
					return fmt.Errorf("failed to update index.html: %w", err)
				}
			}
		}

		log.Info("Web project %s bootstrapped successfully!", projectName)
		return nil
	},
}
