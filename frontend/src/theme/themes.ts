export type Theme = "light" | "ember" | "nord" | "oled";

export const themes: { id: Theme; label: string }[] = [
  { id: "light", label: "Light Minimal" },
  { id: "ember", label: "Dark Ember" },
  { id: "nord", label: "Nord Calm" },
  { id: "oled", label: "OLED Black" }
];

const STORAGE_KEY = "firegoals-theme";

export function getStoredTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
  return stored ?? "ember";
}

export function applyTheme(theme: Theme) {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem(STORAGE_KEY, theme);
}
