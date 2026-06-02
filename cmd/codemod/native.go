package codemod

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/log"
)

var NativeCmd = &cobra.Command{
	Use:   "native [projectName]",
	Short: "Bootstrap a new React Native Expo project with gvtech-design",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		targetDir := filepath.Join(os.Getenv("PWD"), projectName)

		if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
			return fmt.Errorf("directory %s already exists", projectName)
		}

		fmt.Printf("🚀 Bootstrapping Native project: %s...\n", projectName)
		
		// Create Expo App
		createCmd := execCommand("bunx", "create-expo-app@latest", projectName)
		createCmd.Stdout = log.Writer()
		createCmd.Stderr = log.ErrorWriter()
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create expo app: %w", err)
		}	

		log.Info("Installing gvtech-design Native dependencies...")
		installCmd := execCommand("bun", "add", "nativewind@^4.2.2", "tailwindcss@^3.4.1", "react-native-reanimated@^4.3.1", "lucide-react-native", "@gv-tech/ui-native", "@gv-tech/ui-core", "@gv-tech/design-tokens", "clsx", "tailwind-merge")
		installCmd.Dir = targetDir
		installCmd.Stdout = log.Writer()
		installCmd.Stderr = log.ErrorWriter()
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		log.Info("Configuring Tailwind & NativeWind...")
		
		tailwindConfig := `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,jsx,ts,tsx}",
    "./components/**/*.{js,jsx,ts,tsx}",
    "./node_modules/@gv-tech/ui-native/src/**/*.{js,jsx,ts,tsx}"
  ],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {},
  },
  plugins: [],
};
`
		if err := os.WriteFile(filepath.Join(targetDir, "tailwind.config.js"), []byte(tailwindConfig), 0644); err != nil {
			return fmt.Errorf("failed to write tailwind.config.js: %w", err)
		}

		babelConfig := `module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ["babel-preset-expo", { jsxImportSource: "nativewind" }],
      "nativewind/babel",
    ],
    plugins: [
      "react-native-reanimated/plugin",
    ],
  };
};
`
		if err := os.WriteFile(filepath.Join(targetDir, "babel.config.js"), []byte(babelConfig), 0644); err != nil {
			return fmt.Errorf("failed to write babel.config.js: %w", err)
		}

		appIndex := `import { View } from "react-native";
import { Text } from "@gv-tech/ui-native/Text";
import { Button } from "@gv-tech/ui-native/Button";

export default function Index() {
  return (
    <View className="flex-1 justify-center items-center p-4">
      <Text className="text-2xl font-bold mb-4">gvtech-design Native</Text>
      <Button onPress={() => console.log("Pressed")}>
        <Text>Click Me</Text>
      </Button>
    </View>
  );
}
`
		appDir := filepath.Join(targetDir, "app")
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return fmt.Errorf("failed to create app dir: %w", err)
		}
		
		if err := os.WriteFile(filepath.Join(appDir, "index.tsx"), []byte(appIndex), 0644); err != nil {
			return fmt.Errorf("failed to write app/index.tsx: %w", err)
		}

		log.Info("Native project %s bootstrapped successfully!", projectName)
		return nil
	},
}
