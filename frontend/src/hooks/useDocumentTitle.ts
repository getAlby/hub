import React from "react";
import { useMatches } from "react-router-dom";

/**
 * Custom hook to manage document.title based on route handles.
 * This ensures the browser's history entries include a proper title
 * (fixes: back gesture / long-press back showing empty/incorrect titles).
 *
 * Looks for a `title` property in route handles (string or function),
 * falling back to "Alby Hub" if not present.
 */
export function useDocumentTitle() {
  const matches = useMatches();

  React.useEffect(() => {
    try {
      // Extract title from route handles. Use a typed helper to get the title
      // from the handle object (if present).
      const getTitleFromHandle = (handle: unknown): string | null => {
        if (handle && typeof handle === "object") {
          const h = handle as { title?: unknown };
          if (typeof h.title === "string") {
            return h.title;
          }
          if (typeof h.title === "function") {
            try {
              return (h.title as () => string)();
            } catch (err) {
              return null;
            }
          }
        }
        return null;
      };

      // Find the last (most specific) route with a title, or default to "Alby Hub"
      const routeTitle =
        matches
          .map((m) => getTitleFromHandle(m.handle))
          .filter(Boolean)
          .pop() || "Alby Hub";

      // Set document title
      document.title = routeTitle;
    } catch (err) {
      console.error("Failed to compute page title", err);
    }
  }, [matches]);
}
