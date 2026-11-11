import { NavLink } from "react-router-dom"
import { AlertTriangle, BellRing, MapPinned, Trash2, Truck } from "lucide-react"
import type { ReactNode } from "react"
import type { AlertConfig, Geofence, Vehicle, Violation } from "../types/api"
import { Badge } from "./ui/badge"
import { Button, buttonVariants } from "./ui/button"
import { Card, CardContent, CardHeader, CardDescription, CardTitle } from "./ui/card"

function EmptyState({ text }: { text: string }) {
	return <p className="py-6 text-center text-sm text-muted-foreground">{text}</p>
}

function ListCard({ children }: { children: ReactNode }) {
	return (
		<Card>
			<CardContent className="p-2">{children}</CardContent>
		</Card>
	)
}

function PageSection({
	title,
	count,
	action,
	actionContent,
	children,
}: {
	title: string
	count: string
	action: string
	actionContent?: ReactNode
	children: ReactNode
}) {
	return (
		<section className="flex flex-col gap-4 py-5">
			<div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
				<div className="min-w-0">
					<h2 className="text-lg font-semibold">{title}</h2>
					<p className="text-sm text-muted-foreground">{count}</p>
				</div>
				<div className="flex w-full flex-col gap-2 sm:w-auto sm:items-end **:data-[slot=button]:w-full sm:**:data-[slot=button]:w-auto">
					{actionContent ?? <Button className="w-full sm:w-auto">{action}</Button>}
				</div>
			</div>
			{children}
		</section>
	)
}

function MetricCard({ label, value, detail, icon }: { label: string; value: number; detail: string; icon: ReactNode }) {
	return (
		<Card>
			<CardContent className="p-4 sm:p-5">
				<div className="mb-3 flex size-8 items-center justify-center rounded-md bg-primary/10 text-primary">{icon}</div>
				<p className="text-sm text-muted-foreground">{label}</p>
				<p className="mt-1 text-2xl font-semibold tabular-nums sm:text-3xl">{value}</p>
				<p className="mt-1 text-xs text-muted-foreground">{detail}</p>
			</CardContent>
		</Card>
	)
}

function Panel({ title, description, link, children }: { title: string; description: string; link: string; children: ReactNode }) {
	return (
		<Card>
			<CardHeader className="flex flex-row items-start justify-between gap-4">
				<div>
					<CardTitle>{title}</CardTitle>
					<CardDescription>{description}</CardDescription>
				</div>
				<NavLink className={buttonVariants({ variant: "ghost", size: "sm" })} to={link}>
					View all
				</NavLink>
			</CardHeader>
			<CardContent>{children}</CardContent>
		</Card>
	)
}

function formatDate(value: string) {
	return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value))
}

function GeofenceList({
	items,
	selectedId,
	onSelect,
	onDelete,
}: {
	items: Geofence[]
	selectedId?: string | null
	onSelect?: (id: string) => void
	onDelete?: (geofence: Geofence) => void
}) {
	if (items.length === 0) return <EmptyState text="No geofences configured yet." />

	return (
		<div className="grid gap-1.5 overflow-hidden">
			{items.map((item) => {
				const selected = item.id === selectedId
				return (
					<div
						className={`grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2 rounded-md p-1.5 transition-colors ${selected ? "bg-primary/10 ring-1 ring-primary/50" : "bg-muted/50"}`}
						key={item.id}
					>
						<Button
							type="button"
							variant="ghost"
							className="h-auto min-h-12 w-full min-w-0 justify-start gap-3 overflow-hidden border-0 px-2 py-2 text-left"
							aria-pressed={selected}
							onClick={() => onSelect?.(item.id)}
						>
							<span className={`flex size-8 shrink-0 items-center justify-center rounded-md ${selected ? "bg-primary text-primary-foreground" : "bg-primary/15 text-primary"}`}>
								<MapPinned />
							</span>
							<span className="min-w-0 flex-1 overflow-hidden">
								<span className="block truncate text-sm font-medium" title={item.name}>{item.name}</span>
								<span className="block truncate text-xs text-muted-foreground" title={item.description || "No description"}>
									{item.description || "No description"}
								</span>
							</span>
						</Button>
						<div className="flex shrink-0 items-center gap-1">
							<Badge variant="secondary" className="hidden md:inline-flex">{item.category.replaceAll("_", " ")}</Badge>
							{onDelete && (
								<Button type="button" variant="ghost" size="icon" aria-label={`Remove ${item.name}`} onClick={() => onDelete(item)}>
									<Trash2 />
								</Button>
							)}
						</div>
					</div>
				)
			})}
		</div>
	)
}

function VehicleList({ items }: { items: Vehicle[] }) {
	if (items.length === 0) return <EmptyState text="No vehicles registered yet." />

	return (
		<div className="grid gap-2">
			{items.map((item) => (
				<div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}>
					<div className="flex min-w-0 items-center gap-3">
						<span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary">
							<Truck size={16} />
						</span>
						<div className="min-w-0">
							<p className="truncate text-sm font-medium">{item.vehicle_number}</p>
							<p className="truncate text-xs text-muted-foreground">{item.driver_name} · {item.vehicle_type}</p>
						</div>
					</div>
					<Badge>{item.status}</Badge>
				</div>
			))}
		</div>
	)
}

function AlertList({ items }: { items: AlertConfig[] }) {
	if (items.length === 0) return <EmptyState text="No alert rules configured." />

	return (
		<div className="grid gap-2">
			{items.map((item) => (
				<div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.alert_id}>
					<div className="flex min-w-0 items-center gap-3">
						<span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary">
							<BellRing size={16} />
						</span>
						<div className="min-w-0">
							<p className="truncate text-sm font-medium">{item.geofence_name}</p>
							<p className="truncate text-xs text-muted-foreground">{item.vehicle_number || "All vehicles"} · {item.event_type}</p>
						</div>
					</div>
					<Badge>{item.status}</Badge>
				</div>
			))}
		</div>
	)
}

function ViolationList({ items }: { items: Violation[] }) {
	if (items.length === 0) return <EmptyState text="No violation events recorded." />

	return (
		<div className="grid gap-2">
			{items.map((item) => (
				<div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}>
					<div className="flex min-w-0 items-center gap-3">
						<span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-destructive/15 text-destructive">
							<AlertTriangle size={16} />
						</span>
						<div className="min-w-0">
							<p className="truncate text-sm font-medium">{item.geofence_name}</p>
							<p className="truncate text-xs text-muted-foreground">{item.vehicle_number} · {formatDate(item.timestamp)}</p>
						</div>
					</div>
					<Badge variant="destructive">{item.event_type}</Badge>
				</div>
			))}
		</div>
	)
}

export { AlertList, GeofenceList, ListCard, MetricCard, PageSection, Panel, VehicleList, ViolationList }
