import { RouterProvider, createHashRouter } from "react-router-dom";

import { ThemeProvider } from "src/components/ui/theme-provider";
import { usePosthog } from "./hooks/usePosthog";

import { Toaster } from "src/components/ui/toaster";
import routes from "src/routes.tsx";

function App() {
  usePosthog();

  const router = createHashRouter(routes);

  return (
    <>
      <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
        <Toaster />
        <RouterProvider router={router} />
      </ThemeProvider>
    </>
  );
}

export default App;
