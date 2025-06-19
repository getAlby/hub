import {
  RouterProvider,
  createBrowserRouter,
  createHashRouter,
} from "react-router-dom";

import { ThemeProvider } from "src/components/ui/theme-provider";

import { Toaster } from "src/components/ui/toaster";
import { TouchProvider } from "src/components/ui/tooltip";
import { useInfo } from "src/hooks/useInfo";
import routes from "src/routes.tsx";
import { isHttpMode } from "src/utils/isHttpMode";

const createRouterFunc = isHttpMode() ? createBrowserRouter : createHashRouter;
const router = createRouterFunc(routes, {
  // if running on a subpath, use the subpath as the router basename
  basename:
    import.meta.env.BASE_URL !== "/" ? import.meta.env.BASE_URL : undefined,
});

function App() {
  const { data: info } = useInfo();

  return (
    <>
      <TouchProvider>
        <ThemeProvider
          defaultTheme="default"
          defaultDarkMode="system"
          storageKey="vite-ui-theme"
        >
          <Toaster />
          {info && <RouterProvider router={router} />}
        </ThemeProvider>
      </TouchProvider>
    </>
  );
}

export default App;
