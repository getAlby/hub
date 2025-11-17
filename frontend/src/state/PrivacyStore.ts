import { localStorageKeys } from "src/constants";
import { create } from "zustand";

interface PrivacyStore {
  readonly privacyMode: boolean;
  setPrivacyMode(enabled: boolean): void;
}

const savedPrivacyMode = localStorage.getItem(localStorageKeys.privacyMode);
const usePrivacyStore = create<PrivacyStore>((set) => ({
  privacyMode: savedPrivacyMode === "true",
  setPrivacyMode: (enabled) => {
    set({ privacyMode: enabled });
    localStorage.setItem(localStorageKeys.privacyMode, String(enabled));

    // Apply or remove the privacy-mode class to the body
    if (enabled) {
      document.body.classList.add("privacy-mode");
    } else {
      document.body.classList.remove("privacy-mode");
    }
  },
}));

export default usePrivacyStore;
