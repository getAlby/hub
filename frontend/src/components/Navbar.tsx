import { Link, Outlet, useLocation } from "react-router-dom";

import nwcLogo from "src/assets/images/nwc-logo.svg";
import { useInfo } from "src/hooks/useInfo";

function Navbar() {
  const location = useLocation();
  const { data: info } = useInfo();

  const linkStyles =
    "font-medium text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 transition px-2 ";
  const selectedLinkStyles =
    "text-gray-900 hover:text-gray-900 dark:text-gray-100 dark:hover:text-gray-100";

  return (
    <>
      <div className="bg-gray-50 dark:bg-surface-00dp">
        <div className="bg-white border-b border-gray-200 dark:bg-surface-01dp dark:border-neutral-700 mb-6">
          <nav className="container max-w-screen-lg mx-auto px-4 2xl:px-0 py-3">
            <div className="flex justify-between items-center">
              <div className="flex items-center space-x-12">
                <Link
                  to="/"
                  className="font-headline text-[20px] dark:text-white flex gap-2 justify-center items-center"
                >
                  <img
                    alt="NWC Logo"
                    className="w-8 inline"
                    width="128"
                    height="120"
                    src={nwcLogo}
                  />
                  <span className="dark:text-white text-lg font-semibold hidden sm:inline">
                    Nostr Wallet Connect
                  </span>
                </Link>

                <div className="flex space-x-4">
                  <Link
                    className={`${linkStyles} ${
                      location.pathname.startsWith("/apps") &&
                      selectedLinkStyles
                    }`}
                    to="/apps"
                  >
                    Connections
                  </Link>
                  {info?.running && info.backendType === "GREENLIGHT" && (
                    <Link
                      className={`${linkStyles} ${
                        location.pathname.startsWith("/channels") &&
                        selectedLinkStyles
                      }`}
                      to="/channels"
                    >
                      Channels
                    </Link>
                  )}
                  <Link
                    className={`${linkStyles} ${
                      location.pathname === "/setup" && selectedLinkStyles
                    }`}
                    to="/setup"
                  >
                    Setup
                  </Link>
                  <Link
                    className={`${linkStyles} ${
                      location.pathname === "/about" && selectedLinkStyles
                    }`}
                    to="/about"
                  >
                    About
                  </Link>
                </div>
              </div>
            </div>
          </nav>
        </div>
      </div>
      <div className="flex justify-center">
        <div className="container max-w-screen-lg px-2">
          <Outlet />
        </div>
      </div>
    </>
  );
}

export default Navbar;
