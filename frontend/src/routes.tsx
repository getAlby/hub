import { Navigate, RouteObject } from "react-router-dom";
import AppLayout from "src/components/layouts/AppLayout";
import SettingsLayout from "src/components/layouts/SettingsLayout";
import TwoColumnFullScreenLayout from "src/components/layouts/TwoColumnFullScreenLayout";
import { DefaultRedirect } from "src/components/redirects/DefaultRedirect";
import { HomeRedirect } from "src/components/redirects/HomeRedirect";
import { SetupRedirect } from "src/components/redirects/SetupRedirect";
import { StartRedirect } from "src/components/redirects/StartRedirect";
import { ConnectAlbyAccount } from "src/screens/ConnectAlbyAccount";
import { CreateNodeMigrationFileSuccess } from "src/screens/CreateNodeMigrationFileSuccess";
import Home from "src/screens/Home";
import { Intro } from "src/screens/Intro";
import { MigrateNode } from "src/screens/MigrateNode";
import NotFound from "src/screens/NotFound";
import Start from "src/screens/Start";
import Unlock from "src/screens/Unlock";
import { Welcome } from "src/screens/Welcome";
import AlbyAuthRedirect from "src/screens/alby/AlbyAuthRedirect";
import { AlbyReviews } from "src/screens/alby/AlbyReviews";
import SupportAlby from "src/screens/alby/SupportAlby";
import AppDetails from "src/screens/apps/AppDetails";
import { AppsCleanup } from "src/screens/apps/AppsCleanup";
import { Connections } from "src/screens/apps/Connections";
import NewApp from "src/screens/apps/NewApp";
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
import { Bitrefill } from "src/screens/internal-apps/Bitrefill";
import { BuzzPay } from "src/screens/internal-apps/BuzzPay";
import { Claude } from "src/screens/internal-apps/Claude";
import { Goose } from "src/screens/internal-apps/Goose";
import { LightningMessageboard } from "src/screens/internal-apps/LightningMessageboard";
import { SimpleBoost } from "src/screens/internal-apps/SimpleBoost";
import { Tictactoe } from "src/screens/internal-apps/Tictactoe";
import { ZapPlanner } from "src/screens/internal-apps/ZapPlanner";
import BuyBitcoin from "src/screens/onchain/BuyBitcoin";
import DepositBitcoin from "src/screens/onchain/DepositBitcoin";
import ConnectPeer from "src/screens/peers/ConnectPeer";
import Peers from "src/screens/peers/Peers";
import { About } from "src/screens/settings/About";
import { AlbyAccount } from "src/screens/settings/AlbyAccount";
import { AutoUnlock } from "src/screens/settings/AutoUnlock";
import Backup from "src/screens/settings/Backup";
import { ChangeUnlockPassword } from "src/screens/settings/ChangeUnlockPassword";
import DebugTools from "src/screens/settings/DebugTools";
import DeveloperSettings from "src/screens/settings/DeveloperSettings";
import Settings from "src/screens/settings/Settings";
import TorSettings from "src/screens/settings/TorSettings";

import { ImportMnemonic } from "src/screens/setup/ImportMnemonic";
import { RestoreNode } from "src/screens/setup/RestoreNode";
import { SetupAdvanced } from "src/screens/setup/SetupAdvanced";
import { SetupFinish } from "src/screens/setup/SetupFinish";
import { SetupNode } from "src/screens/setup/SetupNode";
import { SetupPassword } from "src/screens/setup/SetupPassword";
import { SetupSecurity } from "src/screens/setup/SetupSecurity";
import { CashuForm } from "src/screens/setup/node/CashuForm";
import { LDKForm } from "src/screens/setup/node/LDKForm";
import { LNDForm } from "src/screens/setup/node/LNDForm";
import { PhoenixdForm } from "src/screens/setup/node/PhoenixdForm";
import { PresetNodeForm } from "src/screens/setup/node/PresetNodeForm";
import { NewSubwallet } from "src/screens/subwallets/NewSubwallet";
import { SubwalletCreated } from "src/screens/subwallets/SubwalletCreated";
import { SubwalletList } from "src/screens/subwallets/SubwalletList";
import Wallet from "src/screens/wallet";
import NodeAlias from "src/screens/wallet/NodeAlias";
import Receive from "src/screens/wallet/Receive";
import Send from "src/screens/wallet/Send";
import SignMessage from "src/screens/wallet/SignMessage";
import WithdrawOnchainFunds from "src/screens/wallet/WithdrawOnchainFunds";
import ReceiveInvoice from "src/screens/wallet/receive/ReceiveInvoice";
import ReceiveOffer from "src/screens/wallet/receive/ReceiveOffer";
import ReceiveOnchain from "src/screens/wallet/receive/ReceiveOnchain";
import ConfirmPayment from "src/screens/wallet/send/ConfirmPayment";
import LnurlPay from "src/screens/wallet/send/LnurlPay";
import Onchain from "src/screens/wallet/send/Onchain";
import OnchainSuccess from "src/screens/wallet/send/OnchainSuccess";
import PaymentSuccess from "src/screens/wallet/send/PaymentSuccess";
import ZeroAmount from "src/screens/wallet/send/ZeroAmount";
import Swap from "src/screens/wallet/swap";
import AutoSwap from "src/screens/wallet/swap/AutoSwap";
import SwapInStatus from "src/screens/wallet/swap/SwapInStatus";
import SwapOutStatus from "src/screens/wallet/swap/SwapOutStatus";

