import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import { UserProvider } from "./context/UserContext";
import RequireAuth from "./context/RequireAuth";
import Navbar from './components/navbar';
import Footer from './components/footer';
import NotFound from './screens/NotFound';
import About from "./screens/About";
import Connections from "./screens/apps/Index";

import './App.css'
import Show from "./screens/apps/Show";

function App() {
  return (
    <UserProvider>
      <BrowserRouter>
        <Routes>
          <Route
            path="/"
            element={
              <RequireAuth>
                <Navbar/>
              </RequireAuth>
            }
          >
            <Route index element={<Navigate to="/apps" replace />} />
            <Route path="apps" element={<Connections />} />
            <Route path="apps/:pubkey" element={<Show />} />
            <Route path="about" element={<About />} />
          </Route>
          <Route path="login" element={<h2>Login</h2>} />
          <Route path="/*" element={<NotFound />} />
        </Routes>
        <Footer />
      </BrowserRouter>
    </UserProvider>
  )
}

export default App
