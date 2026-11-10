import { NavLink } from "react-router-dom"
import type { ReactNode } from "react"
import { AlertTriangle, BellRing, CircleGauge, MapPinned, PanelLeftIcon, Radio, RefreshCw, Truck } from "lucide-react"
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
  useSidebar,
} from "./ui/sidebar"
import { pageDescriptions, type DashboardData } from "../lib/dashboard"
import type { AlertEvent } from "../lib/api"
import { useLocation } from "react-router-dom"
import { LiveAlerts } from "./live-alerts"

export const navItems = [
  { to: "/", label: "Overview", icon: CircleGauge, end: true },
  { to: "/geofences", label: "Geofences", icon: MapPinned },
  { to: "/vehicles", label: "Vehicles", icon: Truck },
  { to: "/alerts", label: "Alert rules", icon: BellRing },
  { to: "/violations", label: "Violations", icon: AlertTriangle },
]

function SidebarBrandToggle() {
  const { isMobile, openMobile, state, toggleSidebar } = useSidebar()
  const expanded = isMobile ? openMobile : state === "expanded"
  const label = isMobile
    ? expanded ? "Close navigation" : "Open navigation"
    : expanded ? "Collapse sidebar" : "Expand sidebar"

  return (
    <div className="flex w-full items-center gap-2 px-1 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="group relative rounded-lg border-0 bg-primary p-0 text-primary-foreground hover:bg-primary/90"
        onClick={toggleSidebar}
        aria-label={label}
        title={label}
      >
        <Radio className="size-[18px] transition-opacity group-hover:opacity-0 group-focus-visible:opacity-0" />
        <PanelLeftIcon className="absolute size-[18px] opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" />
      </Button>
      <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
        <span className="truncate font-semibold text-sidebar-foreground">GeoFence</span>
        <span className="truncate text-xs text-sidebar-foreground/60">Control center</span>
      </div>
    </div>
  )
}

function WorkspaceNavigation() {
  const { isMobile, setOpenMobile } = useSidebar()
  return <SidebarContent className="overflow-hidden"><ScrollArea className="h-full"><SidebarGroup><SidebarGroupLabel>Workspace</SidebarGroupLabel><SidebarGroupContent><SidebarMenu>{navItems.map(({ to, label, icon: Icon, end }) => <SidebarMenuItem key={to}><NavLink to={to} end={end} onClick={() => { if (isMobile) setOpenMobile(false) }}>{({ isActive }) => <SidebarMenuButton render={<span />} isActive={isActive} tooltip={label}><Icon /><span>{label}</span></SidebarMenuButton>}</NavLink></SidebarMenuItem>)}</SidebarMenu></SidebarGroupContent></SidebarGroup></ScrollArea></SidebarContent>
}

function PageHeader({ data, loading, onRefresh }: { data: DashboardData; loading: boolean; onRefresh: () => void }) {
  const location = useLocation()
  const current = navItems.find((item) => item.to === location.pathname)
  return (
    <header className="flex flex-col gap-4 border-b py-6 sm:flex-row sm:items-end sm:justify-between sm:py-7">
      <div className="min-w-0"><p className="mb-2 text-xs text-muted-foreground">Workspace / {current?.label ?? "Overview"}</p><h1 className="text-2xl font-semibold sm:text-3xl">{current?.label ?? "Overview"}</h1><p className="mt-2 text-sm text-muted-foreground">{pageDescriptions[location.pathname] ?? pageDescriptions["/"]}</p></div>
      <div className="flex shrink-0 items-center gap-3"><Badge variant={data.apiStatus === "ok" ? "secondary" : data.apiStatus === "checking" ? "outline" : "destructive"}>API {data.apiStatus}</Badge><Button variant="outline" onClick={onRefresh} disabled={loading}><RefreshCw data-icon="inline-start" className={loading ? "animate-spin" : undefined} />{loading ? "Refreshing…" : "Refresh"}</Button></div>
    </header>
  )
}

function AppLayout({ data, loading, error, onRefresh, onAlertEvent, children }: { data: DashboardData; loading: boolean; error: string | null; onRefresh: () => void; onAlertEvent: (event: AlertEvent) => void; children: ReactNode }) {
  const defaultSidebarOpen = typeof window !== "undefined" && window.innerWidth >= 768
  return (
    <><LiveAlerts onEvent={onAlertEvent} /><SidebarProvider defaultOpen={defaultSidebarOpen}>
      <Sidebar collapsible="icon" variant="sidebar">
        <SidebarHeader className="p-3"><SidebarBrandToggle /></SidebarHeader>
        <SidebarSeparator />
        <WorkspaceNavigation />
        <SidebarFooter className="gap-2 p-3 group-data-[collapsible=icon]:p-2">
          <Separator className="bg-sidebar-border" />
          <div className="flex items-center justify-between gap-2 group-data-[collapsible=icon]:justify-center">
            <span className="text-xs text-sidebar-foreground/70 group-data-[collapsible=icon]:hidden">Appearance</span>
            <ModeToggle />
          </div>
          <span className="text-[11px] text-sidebar-foreground/50 group-data-[collapsible=icon]:hidden">v0.1.0 · Local workspace</span>
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="sticky top-0 z-40 flex h-14 shrink-0 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur md:hidden"><SidebarTrigger aria-label="Toggle navigation" /><Separator orientation="vertical" className="h-5" /><span className="text-sm font-semibold">GeoFence</span></header>
        <main className="min-w-0 flex-1 px-4 pb-12 md:px-8 lg:px-12"><PageHeader data={data} loading={loading} onRefresh={onRefresh} />{error && <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive" role="alert">{error}</div>}{children}</main>
      </SidebarInset>
    </SidebarProvider></>
  )
}

export { AppLayout }
