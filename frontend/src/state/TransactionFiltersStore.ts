import {
  defaultTransactionFilters,
  type TransactionFilters,
} from "src/hooks/useTransactions";
import { create } from "zustand";

interface TransactionFiltersStore {
  readonly filters: TransactionFilters;
  setFilters(filters: TransactionFilters): void;
}

const useTransactionFiltersStore = create<TransactionFiltersStore>((set) => ({
  filters: defaultTransactionFilters,
  setFilters: (filters) => set({ filters }),
}));

export default useTransactionFiltersStore;
