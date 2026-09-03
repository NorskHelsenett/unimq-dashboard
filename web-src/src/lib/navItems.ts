import { LayoutDashboard, Bell, Wrench } from "lucide-react"
import { NavItem } from "../types/navItem";

export const NAV_ITEMS: NavItem[] = [
  { label: "Overview", href: "/", icon: LayoutDashboard },
  { label: "Notifications", href: "/notifications", icon: Bell },
  { label: "Maintenance", href: "/maintenance", icon: Wrench },
];
