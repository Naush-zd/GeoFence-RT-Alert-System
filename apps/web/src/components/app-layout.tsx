import { NavLink } from "react-router-dom"
import type { ReactNode } from "react"
import { AlertTriangle, BellRing, CircleGauge, MapPinned, Radio, Truck } from "lucide-react"
import { ModeToggle } from "./mode-toggle"
import { Badge } from "./ui/badge"
import { Button } from "./ui/button"
import { ScrollArea } from "./ui/scroll-area"
import { Separator } from "./ui/separator"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "./ui/sidebar"
import { pageDescriptions, type DashboardData } from "../lib/dashboard"
import { useLocation } from "react-router-dom"
import { LiveAlerts } from "./live-alerts"

export const navItems = [
  { to: "/", label: "Overview", icon: CircleGauge, end: true },
  { to: "/geofences", label: "Geofences", icon: MapPinned },
  { to: "/vehicles", label: "Vehicles", icon: Truck },
  { to: "/alerts", label: "Alert rules", icon: BellRing },
  { to: "/violations", label: "Violations", icon: AlertTriangle },
]

function PageHeader({ data, loading, onRefresh }: { data: DashboardData; loading: boolean; onRefresh: () => void }) {
  const location = useLocation()
  const current = navItems.find((item) => item.to === location.pathname)
  return (
    <header className="flex flex-col gap-6 border-b py-8 md:flex-row md:items-end md:justify-between">
      <div><p className="mb-3 text-xs text-muted-foreground">Workspace / {current?.label ?? "Overview"}</p><h1 className="text-3xl font-semibold tracking-tight">{current?.label ?? "Overview"}</h1><p className="mt-2 text-sm text-muted-foreground">{pageDescriptions[location.pathname] ?? pageDescriptions["/"]}</p></div>
      <div className="flex items-center gap-3"><Badge variant={data.apiStatus === "ok" ? "secondary" : data.apiStatus === "checking" ? "outline" : "destructive"}>API {data.apiStatus}</Badge><Button variant="outline" onClick={onRefresh} disabled={loading}>{loading ? "Refreshing…" : "Refresh"}</Button></div>
    </header>
  )
}

function AppLayout({ data, loading, error, onRefresh, children }: { data: DashboardData; loading: boolean; error: string | null; onRefresh: () => void; children: ReactNode }) {
  return (
    <><LiveAlerts /><SidebarProvider defaultOpen={false}>
      <Sidebar collapsible="icon" variant="sidebar">
        <SidebarHeader className="p-3"><div className="flex items-center gap-2 px-1"><div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground"><Radio size={18} /></div><div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden"><span className="truncate font-semibold text-sidebar-foreground">GeoFence</span><span className="truncate text-xs text-sidebar-foreground/60">Control center</span></div></div></SidebarHeader>
        <SidebarSeparator />
        <SidebarContent className="overflow-hidden"><ScrollArea className="h-full"><SidebarGroup><SidebarGroupLabel>Workspace</SidebarGroupLabel><SidebarGroupContent><SidebarMenu>{navItems.map(({ to, label, icon: Icon, end }) => <SidebarMenuItem key={to}><NavLink to={to} end={end}>{({ isActive }) => <SidebarMenuButton render={<span />} isActive={isActive} tooltip={label}><Icon /><span>{label}</span></SidebarMenuButton>}</NavLink></SidebarMenuItem>)}</SidebarMenu></SidebarGroupContent></SidebarGroup></ScrollArea></SidebarContent>
        <SidebarFooter className="gap-2 p-3"><Separator className="bg-sidebar-border" /><div className="flex items-center gap-2 text-xs"><span className={`size-2 rounded-full ${data.apiStatus === "ok" ? "bg-emerald-400" : data.apiStatus === "checking" ? "bg-amber-400" : "bg-red-400"}`} /><span className="truncate group-data-[collapsible=icon]:hidden">{data.apiStatus === "ok" ? "All systems operational" : "Check API connection"}</span></div><span className="text-[11px] text-sidebar-foreground/50 group-data-[collapsible=icon]:hidden">v0.1.0 · Local workspace</span></SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="flex h-14 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur md:px-6"><SidebarTrigger aria-label="Toggle navigation" /><Separator orientation="vertical" className="h-5" /><span className="text-sm font-semibold">GeoFence</span><span className="ml-auto flex items-center gap-3 text-xs text-muted-foreground"><span className="flex items-center gap-2"><span className={`size-2 rounded-full ${data.apiStatus === "ok" ? "bg-emerald-500" : "bg-amber-500"}`} /> API {data.apiStatus}</span><ModeToggle /></span></header>
        <main className="min-w-0 flex-1 px-4 pb-12 md:px-8 lg:px-12"><PageHeader data={data} loading={loading} onRefresh={onRefresh} />{error && <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive" role="alert">{error}</div>}{children}</main>
      </SidebarInset>
    </SidebarProvider></>
  )
}

export { AppLayout }
