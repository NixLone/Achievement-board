const tabs = [
  { id: "home", label: "Home", icon: "🏠" },
  { id: "plans", label: "Plans", icon: "🗓️" },
  { id: "store", label: "Store", icon: "🎁" },
  { id: "achievements", label: "Awards", icon: "🏅" },
  { id: "workspace", label: "Workspace", icon: "👥" },
  { id: "settings", label: "Settings", icon: "⚙️" }
] as const;

export type TabId = (typeof tabs)[number]["id"];

export function BottomNav({ active, onChange }: { active: TabId; onChange: (tab: TabId) => void }) {
  return (
    <nav className="bottom-nav">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={active === tab.id ? "active" : ""}
          onClick={() => onChange(tab.id)}
        >
          <span>{tab.icon}</span>
          <small>{tab.label}</small>
        </button>
      ))}
    </nav>
  );
}
