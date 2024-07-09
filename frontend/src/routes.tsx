import { Navigate } from "react-router-dom";
import AppLayout from "src/components/layouts/AppLayout";
import SettingsLayout from "src/components/layouts/SettingsLayout";
import TwoColumnFullScreenLayout from "src/components/layouts/TwoColumnFullScreenLayout";
import { DefaultRedirect } from "src/components/redirects/DefaultRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import { OnboardingRedirect } from "src/components/redirects/OnboardingRedirect";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { BackupMnemonic } from "src/screens/BackupMnemonic";
import { BackupNode } from "src/screens/BackupNode";
import { BackupNodeSuccess } from "src/screens/BackupNodeSuccess";
import Home from "src/screens/Home";
import { Intro } from "src/screens/Intro";
import NotFound from "src/screens/NotFound";
import Start from "src/screens/Start";
import Unlock from "src/screens/Unlock";
import { Welcome } from "src/screens/Welcome";
import AlbyAuthRedirect from "src/screens/alby/AlbyAuthRedirect";
import AppCreated from "src/screens/apps/AppCreated";
import AppList from "src/screens/apps/AppList";
import NewApp from "src/screens/apps/NewApp";
import ShowApp from "src/screens/apps/ShowApp";
import AppStore from "src/screens/appstore/AppStore";
import Channels from "src/screens/channels/Channels";
import { CurrentChannelOrder } from "src/screens/channels/CurrentChannelOrder";
import IncreaseIncomingCapacity from "src/screens/channels/IncreaseIncomingCapacity";
import IncreaseOutgoingCapacity from "src/screens/channels/IncreaseOutgoingCapacity";
import MigrateAlbyFunds from "src/screens/onboarding/MigrateAlbyFunds";
import { Success } from "src/screens/onboarding/Success";
import BuyBitcoin from "src/screens/onchain/BuyBitcoin";
import DepositBitcoin from "src/screens/onchain/DepositBitcoin";
import ConnectPeer from "src/screens/peers/ConnectPeer";
import Peers from "src/screens/peers/Peers";
import { ChangeUnlockPassword } from "src/screens/settings/ChangeUnlockPassword";
import DebugTools from "src/screens/settings/DebugTools";
import Settings from "src/screens/settings/Settings";
import { ImportMnemonic } from "src/screens/setup/ImportMnemonic";
import { RestoreNode } from "src/screens/setup/RestoreNode";
import { SetupAdvanced } from "src/screens/setup/SetupAdvanced";
import { SetupFinish } from "src/screens/setup/SetupFinish";
import { SetupNode } from "src/screens/setup/SetupNode";
import { SetupPassword } from "src/screens/setup/SetupPassword";
import { BreezForm } from "src/screens/setup/node/BreezForm";
import { CashuForm } from "src/screens/setup/node/CashuForm";
import { GreenlightForm } from "src/screens/setup/node/GreenlightForm";
import { LDKForm } from "src/screens/setup/node/LDKForm";
import { LNDForm } from "src/screens/setup/node/LNDForm";
import { PhoenixdForm } from "src/screens/setup/node/PhoenixdForm";
import { PresetNodeForm } from "src/screens/setup/node/PresetNodeForm";
import Wallet from "src/screens/wallet";
import Receive from "src/screens/wallet/Receive";
import Send from "src/screens/wallet/Send";
import SignMessage from "src/screens/wallet/SignMessage";

