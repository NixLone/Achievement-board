import { useEffect, useState } from "react";
import { Theme, applyTheme, getStoredTheme } from "./themes";

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(getStoredTheme());

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  return { theme, setTheme };
}
