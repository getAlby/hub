import { NodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: NodeInfo;
  readonly unlockPassword: string;
  updateNodeInfo(nodeInfo: NodeInfo): void;
  setUnlockPassword(unlockPassword: string): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  nodeInfo: {},
  unlockPassword: "",
  updateNodeInfo: (nodeInfoPartial) =>
    set((state) => ({
      nodeInfo: state.nodeInfo
        ? { ...state.nodeInfo, ...nodeInfoPartial }
        : ({ ...nodeInfoPartial } as NodeInfo),
    })),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
}));

export default useSetupStore;
