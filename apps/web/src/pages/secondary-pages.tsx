import type { AlertConfig, Violation } from "../lib/api"
import { AlertList, ListCard, PageSection, ViolationList } from "../components/dashboard-widgets"

export function AlertsPage({ alerts }: { alerts: AlertConfig[] }) { return <PageSection title="Alert rules" count={`${alerts.length} rules`} action="+ Configure alert"><ListCard><AlertList items={alerts} /></ListCard></PageSection> }
export function ViolationsPage({ violations, total }: { violations: Violation[]; total: number }) { return <PageSection title="Violation history" count={`${total} total events`} action="Export history"><ListCard><ViolationList items={violations} /></ListCard></PageSection> }
