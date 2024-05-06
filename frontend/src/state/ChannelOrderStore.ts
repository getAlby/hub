import { localStorageKeys } from "src/constants";
import { NewChannelOrder } from "src/types";
import { create } from "zustand";

interface ChannelOrderStore {
  readonly order: NewChannelOrder | undefined;
  setOrder(order: NewChannelOrder): void;
  removeOrder(): void;
  updateOrder(order: Partial<NewChannelOrder>): void;
}

const savedOrderJSON = localStorage.getItem(localStorageKeys.channelOrder);
const useChannelOrderStore = create<ChannelOrderStore>((set, get) => ({
  order: savedOrderJSON && JSON.parse(savedOrderJSON),
  removeOrder() {
    localStorage.removeItem(localStorageKeys.channelOrder);
    set({ order: undefined });
  },
  updateOrder: (order) => {
    get().setOrder({
      ...get().order,
      ...order,
    } as NewChannelOrder);
  },
  setOrder: (order) => {
    set({ order });
    localStorage.setItem(localStorageKeys.channelOrder, JSON.stringify(order));
  },
}));

export default useChannelOrderStore;
