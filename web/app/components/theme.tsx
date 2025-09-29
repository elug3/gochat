import { createContext, useContext, useEffect, useState, type ReactNode } from "react";

type Theme = 'light' | 'dark';


type ThemeContextType = {
    theme: Theme;
    toggleTheme: () => void;
};

const ThemeContext = createContext<ThemeContextType | null>(null);

export function ThemeProvider({children}: {children: ReactNode}) {
    const [theme, setTheme] = useState<Theme>('light');


    useEffect(() => {
        const saved = localStorage.getItem("theme") as Theme | null;
        if (saved) {
          setTheme(saved);
        } else {
          const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
          console.log("prefers dark:", prefersDark);
          setTheme(prefersDark ? "dark" : "light");
        }
      }, []);
    
      useEffect(() => {
        if (theme === "dark") {
          document.documentElement.classList.add("dark");
        } else {
          document.documentElement.classList.remove("dark");
        }
        localStorage.setItem("theme", theme);
      }, [theme]);
    

    const toggleTheme = () => {
        setTheme((prev) => {
            return prev === 'light' ? 'dark' : 'light';
        }); 
    }
    return (
        <ThemeContext.Provider value={{theme, toggleTheme}}>
            {children}
        </ThemeContext.Provider>
    )
}

export function useTheme(): ThemeContextType {
    const context = useContext(ThemeContext);
    if (context == null) {
        throw new Error("useTheme must be used within a ThemeProvider");
    }
    return context;
}
