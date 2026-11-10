import { useEffect, useState } from "react"
import type { Geofence } from "../lib/api"
import { api as apiClient } from "../lib/api"
import { OperationsMap } from "../components/operations-map"
import { Button } from "../components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "../components/ui/dialog"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"
import { DeleteGeofenceDialog } from "../components/geofence-dialogs"
import { GeofenceList, PageSection } from "../components/dashboard-widgets"

export function GeofencesPage({ geofences, onCreated }: { geofences: Geofence[]; onCreated: () => void }) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draftCoordinates, setDraftCoordinates] = useState<number[][]>([])
  const [selectedGeofenceId, setSelectedGeofenceId] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Geofence | null>(null)
  const [category, setCategory] = useState("all")
  useEffect(() => { if (selectedGeofenceId && !geofences.some((geofence) => geofence.id === selectedGeofenceId)) setSelectedGeofenceId(null) }, [geofences, selectedGeofenceId])
  const filteredGeofences = category === "all" ? geofences : geofences.filter((geofence) => geofence.category === category)
  return <PageSection title="Geofence registry" count={`${filteredGeofences.length} of ${geofences.length} zones`} action="+ New geofence" actionContent={<><div className="flex items-center gap-2"><Select value={category} onValueChange={(value) => setCategory(value ?? "all")}><SelectTrigger className="w-40"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="all">All categories</SelectItem><SelectItem value="delivery_zone">Delivery zones</SelectItem><SelectItem value="restricted_zone">Restricted zones</SelectItem><SelectItem value="toll_zone">Toll zones</SelectItem><SelectItem value="customer_area">Customer areas</SelectItem></SelectContent></Select><CreateGeofenceDialog open={dialogOpen} onOpenChange={setDialogOpen} coordinates={draftCoordinates} onCoordinatesChange={setDraftCoordinates} onCreated={onCreated} /></div><DeleteGeofenceDialog target={deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }} onDeleted={() => { setDeleteTarget(null); setSelectedGeofenceId(null); onCreated() }} /></>}><div className="grid gap-4 xl:grid-cols-[minmax(0,1.3fr)_minmax(20rem,0.7fr)]"><Card className="overflow-hidden"><CardHeader><CardTitle>Live boundary map</CardTitle><CardDescription>Select a boundary to focus and highlight it on the map.</CardDescription></CardHeader><CardContent className="p-0"><div className="h-[28rem]"><OperationsMap geofences={filteredGeofences} selectedGeofenceId={selectedGeofenceId} /></div></CardContent></Card><Card><CardHeader><CardTitle>Configured boundaries</CardTitle><CardDescription>Choose a geofence to inspect or remove it from monitoring.</CardDescription></CardHeader><CardContent className="pt-0"><GeofenceList items={filteredGeofences} selectedId={selectedGeofenceId} onSelect={setSelectedGeofenceId} onDelete={setDeleteTarget} /></CardContent></Card></div></PageSection>
}

function CreateGeofenceDialog({ open, onOpenChange, coordinates, onCoordinatesChange, onCreated }: { open: boolean; onOpenChange: (open: boolean) => void; coordinates: number[][]; onCoordinatesChange: (coordinates: number[][]) => void; onCreated: () => void }) {
  const [name, setName] = useState(""); const [description, setDescription] = useState(""); const [category, setCategory] = useState("delivery_zone"); const [error, setError] = useState<string | null>(null); const [submitting, setSubmitting] = useState(false)
  const reset = () => { setName(""); setDescription(""); setCategory("delivery_zone"); setError(null); onCoordinatesChange([]) }
  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); if (coordinates.length < 3) { setError("Add at least three points on the map."); return }
    setSubmitting(true); setError(null)
    try { await apiClient.createGeofence({ name, description, category, coordinates: [...coordinates, coordinates[0]] }); onCreated(); reset(); onOpenChange(false) }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to create geofence") }
    finally { setSubmitting(false) }
  }
  return <Dialog open={open} onOpenChange={(nextOpen) => { onOpenChange(nextOpen); if (!nextOpen) reset() }}><DialogTrigger render={<Button />}>+ New geofence</DialogTrigger><DialogContent className="max-h-[92vh] overflow-y-auto sm:max-w-4xl"><DialogHeader><DialogTitle>Create geofence</DialogTitle><DialogDescription>Click three or more points, then drag the handles to resize the polygon.</DialogDescription></DialogHeader><form className="flex flex-col gap-4" onSubmit={(event) => void submit(event)}><div className="overflow-hidden rounded-lg border"><div className="h-72"><OperationsMap geofences={[]} draftCoordinates={coordinates} onAddPoint={(point) => onCoordinatesChange([...coordinates, point])} onMovePoint={(index, point) => onCoordinatesChange(coordinates.map((currentPoint, currentIndex) => currentIndex === index ? point : currentPoint))} /></div></div><div className="flex items-center justify-between gap-3"><p className="text-xs text-muted-foreground">{coordinates.length} points selected. A closing point will be added automatically.</p><Button type="button" variant="outline" size="sm" onClick={() => onCoordinatesChange([])} disabled={coordinates.length === 0}>Reset points</Button></div><div className="flex flex-col gap-2"><Label htmlFor="geofence-name">Name</Label><Input id="geofence-name" value={name} onChange={(event) => setName(event.target.value)} required placeholder="Downtown delivery zone" /></div><div className="flex flex-col gap-2"><Label htmlFor="geofence-description">Description</Label><Input id="geofence-description" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Optional context" /></div><div className="flex flex-col gap-2"><Label>Category</Label><Select value={category} onValueChange={(value) => setCategory(value ?? "delivery_zone")}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="delivery_zone">Delivery zone</SelectItem><SelectItem value="restricted_zone">Restricted zone</SelectItem><SelectItem value="toll_zone">Toll zone</SelectItem><SelectItem value="customer_area">Customer area</SelectItem></SelectContent></Select></div>{error && <p className="text-sm text-destructive" role="alert">{error}</p>}<DialogFooter><Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button><Button type="submit" disabled={submitting}>{submitting ? "Creating…" : "Create geofence"}</Button></DialogFooter></form></DialogContent></Dialog>
}
