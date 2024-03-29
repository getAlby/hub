import { HashRouter, Navigate, Route, Routes } from "react-router-dom";

import AppLayout from "src/components/layouts/AppLayout";
import { AppsRedirect } from "src/components/redirects/AppsRedirect";
import { DefaultRedirect } from "src/components/redirects/DefaultRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { ThemeProvider } from "src/components/ui/theme-provider";
import About from "src/screens/About";
import AppCreated from "src/screens/apps/AppCreated";
import AppList from "src/screens/apps/AppList";
import NewApp from "src/screens/apps/NewApp";
import ShowApp from "src/screens/apps/ShowApp";
import AppStore from "src/screens/appstore/AppStore";
import { BackupMnemonic } from "src/screens/BackupMnemonic";
import Channels from "src/screens/channels/Channels";
import FirstChannel from "src/screens/channels/FirstChannel";
import MigrateAlbyFunds from "src/screens/channels/MigrateAlbyFunds";
import NewBlocktankChannel from "src/screens/channels/NewBlocktankChannel";
import NewChannel from "src/screens/channels/NewChannel";
import NewCustomChannel from "src/screens/channels/NewCustomChannel";
import NewInstantChannel from "src/screens/channels/NewInstantChannel";
import RecommendedChannels from "src/screens/channels/RecommendedChannels";
import NotFound from "src/screens/NotFound";
import NewOnchainAddress from "src/screens/onchain/NewAddress";
import Settings from "src/screens/settings/Settings";
import { ImportMnemonic } from "src/screens/setup/ImportMnemonic";
import { SetupFinish } from "src/screens/setup/SetupFinish";
import { SetupNode } from "src/screens/setup/SetupNode";
import { SetupPassword } from "src/screens/setup/SetupPassword";
import { SetupWallet } from "src/screens/setup/SetupWallet";
import Start from "src/screens/Start";
import Unlock from "src/screens/Unlock";
import Wallet from "src/screens/wallet";
import { Welcome } from "src/screens/Welcome";
import { usePosthog } from "./hooks/usePosthog";

import TwoColumnFullScreenLayout from "src/components/layouts/TwoColumnFullScreenLayout";
import Toaster from "src/components/Toast/Toaster";

function App() {
  usePosthog();
  return (
    <>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <Toaster />
        <HashRouter>
          <Routes>
            <Route path="/" element={<AppLayout />}>
              <Route path="" element={<HomeRedirect />} />
              <Route path="settings" element={<DefaultRedirect />}>
                <Route index element={<Settings />} />
              </Route>
              <Route path="wallet" element={<DefaultRedirect />}>
                <Route index element={<Wallet />} />
              </Route>
              {/* TODO: move this under settings later */}
              <Route path="backup" element={<DefaultRedirect />}>
                <Route path="mnemonic" element={<BackupMnemonic />} />
              </Route>
              <Route path="appstore" element={<DefaultRedirect />}>
                <Route index element={<AppStore />} />
              </Route>
              <Route path="apps" element={<AppsRedirect />}>
                <Route path="new" element={<NewApp />} />
                <Route index path="" element={<AppList />} />
                <Route path=":pubkey" element={<ShowApp />} />
                <Route path="created" element={<AppCreated />} />
              </Route>
              <Route path="channels" element={<DefaultRedirect />}>
                <Route path="" element={<Channels />} />
                <Route path="first" element={<FirstChannel />} />
                <Route path="migrate-alby" element={<MigrateAlbyFunds />} />
                <Route path="new" element={<NewChannel />} />
                <Route path="new/instant" element={<NewInstantChannel />} />
                <Route path="new/blocktank" element={<NewBlocktankChannel />} />
                <Route path="recommended" element={<RecommendedChannels />} />
                <Route path="new/custom" element={<NewCustomChannel />} />
                <Route
                  path="onchain/new-address"
                  element={<NewOnchainAddress />}
                />
              </Route>
            </Route>
            <Route element={<TwoColumnFullScreenLayout />}>
              <Route
                path="start"
                element={
                  <StartRedirect>
                    <Start />
                  </StartRedirect>
                }
              ></Route>
              <Route path="unlock" element={<Unlock />} />
              <Route path="welcome" element={<Welcome />}></Route>
              <Route path="setup" element={<SetupRedirect />}>
                <Route path="" element={<Navigate to="password" replace />} />
                <Route path="password" element={<SetupPassword />} />
                <Route path="node" element={<SetupNode />} />
                <Route path="wallet" element={<SetupWallet />} />
                <Route path="import-mnemonic" element={<ImportMnemonic />} />
                <Route path="finish" element={<SetupFinish />} />
              </Route>
            </Route>
            <Route path="about" element={<About />} />
            <Route path="/*" element={<NotFound />} />
          </Routes>
        </HashRouter>{" "}
      </ThemeProvider>
    </>
  );
}

export default App;
