import { useState } from "react"
import type { Geofence } from "../types/api"
import { api as apiClient } from "../lib/api"
import { Button } from "./ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "./ui/dialog"

export function DeleteGeofenceDialog({ target, onOpenChange, onDeleted }: { target: Geofence | null; onOpenChange: (open: boolean) => void; onDeleted: () => void }) {
  const [error, setError] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)
  const remove = async () => {
    if (!target) return
    setDeleting(true); setError(null)
    try { await apiClient.deleteGeofence(target.id); onDeleted() }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to remove geofence") }
    finally { setDeleting(false) }
  }
  return <Dialog open={Boolean(target)} onOpenChange={onOpenChange}><DialogContent className="sm:max-w-md"><DialogHeader><DialogTitle>Remove geofence?</DialogTitle><DialogDescription>This removes {target?.name ?? "this boundary"} and its related alert configurations and violation history.</DialogDescription></DialogHeader>{error && <p className="text-sm text-destructive" role="alert">{error}</p>}<DialogFooter><Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button><Button type="button" variant="destructive" onClick={() => void remove()} disabled={deleting}>{deleting ? "Removing…" : "Remove geofence"}</Button></DialogFooter></DialogContent></Dialog>
}
