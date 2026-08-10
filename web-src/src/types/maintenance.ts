export interface ApiResponse<Scheduled, History> {
    code: number
    message: string
    body: {
        scheduled: Scheduled[]
        history: History[]
    }
}

export interface Maintenance {
    id: string
    description: string
    start: string
    end: string
    status: 'scheduled' | 'in_progress' | 'completed'
    notified: boolean
}
