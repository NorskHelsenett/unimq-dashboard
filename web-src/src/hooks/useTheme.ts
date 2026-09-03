import { useEffect } from "react"
import { useLocalStorage } from "./useLocalStorage"

export function useTheme() {
    const [theme, setTheme] = useLocalStorage<"light" | "dark">("theme", "light")

    useEffect(() => {
        const root = document.documentElement
        if (theme === "dark") {
            root.classList.add("dark")
        } else {
            root.classList.remove("dark")
        }
    }, [theme])

    const toggle = () => setTheme(t => (t === "dark" ? "light" : "dark"))

    return { theme, toggle }
}
