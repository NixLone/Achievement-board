export type Theme = "light-minimal" | "dark-ember" | "nord-calm" | "oled-black";

export const themes: { id: Theme; label: string }[] = [
  { id: "light-minimal", label: "Light Minimal" },
  { id: "dark-ember", label: "Dark Ember" },
  { id: "nord-calm", label: "Nord Calm" },
  { id: "oled-black", label: "OLED Black" }
];

const STORAGE_KEY = "firegoals-theme";

export function getStoredTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
  return stored ?? "dark-ember";
}

export function applyTheme(theme: Theme) {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem(STORAGE_KEY, theme);
}
