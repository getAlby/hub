export default {
  "src/**/*.{ts,tsx,json}": ["eslint --fix --max-warnings 0","prettier --write"],
  "platform_specific/**/*.ts": ["prettier --write"],
  "package.json": ["prettier --write"],
  "src/**/*.ts": () => "tsc --noEmit",
};
