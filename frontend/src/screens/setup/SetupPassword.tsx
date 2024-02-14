import React from "react";
import { useNavigate } from "react-router-dom";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";

import Container from "src/components/Container";
import Alert from "src/components/Alert";
import toast from "src/components/Toast";
import Input from "src/components/Input";
import PasswordViewAdornment from "src/components/PasswordAdornment";

export function SetupPassword() {
  const store = useSetupStore();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [passwordVisible, setPasswordVisible] = React.useState(false);
  const [confirmPasswordVisible, setConfirmPasswordVisible] =
    React.useState(false);
  const navigate = useNavigate();
  const { data: info } = useInfo();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (store.unlockPassword !== confirmPassword) {
      toast.error("Passwords don't match!");
      return;
    }
    navigate("/setup/wallet");
  }

  return (
    <>
      <Container>
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <h1 className="font-semibold text-2xl font-headline mb-2 dark:text-white">
            Choose an unlock password
          </h1>
          <p className="text-center font-light text-md leading-relaxed dark:text-neutral-400 mb-4">
            Your unlock password will be required to access NWC from a different
            device or browser session.
          </p>
          {info?.setupCompleted && (
            <Alert type="warn">
              ⚠️ Your node is already setup! only continue if you actually want
              to change your connection settings.
            </Alert>
          )}
          <div className="w-full my-4">
            <label
              htmlFor="unlock-password"
              className="block mb-2 text-md dark:text-white"
            >
              New Password
            </label>
            <Input
              type={passwordVisible ? "text" : "password"}
              name="unlock-password"
              id="unlock-password"
              placeholder="Enter a password"
              value={store.unlockPassword}
              onChange={(e) => store.setUnlockPassword(e.target.value)}
              required={true}
              endAdornment={
                <PasswordViewAdornment
                  onChange={(passwordView) => {
                    setPasswordVisible(passwordView);
                  }}
                />
              }
            />
          </div>
          <div className="w-full mb-8 ">
            <label
              htmlFor="confirm-password"
              className="block mb-2 text-md dark:text-white"
            >
              Confirm Password
            </label>
            <Input
              type={confirmPasswordVisible ? "text" : "password"}
              name="confirm-password"
              id="confirm-password"
              placeholder="Re-enter the password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required={true}
              endAdornment={
                <PasswordViewAdornment
                  onChange={(passwordView) => {
                    setConfirmPasswordVisible(passwordView);
                  }}
                />
              }
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
