import React from "react";
import { useNavigate } from "react-router-dom";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";

import nwcComboMark from "src/assets/images/nwc-combomark.svg";
import Container from "src/components/Container";

export function SetupPassword() {
  const store = useSetupStore();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const navigate = useNavigate();
  const { data: info } = useInfo();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (store.unlockPassword !== confirmPassword) {
      alert("Passwords don't match!");
      return;
    }
    navigate("/setup/node");
  }

  return (
    <>
      <Container>
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <p className="text-center text-md leading-relaxed dark:text-neutral-400 px-4 mb-8">
            Choose a password to unlock the NWC app
          </p>
          {info?.setupCompleted && (
            <p className="mb-4 text-red-500 text-sm font-bold">
              Your node is already setup! only continue if you actually want to
              change your connection settings.
            </p>
          )}

          <div className="w-full mb-4">
            <label
              htmlFor="unlock-password"
              className="block mb-2 text-md dark:text-white"
            >
              New Password
            </label>
            <input
              type="password"
              name="unlock-password"
              id="unlock-password"
              placeholder="Enter a password"
              value={store.unlockPassword}
              onChange={(e) => store.setUnlockPassword(e.target.value)}
              className="bg-gray-50 border border-gray-300 text-gray-900 text-md rounded-md focus:ring-primary-600 focus:border-primary-600 block w-full px-3 py-2 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-purple-700 dark:focus:border-purple-700"
              required={true}
            />
          </div>
          <div className="w-full mb-8 ">
            <label
              htmlFor="confirm-password"
              className="block mb-2 text-md dark:text-white"
            >
              Confirm Password
            </label>
            <input
              type="password"
              name="confirm-password"
              id="confirm-password"
              placeholder="Re-enter the password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="bg-gray-50 border border-gray-300 text-gray-900 text-md rounded-md focus:ring-primary-600 focus:border-primary-600 block w-full px-3 py-2 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-purple-700 dark:focus:border-purple-700"
              required={true}
            />
          </div>

          <button className="flex-row w-full px-0 py-2 bg-purple-700 border-2 border-transparent text-white hover:bg-purple-800 cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-500 transition duration-150">
            Next
          </button>
        </form>
      </Container>
    </>
  );
}
