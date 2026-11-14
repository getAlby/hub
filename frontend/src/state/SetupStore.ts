import { SetupNodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: SetupNodeInfo;
  readonly unlockPassword: string;
  readonly hasImportedMnemonic: boolean;
  readonly setupCode: string;
  updateNodeInfo(nodeInfo: SetupNodeInfo): void;
  setUnlockPassword(unlockPassword: string): void;
  setHasImportedMnemonic(hasImportedMnemonic: boolean): void;
  setSetupCode(setupCode: string): void;
  reset(): void;
}

const initialState = {
  nodeInfo: {},
  unlockPassword: "",
  hasImportedMnemonic: false,
  setupCode: "",
};

const useSetupStore = create<SetupStore>((set) => ({
  ...initialState,
  updateNodeInfo: (nodeInfo) =>
    set((state) => ({
      nodeInfo: { ...state.nodeInfo, ...nodeInfo },
    })),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
  setHasImportedMnemonic: (hasImportedMnemonic) => set({ hasImportedMnemonic }),
  setSetupCode: (setupCode) => set({ setupCode }),
  reset: () => set(initialState),
}));

export default useSetupStore;
