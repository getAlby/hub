import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import About from "./screens/About";
import AppsList from "./screens/apps/AppsList";
import ShowApp from "./screens/apps/ShowApp";
import Login from "./screens/Login";
import NotFound from "./screens/NotFound";

import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import NewApp from "./screens/apps/NewApp";
import AppCreated from "./screens/apps/AppCreated";

function App() {
  return (
    <div className="dark:bg-black min-h-full">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Navbar />}>
            <Route index element={<Navigate to="/apps" replace />} />
            <Route path="apps" element={<AppsList />} />
            <Route path="apps/:pubkey" element={<ShowApp />} />
            <Route path="apps/new" element={<NewApp />} />
            <Route path="apps/created" element={<AppCreated />} />
            <Route path="about" element={<About />} />
          </Route>
          <Route path="login" element={<Login />} />
          <Route path="/*" element={<NotFound />} />
        </Routes>
        <Footer />
      </BrowserRouter>
    </div>
  );
}

export default App;
