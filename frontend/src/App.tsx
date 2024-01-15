import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import Navbar from "src/components/Navbar";
import Footer from "src/components/Footer";
import Toaster from "src/components/Toast/Toaster";

import About from "src/screens/About";
import Login from "src/screens/Login";
import AppsList from "src/screens/apps/AppsList";
import ShowApp from "src/screens/apps/ShowApp";
import NewApp from "src/screens/apps/NewApp";
import AppCreated from "src/screens/apps/AppCreated";
import NotFound from "src/screens/NotFound";

function App() {
  return (
    <div className="bg:white dark:bg-black min-h-full">
      <Toaster />
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
