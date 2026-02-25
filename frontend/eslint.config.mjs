import eslint from "@eslint/js";
import tsParser from "@typescript-eslint/parser";
import eslintConfigPrettier from "eslint-config-prettier/flat";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import globals from "globals";
import tseslint from "typescript-eslint";

export default [
  {
    ignores: [
      "**/dist",
      "**/node_modules",
      "**/dist",
      "src/components/ui/navigation-menu.tsx",
    ],
  },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  reactRefresh.configs.vite,
  reactHooks.configs.flat["recommended-latest"],
  eslintConfigPrettier,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },

      parser: tsParser,
    },
    files: ["**/*.ts", "**/*.tsx"],
    rules: {
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
