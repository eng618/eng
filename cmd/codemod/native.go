package codemod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		// Patch package.json to fix LightningCSS bug with NativeWind
		packageJsonPath := filepath.Join(targetDir, "package.json")
		packageJsonBytes, err := os.ReadFile(packageJsonPath)
		if err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(packageJsonBytes, &pkg) == nil {
				pkg["resolutions"] = map[string]interface{}{
					"lightningcss": "1.30.1",
				}
				pkg["overrides"] = map[string]interface{}{
					"lightningcss": "1.30.1",
				}
				if newBytes, err := json.MarshalIndent(pkg, "", "  "); err == nil {
					_ = os.WriteFile(packageJsonPath, newBytes, 0o644)
				}
			}
		}

		log.Info("Installing gvtech-design Native dependencies...")
		installCmd := execCommand(
			"bun",
			"add",
			"nativewind@preview",
			"tailwindcss@^4.3.0",
			"@tailwindcss/postcss",
			"postcss",
			"react-native-reanimated@4.3.1",
			"react-native-svg",
			"react-native-css-interop",
			"lucide-react-native@^1.8.0",
			"@gv-tech/ui-native",
			"@gv-tech/ui-core",
			"@gv-tech/design-tokens",
			"clsx",
			"tailwind-merge",
		)
		installCmd.Dir = targetDir
		installCmd.Stdout = log.Writer()
		installCmd.Stderr = log.ErrorWriter()
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		log.Info("Configuring Tailwind & NativeWind v5...")

		postcssConfig := `export default {
  plugins: {
    "@tailwindcss/postcss": {},
  },
};
`
		if err := os.WriteFile(
			filepath.Join(targetDir, "postcss.config.mjs"),
			[]byte(postcssConfig),
			0o644,
		); err != nil {
			return fmt.Errorf("failed to write postcss.config.mjs: %w", err)
		}

		babelConfig := `module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ["babel-preset-expo"],
    ],
    plugins: [
      "react-native-reanimated/plugin",
    ],
  };
};
`
		if err := os.WriteFile(filepath.Join(targetDir, "babel.config.js"), []byte(babelConfig), 0o644); err != nil {
			return fmt.Errorf("failed to write babel.config.js: %w", err)
		}

		metroConfig := `const { getDefaultConfig } = require("expo/metro-config");
const { withNativeWind } = require("nativewind/metro");

const config = getDefaultConfig(__dirname);

module.exports = withNativeWind(config, { input: "./global.css" });
`
		if err := os.WriteFile(filepath.Join(targetDir, "metro.config.js"), []byte(metroConfig), 0o644); err != nil {
			return fmt.Errorf("failed to write metro.config.js: %w", err)
		}

		// Clean up default Expo template files
		_ = os.RemoveAll(filepath.Join(targetDir, "app"))
		_ = os.RemoveAll(filepath.Join(targetDir, "components"))
		_ = os.RemoveAll(filepath.Join(targetDir, "constants"))
		_ = os.RemoveAll(filepath.Join(targetDir, "hooks"))
		_ = os.RemoveAll(filepath.Join(targetDir, "scripts"))
		_ = os.RemoveAll(filepath.Join(targetDir, "src"))

		appDir := filepath.Join(targetDir, "app")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			return fmt.Errorf("failed to create app dir: %w", err)
		}
		componentsDir := filepath.Join(targetDir, "components")
		if err := os.MkdirAll(componentsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create components dir: %w", err)
		}

		// Write global.css
		globalCss := `@import "tailwindcss/theme.css" layer(theme);
@import "tailwindcss/preflight.css" layer(base);
@import "tailwindcss/utilities.css";
@import "nativewind/theme";
@import "@gv-tech/design-tokens/theme.css";
@source "./node_modules/@gv-tech/ui-native";
`
		if err := os.WriteFile(filepath.Join(targetDir, "global.css"), []byte(globalCss), 0o644); err != nil {
			return fmt.Errorf("failed to write global.css: %w", err)
		}

		// Write components/Header.tsx
		headerTsx := `import { View } from "react-native";
import { Text, ThemeToggle } from "@gv-tech/ui-native";

export function Header() {
  return (
    <View className="flex-row items-center justify-between p-4 border-b border-border bg-background pt-12">
      <Text className="text-xl font-bold text-foreground">GV Tech</Text>
      <ThemeToggle />
    </View>
  );
}
`
		if err := os.WriteFile(filepath.Join(componentsDir, "Header.tsx"), []byte(headerTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write components/Header.tsx: %w", err)
		}

		// Write components/Footer.tsx
		footerTsx := `import { View } from "react-native";
import { Text } from "@gv-tech/ui-native";

export function Footer() {
  return (
    <View className="p-4 border-t border-border bg-background items-center pb-8">
      <Text className="text-sm text-muted-foreground">© 2026 GV Tech</Text>
    </View>
  );
}
`
		if err := os.WriteFile(filepath.Join(componentsDir, "Footer.tsx"), []byte(footerTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write components/Footer.tsx: %w", err)
		}

		// Write app/_layout.tsx
		layoutTsx := `import { Stack } from "expo-router";
import { ThemeProvider } from "@gv-tech/ui-native";
import { Header } from "../components/Header";
import { Footer } from "../components/Footer";
import { View, Platform, Appearance } from "react-native";
import { cssInterop, colorScheme } from "react-native-css-interop";
import { Sun, Moon, SunMoon, LayoutDashboard } from "lucide-react-native";

if (Platform.OS === "web") {
  // Polyfill Appearance on web
  const listeners = new Set<any>();
  // @ts-ignore
  Appearance.setColorScheme = (scheme) => {
    colorScheme.set(scheme);
    if (typeof document !== 'undefined') {
      if (scheme === 'dark') document.documentElement.classList.add('dark');
      else document.documentElement.classList.remove('dark');
    }
    for (const listener of listeners) {
      listener({ colorScheme: scheme });
    }
  };
  const originalGetColorScheme = Appearance.getColorScheme;
  Appearance.getColorScheme = () => {
    const current = colorScheme.get();
    return current !== "system" && current !== undefined ? current : (originalGetColorScheme ? originalGetColorScheme() : "light");
  };
  const originalAddChangeListener = Appearance.addChangeListener;
  Appearance.addChangeListener = (listener) => {
    listeners.add(listener);
    const result = originalAddChangeListener ? originalAddChangeListener(listener) : { remove: () => {} };
    return {
      remove: () => {
        listeners.delete(listener);
        if (result.remove) result.remove();
      }
    };
  };
  // @ts-ignore
  if (Appearance.default) {
    // @ts-ignore
    Appearance.default.setColorScheme = Appearance.setColorScheme;
    // @ts-ignore
    Appearance.default.getColorScheme = Appearance.getColorScheme;
    // @ts-ignore
    Appearance.default.addChangeListener = Appearance.addChangeListener;
  }
}

cssInterop(Sun, { className: "style" });
cssInterop(Moon, { className: "style" });
cssInterop(SunMoon, { className: "style" });
cssInterop(LayoutDashboard, { className: "style" });

import "../global.css";

export default function Layout() {
  return (
    <ThemeProvider>
      <View className="flex-1 bg-background">
        <Header />
        <Stack screenOptions={{ headerShown: false }} />
        <Footer />
      </View>
    </ThemeProvider>
  );
}
`
		if err := os.WriteFile(filepath.Join(appDir, "_layout.tsx"), []byte(layoutTsx), 0o644); err != nil {
			return fmt.Errorf("failed to write app/_layout.tsx: %w", err)
		}

		// Patch app.json to use "single" output for web to prevent NativeWind SSR crashes
		appJsonPath := filepath.Join(targetDir, "app.json")
		appJsonBytes, err := os.ReadFile(appJsonPath)
		if err == nil {
			patchedAppJson := strings.Replace(string(appJsonBytes), `"output": "static"`, `"output": "single"`, 1)
			if !strings.Contains(patchedAppJson, `"output": "single"`) {
			    // Fallback if static wasn't explicitly defined
				patchedAppJson = strings.Replace(string(appJsonBytes), `"web": {`, `"web": {
      "output": "single",`, 1)
			}
			_ = os.WriteFile(appJsonPath, []byte(patchedAppJson), 0o644)
		}

		// Write app/index.tsx
		appIndex := `import { View, Linking } from "react-native";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle, Button, Text } from "@gv-tech/ui-native";
import { LayoutDashboard } from "lucide-react-native";

export default function Index() {
  return (
    <View className="flex-1 p-4 bg-background justify-center">
      <Empty>
        <EmptyHeader>
          <EmptyMedia>
            <LayoutDashboard className="text-primary" size={32} />
          </EmptyMedia>
          <EmptyTitle>Welcome to Your New App</EmptyTitle>
          <EmptyDescription>
            This is a fully functional design system boilerplate. Start customizing by editing the <Text className="font-mono text-muted-foreground">app/index.tsx</Text> file.
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent className="mt-4 flex justify-center">
          <Button onPress={() => Linking.openURL("https://design.garciaericn.com/docs/getting-started")}>
            <Text>Read the Docs</Text>
          </Button>
        </EmptyContent>
      </Empty>
    </View>
  );
}
`
		if err := os.WriteFile(filepath.Join(appDir, "index.tsx"), []byte(appIndex), 0o644); err != nil {
			return fmt.Errorf("failed to write app/index.tsx: %w", err)
		}

		// Inject GV Tech icon
		if iconData, err := AssetsFS.ReadFile("assets/icon.png"); err == nil {
			assetsDir := filepath.Join(targetDir, "assets", "images")
			os.MkdirAll(assetsDir, 0o755)
			
			// Replace all standard Expo icons
			os.WriteFile(filepath.Join(assetsDir, "icon.png"), iconData, 0o644)
			os.WriteFile(filepath.Join(assetsDir, "favicon.png"), iconData, 0o644)
			os.WriteFile(filepath.Join(assetsDir, "adaptive-icon.png"), iconData, 0o644)
			os.WriteFile(filepath.Join(assetsDir, "splash-icon.png"), iconData, 0o644)
		}

		log.Info("Native project %s bootstrapped successfully!", projectName)
		return nil
	},
}
