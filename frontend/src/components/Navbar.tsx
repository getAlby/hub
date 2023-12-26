import { Outlet } from "react-router-dom";
import { useInfo } from "../hooks/useInfo";
import { logout } from "../utils/logout";
import { useLogin } from "../hooks/useLogin";
import nwcLogo from "../assets/images/nwc-logo.svg";
import caretIcon from "../assets/icons/caret.svg";
import aboutIcon from "../assets/icons/about.svg";
import React from "react";
import { LogoutIcon } from "./icons/LogoutIcon";

function Navbar() {
  const { data: info } = useInfo();
  return (
    <>
      <div className="bg-gray-50 dark:bg-surface-00dp">
        <div className="bg-white border-b border-gray-200 dark:bg-surface-01dp dark:border-neutral-700 mb-6">
          <nav className="container max-w-screen-lg mx-auto px-4 lg:px-0 py-3">
            <div className="flex justify-between items-center">
              <div className="flex items-center space-x-12">
                <a
                  href="/"
                  className="font-headline text-[20px] dark:text-white"
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
                  className={`${
                    info?.user ? "hidden md:flex" : "flex"
                  } space-x-4`}
                >
                  <a
                    className="text-gray-400 font-medium hover:text-gray-600 dark:hover:text-gray-300 transition"
                    href="/apps"
                  >
                    Connections
                  </a>
                  <a className="text-gray-400 pl-5 font-medium" href="/about">
                    About
                  </a>
                </div>
              </div>
              {info?.backendType === "ALBY" && <ProfileDropdown />}
            </div>
          </nav>
        </div>
      </div>
      <Outlet />
    </>
  );
}

function ProfileDropdown() {
  useLogin();
  const { data: info } = useInfo();
  const [isOpen, setOpen] = React.useState(false);

  if (!info?.user) {
    return null;
  }

  // TODO: add a proper dropdown component
  return (
    <div className="flex items-center relative">
      <p
        className="text-gray-400 text-xs font-medium sm:text-base cursor-pointer select-none "
        onClick={() => setOpen((current) => !current)}
      >
        <span>{info.user.email}</span>
        <img
          id="caret"
          className="inline cursor-pointer w-4 ml-2"
          src={caretIcon}
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
            <img
              className="inline cursor-pointer w-4 mr-3"
              src={aboutIcon}
              alt="about-svg"
            />
            <p className="font-normal">About</p>
          </a>

          {info?.user && (
            <div
              className="flex items-center justify-left py-2 text-red-500"
              onClick={logout}
            >
              <LogoutIcon
                className="inline cursor-pointer w-4 mr-3"
                alt="logout-svg"
              />
              <p className="font-normal">Logout</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Navbar;
