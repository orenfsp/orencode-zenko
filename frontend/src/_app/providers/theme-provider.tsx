"use client";

import * as React from "react";

type Theme = "light" | "dark";
type ThemeOrigin = { x: number; y: number };
type ViewTransitionDocument = Document & {
  startViewTransition?: (update: () => void) => unknown;
};

interface ThemeContextValue {
  theme: Theme;
  setTheme: (theme: Theme, origin?: ThemeOrigin) => void;
  toggleTheme: (origin?: ThemeOrigin) => void;
}

const ThemeContext = React.createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = React.useState<Theme>("light");

  React.useEffect(() => {
    const stored = window.localStorage.getItem("fsp-theme");
    const nextTheme: Theme = stored === "dark" || stored === "light" ? stored : "light";
    document.documentElement.classList.toggle("dark", nextTheme === "dark");
    document.documentElement.style.colorScheme = nextTheme;
    const frame = window.requestAnimationFrame(() => setThemeState(nextTheme));
    return () => window.cancelAnimationFrame(frame);
  }, []);

  const setTheme = React.useCallback((nextTheme: Theme, origin?: ThemeOrigin) => {
    const point = origin ?? { x: window.innerWidth / 2, y: window.innerHeight / 2 };
    const root = document.documentElement;
    root.style.setProperty("--theme-x", `${point.x}px`);
    root.style.setProperty("--theme-y", `${point.y}px`);
    const applyTheme = () => {
      setThemeState(nextTheme);
      window.localStorage.setItem("fsp-theme", nextTheme);
      root.classList.toggle("dark", nextTheme === "dark");
      root.style.colorScheme = nextTheme;
    };
    const viewTransitionDocument = document as ViewTransitionDocument;
    if (viewTransitionDocument.startViewTransition) {
      viewTransitionDocument.startViewTransition(applyTheme);
    } else {
      applyTheme();
    }
  }, []);

  const value = React.useMemo(
    () => ({ theme, setTheme, toggleTheme: (origin?: ThemeOrigin) => setTheme(theme === "dark" ? "light" : "dark", origin) }),
    [setTheme, theme],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const context = React.useContext(ThemeContext);
  if (!context) throw new Error("useTheme must be used within ThemeProvider");
  return context;
}
