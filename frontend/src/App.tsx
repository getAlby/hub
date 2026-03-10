import React from "react";
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
import { isHttpMode } from "src/utils/isHttpMode";

const createRouterFunc = isHttpMode() ? createBrowserRouter : createHashRouter;
const basePath =
  import.meta.env.BASE_URL !== "/" ? import.meta.env.BASE_URL : "";
const router = createRouterFunc(routes, {
  // if running on a subpath, use the subpath as the router basename
  // BASE_URL is set via process.env.BASE_PATH, see https://vite.dev/guide/build#public-base-path
  basename: basePath || undefined,
});

function App() {
  const { data: info } = useInfo();

  React.useEffect(() => {
    if (!isHttpMode() || !("registerProtocolHandler" in navigator)) {
      return;
    }
    try {
      const handlerUrl = `${window.location.origin}${basePath}/wallet/send?bip21=%s`;
      navigator.registerProtocolHandler("bitcoin", handlerUrl);
    } catch (e) {
      console.error("Failed to register bitcoin protocol handler", e);
    }
  }, []);

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
