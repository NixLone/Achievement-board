import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { WorkspaceSnapshot, emptySnapshot, loadSnapshot, saveSnapshot } from "../storage";

type StoreContextValue = {
  snapshot: WorkspaceSnapshot;
  setSnapshot: (next: WorkspaceSnapshot) => void;
};

const StoreContext = createContext<StoreContextValue | undefined>(undefined);

export function StoreProvider({ children }: { children: ReactNode }) {
  const [snapshot, setSnapshotState] = useState<WorkspaceSnapshot>(emptySnapshot());

  useEffect(() => {
    loadSnapshot().then((data) => setSnapshotState(data));
  }, []);

  const setSnapshot = (next: WorkspaceSnapshot) => {
    setSnapshotState(next);
    saveSnapshot(next).catch(() => undefined);
  };

  const value = useMemo(() => ({ snapshot, setSnapshot }), [snapshot]);

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore() {
  const ctx = useContext(StoreContext);
  if (!ctx) {
    throw new Error("StoreProvider is missing");
  }
  return ctx;
}
