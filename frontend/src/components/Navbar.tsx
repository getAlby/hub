import { Link, Outlet, useLocation } from "react-router-dom";

import nwcComboMark from "src/assets/images/nwc-combomark.svg";

function Navbar() {
  const location = useLocation();

  const linkStyles =
    "font-medium text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 transition px-2 ";
  const selectedLinkStyles =
    "text-gray-900 hover:text-gray-900 dark:text-gray-100 dark:hover:text-gray-100";

  return (
    <>
      <div className="bg-gray-50 dark:bg-surface-00dp">
        <div className="bg-white border-b border-gray-200 dark:bg-surface-01dp dark:border-neutral-700 mb-6">
          <nav className="container relative max-w-screen-lg mx-auto px-4 2xl:px-0 py-3">
            <Link
              to="/"
              className="z-10 absolute font-headline dark:text-white"
            >
              <img alt="NWC Logo" className="h-8 inline" src={nwcComboMark} />
            </Link>
            <div className="flex justify-center items-center relative h-8">
              <div className="flex space-x-4">
                <Link
                  className={`${linkStyles} ${
                    location.pathname.startsWith("/apps") && selectedLinkStyles
                  }`}
                  to="/apps"
                >
                  Apps
                </Link>
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
