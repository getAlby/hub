import { HashRouter, Navigate, Route, Routes } from "react-router-dom";

import AppLayout from "src/components/layouts/AppLayout";
import { DefaultRedirect } from "src/components/redirects/DefaultRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { ThemeProvider } from "src/components/ui/theme-provider";
import { BackupMnemonic } from "src/screens/BackupMnemonic";
import NotFound from "src/screens/NotFound";
import Start from "src/screens/Start";
import Unlock from "src/screens/Unlock";
import { Welcome } from "src/screens/Welcome";
import AppCreated from "src/screens/apps/AppCreated";
import AppList from "src/screens/apps/AppList";
import NewApp from "src/screens/apps/NewApp";
import ShowApp from "src/screens/apps/ShowApp";
import AppStore from "src/screens/appstore/AppStore";
import Channels from "src/screens/channels/Channels";
import NewChannel from "src/screens/channels/NewChannel";
import MigrateAlbyFunds from "src/screens/onboarding/MigrateAlbyFunds";
import NewOnchainAddress from "src/screens/onchain/NewAddress";
import ConnectPeer from "src/screens/peers/ConnectPeer";
import Settings from "src/screens/settings/Settings";
import { ImportMnemonic } from "src/screens/setup/ImportMnemonic";
import { SetupFinish } from "src/screens/setup/SetupFinish";
import { SetupNode } from "src/screens/setup/SetupNode";
import { SetupPassword } from "src/screens/setup/SetupPassword";
import { SetupWallet } from "src/screens/setup/SetupWallet";
import Wallet from "src/screens/wallet";
import { usePosthog } from "./hooks/usePosthog";

import SettingsLayout from "src/components/layouts/SettingsLayout";
import TwoColumnFullScreenLayout from "src/components/layouts/TwoColumnFullScreenLayout";
import { OnboardingRedirect } from "src/components/redirects/OnboardingRedirect";
import { Toaster } from "src/components/ui/toaster";
import { Intro } from "src/screens/Intro";
import AlbyAuthRedirect from "src/screens/alby/AlbyAuthRedirect";
import { CurrentChannelOrder } from "src/screens/channels/CurrentChannelOrder";
import { Success } from "src/screens/onboarding/Success";
import { ChangeUnlockPassword } from "src/screens/settings/ChangeUnlockPassword";
import DebugTools from "src/screens/settings/DebugTools";

function App() {
  usePosthog();
  return (
    <>
      <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
        <Toaster />
        <HashRouter>
          <Routes>
            <Route path="/" element={<AppLayout />}>
              <Route path="" element={<HomeRedirect />} />
              <Route path="settings" element={<DefaultRedirect />}>
                <Route element={<SettingsLayout />}>
                  <Route index element={<Settings />} />
                  <Route
                    path="change-unlock-password"
                    element={<ChangeUnlockPassword />}
                  />
                  <Route path="backup" element={<BackupMnemonic />} />
                </Route>
              </Route>
              <Route path="wallet" element={<DefaultRedirect />}>
                <Route index element={<Wallet />} />
              </Route>
              <Route path="appstore" element={<DefaultRedirect />}>
                <Route index element={<AppStore />} />
              </Route>
              <Route path="apps" element={<DefaultRedirect />}>
                <Route path="new" element={<NewApp />} />
                <Route index path="" element={<AppList />} />
                <Route path=":pubkey" element={<ShowApp />} />
                <Route path="created" element={<AppCreated />} />
              </Route>
              <Route path="debug-tools" element={<DefaultRedirect />}>
                <Route index element={<DebugTools />} />
              </Route>
              <Route path="channels" element={<DefaultRedirect />}>
                <Route index path="" element={<Channels />} />
                <Route path="new" element={<NewChannel />} />
                <Route path="order" element={<CurrentChannelOrder />} />
                <Route
                  path="onchain/new-address"
                  element={<NewOnchainAddress />}
                />
              </Route>
              <Route path="peers" element={<DefaultRedirect />}>
                <Route path="new" element={<ConnectPeer />} />
              </Route>
            </Route>
            <Route path="intro" element={<Intro />} />
            <Route element={<TwoColumnFullScreenLayout />}>
              <Route
                path="start"
                element={
                  <StartRedirect>
                    <Start />
                  </StartRedirect>
                }
              />
              <Route path="/alby/auth" element={<AlbyAuthRedirect />} />
              <Route path="unlock" element={<Unlock />} />
              <Route path="welcome" element={<Welcome />} />
              <Route path="setup" element={<SetupRedirect />}>
                <Route path="" element={<Navigate to="password" replace />} />
                <Route path="password" element={<SetupPassword />} />
                <Route path="node" element={<SetupNode />} />
                <Route path="wallet" element={<SetupWallet />} />
                <Route path="import-mnemonic" element={<ImportMnemonic />} />
                <Route path="finish" element={<SetupFinish />} />
              </Route>
              <Route path="onboarding" element={<OnboardingRedirect />}>
                <Route path="lightning">
                  <Route path="migrate-alby" element={<MigrateAlbyFunds />} />
                </Route>
                <Route path="success" element={<Success />} />
              </Route>
            </Route>
            <Route path="/*" element={<NotFound />} />
          </Routes>
        </HashRouter>
      </ThemeProvider>
    </>
  );
}

export default App;
