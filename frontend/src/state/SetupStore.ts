import { SetupNodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: SetupNodeInfo;
  readonly unlockPassword: string;
  updateNodeInfo(nodeInfo: SetupNodeInfo): void;
  setUnlockPassword(unlockPassword: string): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  nodeInfo: {},
  unlockPassword: "",
  updateNodeInfo: (nodeInfo) =>
    set((state) => ({
      nodeInfo: { ...state.nodeInfo, ...nodeInfo },
    })),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
}));

export default useSetupStore;
