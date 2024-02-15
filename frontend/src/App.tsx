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
import { SetupWallet } from "src/screens/setup/SetupWallet";
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
import NewBlocktankChannel from "src/screens/channels/NewBlocktankChannel";
import NewOnchainAddress from "src/screens/onchain/NewAddress";
import NewCustomChannel from "src/screens/channels/NewCustomChannel";
import RecommendedChannels from "src/screens/channels/RecommendedChannels";
import { SetupMnemonic } from "src/screens/setup/SetupMnemonic";
import { SetupFinish } from "src/screens/setup/SetupFinish";

function App() {
  return (
    <div className="bg-zinc-50 min-h-full flex flex-col justify-center dark:bg-zinc-950">
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
              <Route path="wallet" element={<SetupWallet />} />
              <Route path="mnemonic" element={<SetupMnemonic />} />
              <Route path="finish" element={<SetupFinish />} />
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
            <Route
              path="channels/new/blocktank"
              element={<NewBlocktankChannel />}
            />
            <Route
              path="channels/recommended"
              element={<RecommendedChannels />}
            />
            <Route path="channels/new/custom" element={<NewCustomChannel />} />
            <Route path="onchain/new-address" element={<NewOnchainAddress />} />
            <Route path="unlock" element={<Unlock />} />
            <Route path="about" element={<About />} />
          </Route>
          <Route path="/*" element={<NotFound />} />
        </Routes>
      </HashRouter>
      <Footer />
    </div>
  );
}

export default App;
