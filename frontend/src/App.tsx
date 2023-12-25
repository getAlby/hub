import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import { UserProvider } from "./context/UserContext";
import RequireAuth from "./context/RequireAuth";

import About from "./screens/About";
import Connections from "./screens/apps/Index";
import Show from "./screens/apps/Show";
import Login from "./screens/Login";
import NotFound from "./screens/NotFound";

import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import New from "./screens/apps/New";
import Created from "./screens/apps/Created";

function App() {
  return (
    <UserProvider>
      <BrowserRouter>
        <Routes>
          <Route
            path="/"
            element={
              <RequireAuth>
                <Navbar />
              </RequireAuth>
            }
          >
            <Route index element={<Navigate to="/apps" replace />} />
            <Route path="apps" element={<Connections />} />
            <Route path="apps/:pubkey" element={<Show />} />
            <Route path="apps/new" element={<New />} />
            <Route path="apps/created" element={<Created />} />
            <Route path="about" element={<About />} />
          </Route>
          <Route path="login" element={<Login />} />
          <Route path="/*" element={<NotFound />} />
        </Routes>
        <Footer />
      </BrowserRouter>
    </UserProvider>
  );
}

export default App;
