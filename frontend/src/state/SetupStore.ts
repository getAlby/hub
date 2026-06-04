import { SetupNodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: SetupNodeInfo;
  readonly unlockPassword: string;
  readonly hasImportedMnemonic: boolean;
  updateNodeInfo(nodeInfo: SetupNodeInfo): void;
  setUnlockPassword(unlockPassword: string): void;
  setHasImportedMnemonic(hasImportedMnemonic: boolean): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  nodeInfo: {},
  unlockPassword: "",
  hasImportedMnemonic: false,
  updateNodeInfo: (nodeInfo) =>
    set((state) => ({
      nodeInfo: { ...state.nodeInfo, ...nodeInfo },
    })),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
  setHasImportedMnemonic: (hasImportedMnemonic) => set({ hasImportedMnemonic }),
}));

export default useSetupStore;
