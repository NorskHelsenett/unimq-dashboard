import { useCallback, useEffect, useRef } from "react"
import { useLocalStorage } from "./useLocalStorage"

const MIN_WIDTH = 70
const MAX_WIDTH = 250
const DEFAULT_WIDTH = 240

export function useSidebarResize(storageKey = "sidebar-width") {
    const [width, setWidth] = useLocalStorage<number>(storageKey, DEFAULT_WIDTH)
    const isDragging = useRef(false)

    const onMouseDown = useCallback((e: React.MouseEvent) => {
        e.preventDefault()
        isDragging.current = true
    }, [])

    useEffect(() => {
        const onMouseMove = (e: MouseEvent) => {
            if (!isDragging.current) return
            const clamped = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, e.clientX))
            setWidth(clamped)
        }

        const onMouseUp = () => {
            isDragging.current = false
        }

        window.addEventListener("mousemove", onMouseMove)
        window.addEventListener("mouseup", onMouseUp)
        return () => {
            window.removeEventListener("mousemove", onMouseMove)
            window.removeEventListener("mouseup", onMouseUp)
        }
    }, [setWidth])

    const collapsed = width < 90

    return { width, collapsed, onMouseDown }
}
