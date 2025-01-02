import { Navigate } from "react-router-dom";
import AppLayout from "src/components/layouts/AppLayout";
import ReceiveLayout from "src/components/layouts/ReceiveLayout";
import SendLayout from "src/components/layouts/SendLayout";
import SettingsLayout from "src/components/layouts/SettingsLayout";
import TwoColumnFullScreenLayout from "src/components/layouts/TwoColumnFullScreenLayout";
import { DefaultRedirect } from "src/components/redirects/DefaultRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { BackupMnemonic } from "src/screens/BackupMnemonic";
import { BackupNode } from "src/screens/BackupNode";
import { BackupNodeSuccess } from "src/screens/BackupNodeSuccess";
import { ConnectAlbyAccount } from "src/screens/ConnectAlbyAccount";
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
import { AppStoreDetail } from "src/screens/appstore/AppStoreDetail";
import Channels from "src/screens/channels/Channels";
import { CurrentChannelOrder } from "src/screens/channels/CurrentChannelOrder";
import IncreaseIncomingCapacity from "src/screens/channels/IncreaseIncomingCapacity";
import IncreaseOutgoingCapacity from "src/screens/channels/IncreaseOutgoingCapacity";
import { AutoChannel } from "src/screens/channels/auto/AutoChannel";
import { OpenedAutoChannel } from "src/screens/channels/auto/OpenedAutoChannel";
import { OpeningAutoChannel } from "src/screens/channels/auto/OpeningAutoChannel";
import { FirstChannel } from "src/screens/channels/first/FirstChannel";
import { OpenedFirstChannel } from "src/screens/channels/first/OpenedFirstChannel";
import { OpeningFirstChannel } from "src/screens/channels/first/OpeningFirstChannel";
import { BuzzPay } from "src/screens/internal-apps/BuzzPay";
import { UncleJim } from "src/screens/internal-apps/UncleJim";
import BuyBitcoin from "src/screens/onchain/BuyBitcoin";
import DepositBitcoin from "src/screens/onchain/DepositBitcoin";
import ConnectPeer from "src/screens/peers/ConnectPeer";
import Peers from "src/screens/peers/Peers";
import { AlbyAccount } from "src/screens/settings/AlbyAccount";
import Backup from "src/screens/settings/Backup";
import { ChangeUnlockPassword } from "src/screens/settings/ChangeUnlockPassword";
import DebugTools from "src/screens/settings/DebugTools";
import DeveloperSettings from "src/screens/settings/DeveloperSettings";
import Settings from "src/screens/settings/Settings";
import Shutdown from "src/screens/settings/Shutdown";
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
import WithdrawOnchainFunds from "src/screens/wallet/WithdrawOnchainFunds";
import ReceiveInvoice from "src/screens/wallet/receive/ReceiveInvoice";
import ConfirmPayment from "src/screens/wallet/send/ConfirmPayment";
import LnurlPay from "src/screens/wallet/send/LnurlPay";
import PaymentSuccess from "src/screens/wallet/send/PaymentSuccess";
import ZeroAmount from "src/screens/wallet/send/ZeroAmount";

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
            handle: { crumb: () => "Receive" },
            element: <ReceiveLayout />,
            children: [
              {
                index: true,
                element: <Receive />,
              },
              {
                handle: { crumb: () => "Invoice" },
                path: "invoice",
                element: <ReceiveInvoice />,
              },
            ],
          },
          {
            path: "send",
            element: <SendLayout />,
            handle: { crumb: () => "Send" },
            children: [
              {
                index: true,
                element: <Send />,
              },
              {
                path: "lnurl-pay",
                element: <LnurlPay />,
              },
              {
                path: "0-amount",
                element: <ZeroAmount />,
              },
              {
                path: "confirm-payment",
                element: <ConfirmPayment />,
              },
              {
                path: "success",
                element: <PaymentSuccess />,
              },
            ],
          },
          {
            path: "sign-message",
            element: <SignMessage />,
            handle: { crumb: () => "Sign Message" },
          },
          {
            path: "withdraw",
            element: <WithdrawOnchainFunds />,
            handle: { crumb: () => "Withdraw On-Chain Balance" },
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
                path: "backup",
                element: <Backup />,
                handle: { crumb: () => "Backup" },
              },
              {
                path: "mnemonic-backup",
                element: <BackupMnemonic />,
                handle: { crumb: () => "Key Backup" },
              },
              {
                path: "node-backup",
                element: <BackupNode />,
              },
              {
                path: "alby-account",
                element: <AlbyAccount />,
              },
              {
                path: "developer",
                element: <DeveloperSettings />,
              },
              {
                path: "debug-tools",
                element: <DebugTools />,
              },
              {
                path: "shutdown",
                element: <Shutdown />,
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
        path: "internal-apps",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Connections" },
        children: [
          {
            path: "uncle-jim",
            element: <UncleJim />,
          },
          {
            path: "buzzpay",
            element: <BuzzPay />,
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
          {
            path: ":appId",
            element: <AppStoreDetail />,
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
            path: "first",
            handle: { crumb: () => "Your First Channel" },
            children: [
              {
                index: true,
                element: <FirstChannel />,
              },
              {
                path: "opening",
                element: <OpeningFirstChannel />,
              },
              {
                path: "opened",
                element: <OpenedFirstChannel />,
              },
            ],
          },
          {
            path: "auto",
            handle: { crumb: () => "New Channel" },
            children: [
              {
                index: true,
                element: <AutoChannel />,
              },
              {
                path: "opening",
                element: <OpeningAutoChannel />,
              },
              {
                path: "opened",
                element: <OpenedAutoChannel />,
              },
            ],
          },
          {
            path: "outgoing",
            element: <IncreaseOutgoingCapacity />,
            handle: { crumb: () => "Open Channel with On-Chain" },
          },
          {
            path: "incoming",
            element: <IncreaseIncomingCapacity />,
            handle: { crumb: () => "Open Channel with Lightning" },
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
        path: "alby/account",
        element: <ConnectAlbyAccount />,
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
            path: "alby",
            element: <ConnectAlbyAccount connectUrl="/setup/auth" />,
          },
          {
            path: "auth",
            element: <AlbyAuthRedirect />,
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
