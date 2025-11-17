import { useEffect } from "react";
import {
  RouterProvider,
  createBrowserRouter,
  createHashRouter,
} from "react-router-dom";
import { Toaster } from "src/components/ui/sonner";
import { ThemeProvider } from "src/components/ui/theme-provider";
import { TouchProvider } from "src/components/ui/tooltip";
import { useInfo } from "src/hooks/useInfo";
import routes from "src/routes.tsx";
import usePrivacyStore from "src/state/PrivacyStore";
import { isHttpMode } from "src/utils/isHttpMode";

const createRouterFunc = isHttpMode() ? createBrowserRouter : createHashRouter;
const router = createRouterFunc(routes, {
  // if running on a subpath, use the subpath as the router basename
  // BASE_URL is set via process.env.BASE_PATH, see https://vite.dev/guide/build#public-base-path
  basename:
    import.meta.env.BASE_URL !== "/" ? import.meta.env.BASE_URL : undefined,
});

function App() {
  const { data: info } = useInfo();
  const { privacyMode } = usePrivacyStore();

  // Apply privacy mode class on mount and when it changes
  useEffect(() => {
    if (privacyMode) {
      document.body.classList.add("privacy-mode");
    } else {
      document.body.classList.remove("privacy-mode");
    }
  }, [privacyMode]);

  return (
    <>
      <TouchProvider>
        <ThemeProvider
          defaultTheme="default"
          defaultDarkMode="system"
          storageKey="vite-ui-theme"
        >
          {info && <RouterProvider router={router} />}
          <Toaster position="bottom-right" richColors={true} />
        </ThemeProvider>
      </TouchProvider>
    </>
  );
}

export default App;
