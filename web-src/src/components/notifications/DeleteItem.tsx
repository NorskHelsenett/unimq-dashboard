import { useState } from "react"
import { Button } from "../ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose } from "../ui/dialog"
import { AlarmProps } from "@/types/notifications"
import { RecipientsProps } from "./RecipientCard"
import { deleteRule, deleteRecipient } from '@/services/notifications'

interface DeleteAlarmProps {
    alarm?: AlarmProps | undefined
    recipient?: RecipientsProps | undefined
    vhost: string
    open: boolean
    onClose: () => void
    onDeleted?: () => void
}

export function DeleteItem({ recipient, alarm, vhost, open, onClose, onDeleted }: DeleteAlarmProps) {
    const [deleting, setDeleting] = useState(false)
    const [deleted, setDeleted] = useState(false)

    const handleDelete = () => {
        if (!alarm?.id && !recipient?.id) return
        if (recipient?.id) {
            setDeleting(true)
            Promise.all([
                deleteRecipient(vhost, recipient.id),
                new Promise(res => setTimeout(res, 2000)),
            ]).then(() => {
                setDeleting(false)
                setDeleted(true)
                setTimeout(() => onDeleted ? onDeleted() : window.location.reload(), 1500)
            })
            return
        }
        if (!alarm?.id) return
        setDeleting(true)
        Promise.all([
            deleteRule(vhost, alarm.id),
            new Promise(res => setTimeout(res, 2000)),
        ]).then(() => {
            setDeleting(false)
            setDeleted(true)
            setTimeout(() => onDeleted ? onDeleted() : window.location.reload(), 1500)
        })
    }

    return (
        <Dialog open={open} onOpenChange={(open) => { if (!open && !deleting) { setDeleted(false); onClose() } }}>
            <DialogContent>
                {deleted ? (
                    <div className="flex flex-col items-center gap-3 py-4">
                        <div className="flex size-10 items-center justify-center rounded-full bg-green-100">
                            <svg className="size-5 text-green-600" viewBox="0 0 20 20" fill="currentColor">
                                <path fillRule="evenodd" d="M16.707 5.293a1 1 0 00-1.414 0L8 12.586 4.707 9.293a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l8-8a1 1 0 000-1.414z" clipRule="evenodd" />
                            </svg>
                        </div>
                        <p className="text-sm font-medium text-text-primary text-center">
                            {recipient ? (
                                <>
                                    <span className="font-semibold">{recipient.name}</span> deleted from <span className="font-semibold">{vhost}</span>
                                </>
                            ) : (
                                <>
                                    <span className="font-semibold">{alarm?.name}</span> deleted from <span className="font-semibold">{vhost}</span>
                                </>
                            )}
                        </p>
                    </div>
                ) : (
                    <>
                        <DialogHeader>
                            <DialogTitle>Delete {alarm ? 'alarm' : 'recipient'}</DialogTitle>
                            <DialogDescription>
                                Are you sure you want to delete <span className="font-medium text-text-primary">{alarm ? alarm.name : recipient?.name}</span>? This cannot be undone.
                            </DialogDescription>
                        </DialogHeader>
                        <DialogFooter>
                            <DialogClose asChild>
                                <Button variant="outline" className="bg-gray-100" disabled={deleting}>Cancel</Button>
                            </DialogClose>
                            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
                                {deleting ? (
                                    <span className="flex items-center gap-2">
                                        <span className="size-4 border-2 border-red-300 border-t-white rounded-full animate-spin inline-block" />
                                        Deleting…
                                    </span>
                                ) : 'Delete'}
                            </Button>
                        </DialogFooter>
                    </>
                )}
            </DialogContent>
        </Dialog>
    )
}