const routes: RouteObject[] = [
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
            path: "swap",
            handle: { crumb: () => "Swap" },
            children: [
              {
                index: true,
                element: <Swap />,
              },
              {
                path: "out/status/:swapId",
                element: <SwapOutStatus />,
              },
              {
                path: "in/status/:swapId",
                element: <SwapInStatus />,
              },
              {
                path: "auto",
                element: <AutoSwap />,
              },
            ],
          },
          {
            path: "receive",
            handle: { crumb: () => "Receive" },
            children: [
              {
                index: true,
                element: <Receive />,
              },
              {
                handle: { crumb: () => "Receive On-chain" },
                path: "onchain",
                element: <ReceiveOnchain />,
              },
              {
                handle: { crumb: () => "Invoice" },
                path: "invoice",
                element: <ReceiveInvoice />,
              },
              {
                handle: { crumb: () => "BOLT-12 Offer" },
                path: "offer",
                element: <ReceiveOffer />,
              },
            ],
          },
          {
            path: "send",
            handle: { crumb: () => "Send" },
            children: [
              {
                index: true,
                element: <Send />,
              },
              {
                path: "onchain",
                element: <Onchain />,
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
                path: "onchain-success",
                element: <OnchainSuccess />,
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
            path: "node-alias",
            element: <NodeAlias />,
            handle: { crumb: () => "Node Alias" },
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
                path: "about",
                element: <About />,
                handle: { crumb: () => "About" },
              },
              {
                path: "auto-unlock",
                element: <AutoUnlock />,
                handle: { crumb: () => "Auto Unlock" },
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
                path: "node-migrate",
                element: <MigrateNode />,
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
                path: "tor",
                element: <TorSettings />,
                handle: { crumb: () => "Tor" },
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
            element: <Connections />,
          },
          {
            path: ":id",
            element: <AppDetails />,
          },
          {
            path: "new",
            element: <NewApp />,
            handle: { crumb: () => "New App" },
          },
          {
            path: "cleanup",
            element: <AppsCleanup />,
          },
        ],
      },
      {
        path: "sub-wallets",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Sub-wallets" },

        children: [
          {
            index: true,
            element: <SubwalletList />,
          },
          {
            path: "new",
            element: <NewSubwallet />,
          },
          {
            path: "created",
            element: <SubwalletCreated />,
          },
        ],
      },
      {
        path: "internal-apps",
        element: <DefaultRedirect />,
        handle: { crumb: () => "Connections" },
        children: [
          {
            path: "buzzpay",
            element: <BuzzPay />,
          },
          {
            path: "simpleboost",
            element: <SimpleBoost />,
          },
          {
            path: "lightning-messageboard",
            element: <LightningMessageboard />,
          },
          {
            path: "zapplanner",
            element: <ZapPlanner />,
          },
          {
            path: "bitrefill",
            element: <Bitrefill />,
          },
          {
            path: "goose",
            element: <Goose />,
          },
          {
            path: "claude",
            element: <Claude />,
          },
          {
            path: "tictactoe",
            element: <Tictactoe />,
          },
        ],
      },
      {
        path: "appstore",
        element: <DefaultRedirect />,
        handle: { crumb: () => "App Store" },
        children: [
          {
            path: ":appStoreId",
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
      {
        path: "support-alby",
        element: <SupportAlby />,
      },
      {
        path: "review-earn",
        element: <AlbyReviews />,
        handle: { crumb: () => "Review & Earn" },
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
            path: "security",
            element: <SetupSecurity />,
          },
          {
            path: "node",
            children: [
              {
                index: true,
                element: <SetupNode />,
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
    path: "create-node-migration-file-success",
    element: <CreateNodeMigrationFileSuccess />,
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
