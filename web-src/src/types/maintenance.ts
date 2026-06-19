export interface MaintenanceEntry {
  id: string;
  description: string;
  start: string;
  end: string;
  status: "scheduled" | "done" | "skipped";
}
