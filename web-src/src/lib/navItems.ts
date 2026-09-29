import { LayoutDashboard, Bell, Wrench, ShieldUser } from "lucide-react"
import { NavItem } from "../types/navItem";

export const NAV_ITEMS: NavItem[] = [
  { label: "Overview", href: "/", icon: LayoutDashboard },
  { label: "Notifications", href: "/notifications", icon: Bell },
  { label: "Maintenance", href: "/maintenance", icon: Wrench },
  //{ label: "Access control", href: "/access-control", icon: Shield }
]

export const NAV_ITEMS_ADMIN: NavItem[] = [
  { label: "Access control", href: "/access-control", icon: ShieldUser }
]
