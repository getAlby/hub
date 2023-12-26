import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import About from "./screens/About";
import Apps from "./screens/apps/Apps";
import Show from "./screens/apps/Show";
import Login from "./screens/Login";
import NotFound from "./screens/NotFound";

import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import New from "./screens/apps/New";
import Created from "./screens/apps/Created";

function App() {
  return (
    <div className="p-4 dark:bg-black min-h-full">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Navbar />}>
            <Route index element={<Navigate to="/apps" replace />} />
            <Route path="apps" element={<Apps />} />
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
    </div>
  );
}

export default App;
