import { NodeInfo } from "src/types";
import { create } from "zustand";

interface SetupStore {
  readonly nodeInfo: NodeInfo | null;
  readonly unlockPassword: string;
  setNodeInfo(nodeInfo: Partial<NodeInfo>): void;
  setUnlockPassword(unlockPassword: string): void;
}

const useSetupStore = create<SetupStore>((set) => ({
  nodeInfo: null,
  unlockPassword: "",
  setNodeInfo: (nodeInfoPartial) =>
    set((state) => ({
      nodeInfo: state.nodeInfo
        ? { ...state.nodeInfo, ...nodeInfoPartial }
        : ({ ...nodeInfoPartial } as NodeInfo),
    })),
  setUnlockPassword: (unlockPassword) => set({ unlockPassword }),
}));

export default useSetupStore;
