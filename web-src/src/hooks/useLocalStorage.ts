import { useState } from "react"

export function useLocalStorage<T>(key: string, initialValue: T) {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = localStorage.getItem(key)
            return item !== null ? (JSON.parse(item) as T) : initialValue
        } catch {
            return initialValue
        }
    })

    const setValue = (value: T | ((prev: T) => T)) => {
        setStoredValue(prev => {
            const next = typeof value === "function" ? (value as (prev: T) => T)(prev) : value
            try {
                localStorage.setItem(key, JSON.stringify(next))
            } catch {}
            return next
        })
    }

    return [storedValue, setValue] as const
}
