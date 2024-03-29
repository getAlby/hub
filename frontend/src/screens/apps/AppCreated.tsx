import { CopyIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Navigate, useLocation } from "react-router-dom";

import QRCode from "src/components/QRCode";
import { EyeIcon } from "src/components/icons/EyeIcon";
import { LogoIcon } from "src/components/icons/LogoIcon";
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
      <h2 className="font-bold text-2xl font-headline mb-2 dark:text-white text-center">
        🚀 Almost there!
      </h2>
      <div className="font-medium text-center mb-6 dark:text-white">
        Complete the last step of the setup by pasting or scanning your
        connection's pairing secret in the desired app to finalise the
        connection.
      </div>

      <div className="flex flex-col">
        <a
          href={pairingUri}
          className="w-full inline-flex bg-purple-700 cursor-pointer duration-150 focus:outline-none hover:bg-purple-900 items-center justify-center px-5 py-4 rounded-md shadow text-white transition mb-2"
        >
          <LogoIcon className="inline w-6 mr-2" />
          <p className="font-medium">Open in supported app</p>
        </a>
        <div className="text-center text-xs text-gray-600 dark:text-neutral-500">
          Only connect with apps you trust!
        </div>

        <div className="dark:text-white text-sm text-center mt-8 mb-1">
          Manually pair app ↓
        </div>
        <button
          id="copy-button"
          className={`w-full inline-flex items-center justify-center px-3 py-2 cursor-pointer duration-150 transition border dark:border-white/10 ${
            copied
              ? "bg-green-600 text-white"
              : "bg-white dark:bg-surface-02dp text-purple-700 dark:text-neutral-200 hover:bg-gray-50  dark:hover:bg-surface-16dp"
          } bg-origin-border rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700 my-2`}
          onClick={copyToClipboard}
        >
          <CopyIcon className="inline w-6 mr-2" />
          <span id="copy-text">
            {copied ? "Copied to clipboard!" : "Copy pairing secret"}
          </span>
        </button>

        <button
          id="copy-button"
          className={`w-full inline-flex items-center justify-center px-3 py-2 cursor-pointer duration-150 transition border dark:border-white/10 bg-white dark:bg-surface-02dp text-purple-700 dark:text-neutral-200 hover:bg-gray-50  dark:hover:bg-surface-16dp bg-origin-border rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700 mb-2`}
          onClick={togglePopup}
        >
          <EyeIcon className="inline w-6 mr-2" />
          <span id="copy-text">QR Code</span>
        </button>
        {/* ... Remaining JSX conversion for QR code and other elements */}
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
