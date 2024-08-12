import { RouterProvider, createHashRouter } from "react-router-dom";

import { ThemeProvider } from "src/components/ui/theme-provider";

import { Toaster } from "src/components/ui/toaster";
import { useInfo } from "src/hooks/useInfo";
import routes from "src/routes.tsx";

const router = createHashRouter(routes);

function App() {
  const { data: info } = useInfo();

  return (
    <>
      <ThemeProvider
        defaultTheme="default"
        defaultDarkMode="system"
        storageKey="vite-ui-theme"
      >
        <Toaster />
        {info && <RouterProvider router={router} />}
      </ThemeProvider>
    </>
  );
}

export default App;
