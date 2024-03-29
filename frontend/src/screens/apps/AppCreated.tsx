import { CopyIcon, EyeIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, Navigate, useLocation } from "react-router-dom";

import QRCode from "src/components/QRCode";
import { NostrWalletConnectIcon } from "src/components/icons/NostrWalletConnectIcon";
import { Button } from "src/components/ui/button";
import { CreateAppResponse } from "src/types";

export default function AppCreated() {
  const { state } = useLocation();
  const createAppResponse = state as CreateAppResponse;

  const [copied, setCopied] = useState(false);
  const [isPopupVisible, setPopupVisible] = useState(false);
  const popupRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // dispatch a success event which can be listened to by the opener or by the app that embedded the webview
    // this gives those apps the chance to know the user has enabled the connection
    const nwcEvent = new CustomEvent("nwc:success", { detail: {} });
    window.dispatchEvent(nwcEvent);

    // notify the opener of the successful connection
    if (window.opener) {
      window.opener.postMessage(
        {
          type: "nwc:success",
          payload: { success: true },
        },
        "*"
      );
    }
  }, []);

  // TODO: use a modal library instead of doing this manually
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        popupRef.current &&
        !popupRef.current.contains(event.target as Node)
      ) {
        setPopupVisible(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  if (!createAppResponse) {
    return <Navigate to="/apps/new" />;
  }

  const pairingUri = createAppResponse.pairingUri;

  const copyToClipboard = () => {
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(pairingUri);
    } else {
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = pairingUri;
      textArea.style.position = "absolute";
      textArea.style.opacity = "0";
      document.body.appendChild(textArea);
      textArea.select();
      new Promise((res, rej) => {
        document.execCommand("copy") ? res(pairingUri) : rej();
        textArea.remove();
      });
    }
    setCopied(true);
    setTimeout(() => {
      setCopied(false);
    }, 2000);
  };

  const togglePopup = () => {
    setPopupVisible(!isPopupVisible);
  };

  return (
    <div className="w-full max-w-screen-sm mx-auto mt-6 md:px-4 ph-no-capture">
      <h2 className="font-bold text-2xl font-headline mb-2 text-center">
        🚀 Almost there!
      </h2>
      <div className="font-medium text-center mb-6">
        Complete the last step of the setup by pasting or scanning your
        connection's pairing secret in the desired app to finalise the
        connection.
      </div>

      <div className="flex flex-col items-center">
        <Link to={pairingUri}>
          <Button size="lg">
            <NostrWalletConnectIcon className="inline w-6 mr-2" />
            <p className="font-medium">Open in supported app</p>
          </Button>
        </Link>
        <div className="text-center text-xs text-muted-foreground mt-2">
          Only connect with apps you trust!
        </div>

        <div className="dark:text-white text-sm text-center mt-8 mb-1"></div>
        <div className="flex flex-col gap-3">
          <div className=" text-center text-sm">Manually pair app ↓</div>
          <Button variant="secondary" onClick={copyToClipboard}>
            <CopyIcon className="inline w-6 mr-2" />
            {copied ? "Copied to clipboard!" : "Copy pairing secret"}
          </Button>
          <Button variant="secondary" onClick={togglePopup}>
            <EyeIcon className="inline w-6 mr-2" />
            QR Code
          </Button>
        </div>
      </div>
      <div
        className={`fixed inset-0 items-center justify-center ${
          isPopupVisible ? "flex" : "hidden"
        }`}
      >
        <div
          onClick={togglePopup}
          className="fixed inset-0 bg-gray-900 opacity-50"
        ></div>
        <div className="bg-white dark:bg-surface-02dp p-4 lg:px-6 rounded shadow relative">
          <h2 className="mb-4 font-semibold text-lg lg:text-xl font-headline dark:text-white">
            Scan QR Code in the app to pair
          </h2>
          <a
            href={pairingUri}
            target="_blank"
            className="block border-4 border-purple-600 rounded-lg p-4 lg:p-6"
          >
            <QRCode value={pairingUri} size={256} />
          </a>
          <button
            onClick={togglePopup}
            className="w-full inline-flex font-semibold items-center justify-center px-3 py-2 cursor-pointer duration-150 transition bg-white text-gray-700 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700 mt-4"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
