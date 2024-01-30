import React from "react";
import { useNavigate } from "react-router-dom";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";

export function SetupPassword() {
  const store = useSetupStore();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const navigate = useNavigate();
  const { data: info } = useInfo();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (store.unlockPassword !== confirmPassword) {
      alert("Password doesn't match!");
      return;
    }
    navigate("/setup/node");
  }

  return (
    <>
      {info?.setupCompleted && (
        <p className="mb-4 text-red-500">
          Your node is already setup! only continue if you actually want to
          change your connection settings.
        </p>
      )}
      <form onSubmit={onSubmit}>
        <div>
          <label
            htmlFor="unlock-password"
            className="block font-medium text-gray-900 dark:text-white"
          >
            Choose an unlock password
          </label>
          <p className="text-sm">
            This will encrypt your node information for next time you start NWC.
            You'll also be asked to confirm your password when you setup a new
            app connection.
          </p>
          <input
            name="unlock-password"
            onChange={(e) => store.setUnlockPassword(e.target.value)}
            value={store.unlockPassword}
            type="password"
            id="unlock-password"
            className="mb-4 dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
          />
          <label
            htmlFor="confirm-password"
            className="block font-medium text-gray-900 dark:text-white"
          >
            Confirm password
          </label>
          <input
            name="confirm-password"
            onChange={(e) => setConfirmPassword(e.target.value)}
            value={confirmPassword}
            type="password"
            id="confirm-password"
            className="mb-4 dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
          />
        </div>
        <button>Submit</button>
      </form>
    </>
  );
}
