import { HashRouter, Routes, Route, Navigate } from "react-router-dom";

import Navbar from "src/components/Navbar";
import Footer from "src/components/Footer";
import Toaster from "src/components/Toast/Toaster";

import About from "src/screens/About";
import AppsList from "src/screens/apps/AppsList";
import ShowApp from "src/screens/apps/ShowApp";
import NewApp from "src/screens/apps/NewApp";
import AppCreated from "src/screens/apps/AppCreated";
import NotFound from "src/screens/NotFound";
import { SetupNode } from "src/screens/setup/SetupNode";
import { Welcome } from "src/screens/Welcome";
import { SetupPassword } from "src/screens/setup/SetupPassword";
import Start from "src/screens/Start";
import { AppsRedirect } from "src/components/redirects/AppsRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import Unlock from "src/screens/Unlock";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import Channels from "src/screens/channels/Channels";
import NewChannel from "src/screens/channels/NewChannel";
import Blocktank from "src/screens/channels/Blocktank";
import NewOnchainAddress from "src/screens/onchain/NewAddress";

function App() {
  return (
    <div className="bg:white min-h-full dark:bg-black">
      <Toaster />
      <HashRouter>
        <Routes>
          <Route path="/" element={<Navbar />}>
            <Route path="" element={<HomeRedirect />} />
            <Route
              path="start"
              element={
                <StartRedirect>
                  <Start />
                </StartRedirect>
              }
            ></Route>
            <Route path="welcome" element={<Welcome />}></Route>
            <Route path="setup" element={<SetupRedirect />}>
              <Route path="" element={<Navigate to="password" replace />} />
              <Route path="password" element={<SetupPassword />} />
              <Route path="node" element={<SetupNode />} />
            </Route>
            <Route path="apps" element={<AppsRedirect />}>
              <Route index path="" element={<AppsList />} />
              <Route path=":pubkey" element={<ShowApp />} />
              <Route path="new" element={<NewApp />} />
              <Route path="created" element={<AppCreated />} />
              <Route path="*" element={<NotFound />} />
            </Route>
            <Route path="channels" element={<Channels />} />
            <Route path="channels/new" element={<NewChannel />} />
            <Route path="channels/new/blocktank" element={<Blocktank />} />
            <Route path="onchain/new-address" element={<NewOnchainAddress />} />
            <Route path="unlock" element={<Unlock />} />
            <Route path="about" element={<About />} />
          </Route>
          <Route path="/*" element={<NotFound />} />
        </Routes>
        <Footer />
      </HashRouter>
    </div>
  );
}

export default App;
