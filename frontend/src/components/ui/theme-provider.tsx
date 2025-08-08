import { createContext, useContext, useEffect, useState } from "react";

export type DarkMode = "system" | "light" | "dark";
export const Themes = [
  "default",
  "alby",
  "bitcoin",
  "nostr",
  "matrix",
] as const;
export type Theme = (typeof Themes)[number];

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  defaultDarkMode?: DarkMode;
  storageKey?: string;
};

type ThemeProviderState = {
  theme: string;
  darkMode: string;
  setTheme: (theme: Theme) => void;
  setDarkMode: (mode: DarkMode) => void;
  isDarkMode: boolean;
};

const initialState: ThemeProviderState = {
  theme: "default",
  setTheme: () => null,
  darkMode: "system",
  setDarkMode: () => null,
  isDarkMode: false,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  defaultTheme = "default",
  defaultDarkMode = "system",
  storageKey = "vite-ui-theme",
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(() => {
    const themeFromStorage = localStorage.getItem(storageKey) as Theme;
    return Themes.includes(themeFromStorage) ? themeFromStorage : defaultTheme;
  });

  const [darkMode, setDarkMode] = useState<DarkMode>(() => {
    return (
      (localStorage.getItem(storageKey + "-darkmode") as DarkMode) ||
      defaultDarkMode
    );
  });

  const [isDarkMode, setIsDarkMode] = useState<boolean>(false);

  useEffect(() => {
    const root = window.document.documentElement;

    // Find and remove classes that start with 'theme-'
    const classList = root.classList;
    classList.forEach((className) => {
      if (className.startsWith("theme-")) {
        classList.remove(className);
      }
    });

    classList.add(`theme-${theme}`);

    let prefersDark = false;
    if (darkMode == "system") {
      prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
    } else {
      prefersDark = darkMode === "dark";
    }

    setIsDarkMode(prefersDark);

    if (prefersDark) {
      classList.add("dark");
    } else {
      classList.remove("dark");
    }
  }, [theme, darkMode]);

  const value = {
    theme,
    setTheme: (theme: Theme) => {
      localStorage.setItem(storageKey, theme);
      setTheme(theme);
    },
    darkMode,
    setDarkMode: (darkMode: DarkMode) => {
      localStorage.setItem(storageKey + "-darkmode", darkMode);
      setDarkMode(darkMode);
    },
    isDarkMode,
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }

  return context;
};
