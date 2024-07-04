import { RouterProvider, createHashRouter } from "react-router-dom";

import { ThemeProvider } from "src/components/ui/theme-provider";
import { usePosthog } from "./hooks/usePosthog";

import { Toaster } from "src/components/ui/toaster";
import routes from "src/routes.tsx";

const router = createHashRouter(routes);

function App() {
  usePosthog();

  return (
    <>
      <ThemeProvider
        defaultTheme="default"
        defaultDarkMode="system"
        storageKey="vite-ui-theme"
      >
        <Toaster />
        <RouterProvider router={router} />
      </ThemeProvider>
    </>
  );
}

export default App;
