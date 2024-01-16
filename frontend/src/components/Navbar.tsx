import React from "react";
import { Outlet, useLocation } from "react-router-dom";

import { useInfo } from "src/hooks/useInfo";
import { useLogin } from "src/hooks/useLogin";
import { useCSRF } from "src/hooks/useCSRF";
import { useUser } from "src/hooks/useUser";
import { LogoutIcon } from "src/components/icons/LogoutIcon";
import { AboutIcon } from "src/components/icons/AboutIcon";
import { CaretIcon } from "src/components/icons/CaretIcon";
import nwcLogo from "src/assets/images/nwc-logo.svg";
import { handleFetchError } from "src/utils/request";

function Navbar() {
  const { data: info } = useInfo();
  const { data: user } = useUser();
  const location = useLocation();

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
                <a
                  href="/"
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
                </a>

                <div
                  className={`${user ? "hidden md:flex" : "flex"} space-x-4`}
                >
                  <a
                    className={`${linkStyles} ${
                      location.pathname.startsWith("/apps") &&
                      selectedLinkStyles
                    }`}
                    href="/apps"
                  >
                    Connections
                  </a>
                  <a
                    className={`${linkStyles} ${
                      location.pathname === "/setup" && selectedLinkStyles
                    }`}
                    href="/setup"
                  >
                    Setup
                  </a>
                  <a
                    className={`${linkStyles} ${
                      location.pathname === "/about" && selectedLinkStyles
                    }`}
                    href="/about"
                  >
                    About
                  </a>
                </div>
              </div>
              {info?.backendType === "ALBY" && <ProfileDropdown />}
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

function ProfileDropdown() {
  useLogin();
  const { data: csrf } = useCSRF();
  const { data: user } = useUser();
  const [isOpen, setOpen] = React.useState(false);

  if (!user) {
    return null;
  }

  async function logout() {
    try {
      if (!csrf) {
        throw new Error("info not loaded");
      }
      await fetch("/api/logout", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
        },
      });
      window.location.href = "/";
    } catch (error) {
      handleFetchError("Failed to logout", error);
    }
  }

  // TODO: add a proper dropdown component
  return (
    <div className="flex items-center relative">
      <p
        className="text-gray-400 text-xs font-medium sm:text-base cursor-pointer select-none"
        onClick={() => setOpen((current) => !current)}
      >
        <span className="mx-1">{user.email}</span>
        <CaretIcon
          className={`inline cursor-pointer w-4 ml-2 ${
            isOpen ? "rotate-180" : ""
          }`}
        />
      </p>

      {isOpen && (
        <div
          className="font-medium flex flex-col px-4 w-40 logout absolute right top-8 right-0 justify-left cursor-pointer rounded-lg border border-gray-200 dark:border-gray-600 text-center bg-white dark:bg-surface-01dp shadow"
          id="dropdown"
        >
          <a
            className="md:hidden flex items-center justify-left  py-2 text-gray-400 dark:text-gray-400"
            href="/about"
          >
            <AboutIcon className="inline cursor-pointer w-4 mr-3" />
            <p className="font-normal">About</p>
          </a>

          <div
            className="flex items-center justify-left py-2 text-red-500"
            onClick={logout}
          >
            <LogoutIcon className="inline cursor-pointer w-4 mr-3" />
            <p className="font-normal">Logout</p>
          </div>
        </div>
      )}
    </div>
  );
}

export default Navbar;
