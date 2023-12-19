import { useState } from 'react';
// import QrCreator from 'qr-creator'; // Ensure this module is installed and imported correctly

const Create = () => {
  const pairingUri = "YOUR_PAIRING_URI"; // Replace with actual data or props
  const [copied, setCopied] = useState(false);

  // useEffect(() => {
  //   // This would replace the script to create QR Code
  //   QrCreator.render(
  //     {
  //       fill: window.matchMedia('(prefers-color-scheme: dark)').matches ? "#FFF" : "#000",
  //       text: pairingUri,
  //       size: 250,
  //     },
  //     document.getElementById("connect-qrcode")
  //   );
  // }, [pairingUri]);

  const copyToClipboard = async () => {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(pairingUri);
    } else {
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = pairingUri;
      textArea.style.position = "absolute";
      textArea.style.opacity = "0";
      document.body.appendChild(textArea);
      textArea.select();
      await document.execCommand("copy");
      textArea.remove();
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="w-full max-w-screen-sm mx-auto">
      <h2 className="font-bold text-2xl font-headline mb-2 dark:text-white text-center">
        🚀 Almost there!
      </h2>
      <div className="font-medium text-center mb-8 dark:text-white">
        Complete the last step of the setup by pasting or scanning your connection's pairing secret in the desired app to finalise the connection.
      </div>

      <div className="flex flex-col">
        <a href={pairingUri} className="w-full inline-flex bg-purple-700 cursor-pointer duration-150 focus:outline-none font-medium hover:bg-purple-900 items-center justify-center px-5 py-4 rounded-md shadow text-white transition mb-2">
          {/* SVGs are preserved */}
          {/* ... SVG content */}
          Open in supported app
        </a>
        <div className="text-center text-xs text-gray-600 dark:text-neutral-500">
          Only connect with apps you trust!
        </div>

        <div className="dark:text-white text-sm text-center mt-8 mb-1">Manually pair app ↓</div>
        <button id="copy-button" className={`w-full inline-flex items-center justify-center px-3 py-2 cursor-pointer duration-150 transition bg-white text-purple-700 dark:bg-surface-02dp dark:text-neutral-200 border dark:border-white/10 ${copied ? 'bg-green-600 text-white' : 'hover:bg-gray-50 dark:hover:bg-surface-16dp'} bg-origin-border rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-primary mt-2 mb-2`} onClick={copyToClipboard}>
          {/* ... SVG content */}
          <span id="copy-text">{copied ? "Copied to clipboard!" : "Copy pairing secret"}</span>
        </button>

        {/* ... Remaining JSX conversion for QR code and other elements */}
      </div>
    </div>
  );
};

export default Create;