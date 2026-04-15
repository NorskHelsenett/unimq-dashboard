import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose } from "../ui/dialog"
import { Button } from "../ui/button"

export function Response({ open, onClose, status, message }: { open: boolean, onClose: () => void, status: 'success' | 'error', message: string }) {
    return(
        <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle className={status === 'success' ? 'text-green-700' : 'text-destructive'}>
                        {status === 'success' ? 'Success' : 'Error'}
                    </DialogTitle>
                    <DialogDescription>{message}</DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <DialogClose asChild>
                        <Button variant="outline" onClick={onClose}>Close</Button>
                    </DialogClose>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}