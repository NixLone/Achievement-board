export function mergeById<T extends { id: string; updated_at?: string; version?: number }>(
  existing: T[],
  incoming: T[]
): T[] {
  const map = new Map(existing.map((item) => [item.id, item]));
  incoming.forEach((item) => {
    const current = map.get(item.id);
    if (!current || isNewer(item, current)) {
      map.set(item.id, item);
    }
  });
  return Array.from(map.values());
}

function isNewer<T extends { updated_at?: string; version?: number }>(a: T, b: T): boolean {
  if (a.updated_at && b.updated_at) {
    const aTime = new Date(a.updated_at).getTime();
    const bTime = new Date(b.updated_at).getTime();
    if (aTime !== bTime) {
      return aTime > bTime;
    }
  }
  const aVersion = a.version ?? 0;
  const bVersion = b.version ?? 0;
  return aVersion >= bVersion;
}
