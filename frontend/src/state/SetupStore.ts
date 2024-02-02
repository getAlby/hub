import { create } from "zustand";

interface SetupStore {
  readonly unlockPassword: string;
  setUnlockPassword(unlockPassword: string): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  unlockPassword: "",
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
}));

export default useSetupStore;
