import { NodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: NodeInfo | null;
  readonly unlockPassword: string;
  setNodeInfo(nodeInfo: NodeInfo): void;
  setUnlockPassword(unlockPassword: string): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  nodeInfo: null,
  unlockPassword: "",
  setNodeInfo: (nodeInfo) => set({ nodeInfo }),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
}));

export default useSetupStore;
