import { useState } from "react"
import { Button } from "../ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose } from "../ui/dialog"
import { ACL } from "@/types/acl"
import { deleteAcl } from "@/services/acl"

interface DeleteAclProps {
    acl: ACL
    open: boolean
    onClose: () => void
    onDeleted?: () => void
}

export function DeleteAcl({ acl, open, onClose, onDeleted }: DeleteAclProps) {
    const [deleting, setDeleting] = useState(false)
    const [deleted, setDeleted] = useState(false)

    const handleDelete = () => {
        setDeleting(true)
        Promise.all([
            deleteAcl(acl.group),
            new Promise(res => setTimeout(res, 2000)),
        ])
            .then(() => {
                setDeleting(false)
                setDeleted(true)
                setTimeout(() => onDeleted ? onDeleted() : window.location.reload(), 1500)
            })
            .catch(() => {
                setDeleting(false)
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
                            <span className="font-semibold">{acl.group}</span> deleted
                        </p>
                    </div>
                ) : (
                    <>
                        <DialogHeader>
                            <DialogTitle>Delete ACL</DialogTitle>
                            <DialogDescription>
                                Are you sure you want to delete <span className="font-medium text-text-primary">{acl.group}</span>? This cannot be undone.
                            </DialogDescription>
                        </DialogHeader>
                        <DialogFooter>
                            <DialogClose asChild>
                                <Button variant="outline" className="bg-surface-page" disabled={deleting}>Cancel</Button>
                            </DialogClose>
                            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
                                {deleting ? (
                                    <span className="flex items-center gap-2">
                                        <span className="size-4 border-2 border-red-300 border-t-surface-card rounded-full animate-spin inline-block" />
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
