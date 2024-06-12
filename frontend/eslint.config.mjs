import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import { FlatCompat } from "@eslint/eslintrc";
import js from "@eslint/js";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import reactRefresh from "eslint-plugin-react-refresh";
import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  {
    ignores: [
      "**/dist",
      "**/node_modules",
      "**/dist",
      "src/components/ui/navigation-menu.tsx",
    ],
  },
  ...fixupConfigRules(
    compat.extends(
      "eslint:recommended",
      "plugin:@typescript-eslint/recommended",
      "plugin:react-hooks/recommended",
      "prettier"
    )
  ),
  {
    plugins: {
      "react-refresh": reactRefresh,
      "@typescript-eslint": fixupPluginRules(typescriptEslint),
    },

    languageOptions: {
      globals: {
        ...globals.browser,
      },

      parser: tsParser,
    },
    files: ["**/*.ts", "**/*.tsx"],
    rules: {
      "react-refresh/only-export-components": [
        "warn",
        {
          allowConstantExport: true,
        },
      ],

      "@typescript-eslint/ban-ts-comment": [
        "error",
        {
          "ts-ignore": "allow-with-description",
        },
      ],

      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          args: "none",
        },
      ],

      "no-console": [
        "error",
        {
          allow: ["info", "warn", "error"],
        },
      ],

      "no-constant-binary-expression": "error",
      curly: "error",
    },
  },
];
