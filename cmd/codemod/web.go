package codemod

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/log"
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
			"tailwindcss",
			"@tailwindcss/vite",
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

/* 
 * This imports the design tokens from gvtech-design
 * In a real app you might want to import the CSS directly if provided by the package
 */
@layer theme {
  :root {
    /* Define base tokens here or import from @gv-tech/design-tokens */
  }
}
`
		if err := os.WriteFile(filepath.Join(targetDir, "src", "index.css"), []byte(indexCss), 0o644); err != nil {
			return fmt.Errorf("failed to write index.css: %w", err)
		}

		appTsx := `import { useState } from 'react'
import { Button } from '@gv-tech/ui-web/Button'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4 bg-background">
      <h1 className="text-4xl font-bold mb-8">gvtech-design Web</h1>
      <div className="flex flex-col items-center gap-4">
        <Button onClick={() => setCount((c) => c + 1)}>
          Count is {count}
        </Button>
      </div>
    </div>
  )
}

export default App
`
		if err := os.WriteFile(filepath.Join(targetDir, "src", "App.tsx"), []byte(appTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write App.tsx: %w", err)
		}

		log.Info("Web project %s bootstrapped successfully!", projectName)
		return nil
	},
}
