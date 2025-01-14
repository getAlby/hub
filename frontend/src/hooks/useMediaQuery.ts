import * as React from "react";

export function useMediaQuery(query: string) {
  const [value, setValue] = React.useState(window.matchMedia(query).matches);

  React.useEffect(() => {
    function onChange(event: MediaQueryListEvent) {
      setValue(event.matches);
    }

    const result = matchMedia(query);
    result.addEventListener("change", onChange);
    setValue(result.matches);

    return () => result.removeEventListener("change", onChange);
  }, [query]);

  return value;
}

export function useIsDesktop() {
  const desktop = "(min-width: 1366px)";
  return useMediaQuery(desktop);
}
