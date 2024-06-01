module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  extends: [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    "prettier",
  ],
  ignorePatterns: ["dist", ".eslintrc.cjs"],
  parser: "@typescript-eslint/parser",
  plugins: ["react-refresh", "@typescript-eslint"],
  rules: {
    "react-refresh/only-export-components": [
      "warn",
      { allowConstantExport: true },
    ],
    "@typescript-eslint/ban-ts-comment": [
      "error",
      {
        "ts-ignore": "allow-with-description",
      },
    ],
    "@typescript-eslint/no-unused-vars": ["warn", { args: "none" }],
    "no-console": ["error", { allow: ["info", "warn", "error"] }],
    "no-constant-binary-expression": "error",
    curly: "error",
  },
};
