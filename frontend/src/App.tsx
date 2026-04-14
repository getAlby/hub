import { createBrowserRouter, createHashRouter } from "react-router";
import { RouterProvider } from "react-router/dom";
import { Toaster } from "src/components/ui/sonner";
import { ThemeProvider } from "src/components/ui/theme-provider";
import { TouchProvider } from "src/components/ui/tooltip";
import { useInfo } from "src/hooks/useInfo";
import { useRegisterProtocolHandler } from "src/hooks/useRegisterProtocolHandler";
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

  useRegisterProtocolHandler(basePath);

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