const routes = [
  {
    path: "/",
    element: <AppLayout />,
    handle: { crumb: () => "Home" },
    children: [
      {
        index: true,
        element: <HomeRedirect />,
      },
      {
        path: "home",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Dashboard" },
        children: [
          {
            index: true,
            element: <Home />,
          },
        ],
      },
      {
        path: "wallet",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Wallet" },
        children: [
          {
            index: true,
            element: <Wallet />,
          },
          {
            path: "receive",
            element: <Receive />,
            handle: { crumb: () => "Receive" },
          },
          {
            path: "send",
            element: <Send />,
            handle: { crumb: () => "Send" },
          },
          {
            path: "sign-message",
            element: <SignMessage />,
            handle: { crumb: () => "Sign Message" },
          },
        ],
      },
      {
        path: "settings",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Settings" },
        children: [
          {
            path: "",
            element: <SettingsLayout />,
            children: [
              {
                index: true,
                element: <Settings />,
              },
              {
                path: "change-unlock-password",
                element: <ChangeUnlockPassword />,
                handle: { crumb: () => "Unlock Password" },
              },
              {
                path: "key-backup",
                element: <BackupMnemonic />,
                handle: { crumb: () => "Key Backup" },
              },
              {
                path: "node-backup",
                element: <BackupNode />,
              },
            ],
          },
        ],
      },
      {
        path: "apps",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Connections" },
        children: [
          {
            index: true,
            element: <AppList />,
          },

          {
            path: ":pubkey",
            element: <ShowApp />,
          },
          {
            path: "new",
            element: <NewApp />,
            handle: { crumb: () => "New App" },
          },
          {
            path: "created",
            element: <AppCreated />,
          },
        ],
      },
      {
        path: "appstore",
        element: <DefaultRedirect />,
        handle: { crumb: () => "App Store" },
        children: [
          {
            index: true,
            element: <AppStore />,
          },
        ],
      },
      {
        path: "channels",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Node" },
        children: [
          {
            index: true,
            element: <Channels />,
          },
          {
            path: "outgoing",
            element: <IncreaseOutgoingCapacity />,
            handle: { crumb: () => "Increase Spending Balance" },
          },
          {
            path: "incoming",
            element: <IncreaseIncomingCapacity />,
            handle: { crumb: () => "Increase Receiving Capacity" },
          },
          {
            path: "order",
            element: <CurrentChannelOrder />,
            handle: { crumb: () => "Current Order" },
          },
          {
            path: "onchain/buy-bitcoin",
            element: <BuyBitcoin />,
            handle: { crumb: () => "Buy Bitcoin" },
          },
          {
            path: "onchain/deposit-bitcoin",
            element: <DepositBitcoin />,
            handle: { crumb: () => "Deposit Bitcoin" },
          },
        ],
      },
      {
        path: "peers",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Peers" },
        children: [
          {
            index: true,
            element: <Peers />,
          },
          {
            path: "new",
            element: <ConnectPeer />,
            handle: { crumb: () => "Connect Peer" },
          },
        ],
      },
      {
        path: "debug-tools",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Debug" },
        children: [
          {
            index: true,
            element: <DebugTools />,
          },
        ],
      },
    ],
  },
  {
    element: <TwoColumnFullScreenLayout />,
    children: [
      {
        path: "start",
        element: (
          <StartRedirect>
            <Start />
          </StartRedirect>
        ),
      },
      {
        path: "alby/auth",
        element: <AlbyAuthRedirect />,
      },
      {
        path: "unlock",
        element: <Unlock />,
      },
      {
        path: "welcome",
        element: <Welcome />,
      },
      {
        path: "setup",
        element: <SetupRedirect />,
        children: [
          {
            element: <Navigate to="password" replace />,
          },
          {
            path: "password",
            element: <SetupPassword />,
          },
          {
            path: "node",
            children: [
              {
                index: true,
                element: <SetupNode />,
              },
              {
                path: "breez",
                element: <BreezForm />,
              },
              {
                path: "greenlight",
                element: <GreenlightForm />,
              },
              {
                path: "cashu",
                element: <CashuForm />,
              },
              {
                path: "phoenix",
                element: <PhoenixdForm />,
              },
              {
                path: "lnd",
                element: <LNDForm />,
              },
              {
                path: "ldk",
                element: <LDKForm />,
              },
              {
                path: "preset",
                element: <PresetNodeForm />,
              },
            ],
          },
          {
            path: "advanced",
            element: <SetupAdvanced />,
          },
          {
            path: "import-mnemonic",
            element: <ImportMnemonic />,
          },
          {
            path: "node-restore",
            element: <RestoreNode />,
          },
          {
            path: "finish",
            element: <SetupFinish />,
          },
        ],
      },
      {
        path: "onboarding",
        element: <OnboardingRedirect />,
        children: [
          {
            path: "lightning/migrate-alby",
            element: <MigrateAlbyFunds />,
          },
          {
            path: "success",
            element: <Success />,
          },
        ],
      },
      {
        path: "alby/auth",
        element: <AlbyAuthRedirect />,
      },
    ],
  },
  {
    path: "node-backup-success",
    element: <BackupNodeSuccess />,
  },
  {
    path: "intro",
    element: <Intro />,
  },
  {
    path: "/*",
    element: <NotFound />,
  },
];

export default routes;
