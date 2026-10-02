import { ACL } from '@/types/acl'
import { Scope } from '@/types/acl'
import { Sheet, SheetContent, SheetFooter, SheetHeader, SheetTitle } from "../ui/sheet"
import { useEffect, useState } from "react"
import { MultiSelect } from '../ui/multi-select'
import { Button } from '../ui/button'
import { Info, ShieldCheck, UserRound } from 'lucide-react'
import { upsertACL } from '@/services/acl'
import { Response } from '../ui/response'


interface EditAclProps {
    acl: ACL
    vhosts: string[]
    open: boolean
    onClose: () => void
}

export const EditAcl = ({ acl, vhosts, open, onClose }: EditAclProps) => {
    const [vhostIds, setVhostIds] = useState<string[]>(acl.vhost_ids)
    const [permissions, setPermissions] = useState<Scope[]>(acl.permissions)
    const [isSaving, setIsSaving] = useState(false)
    const [response, setResponse] = useState<{ open: boolean, status: 'success' | 'error', message: string }>({ open: false, status: 'success', message: '' })

    useEffect(() => {
        setVhostIds(acl.vhost_ids)
        setPermissions(acl.permissions)
    }, [acl])

    const togglePermission = (permission: Scope) => {
        setPermissions(current => current.includes(permission)
            ? current.filter(item => item !== permission)
            : [...current, permission]
        )
    }

    const handleSave = async () => {
        if (vhostIds.length === 0 || permissions.length === 0) return

        setIsSaving(true)
        try {
            await upsertACL({ group: acl.group, vhost_ids: vhostIds, permissions })
            setResponse({ open: true, status: 'success', message: 'ACL updated successfully.' })
        } catch (error) {
            setResponse({
                open: true,
                status: 'error',
                message: error instanceof Error ? error.message : 'Failed to update ACL.',
            })
        } finally {
            setIsSaving(false)
        }
    }

    const permissionDetails: { value: Scope; label: string; description: string; color: string }[] = [
        { value: 'read', label: 'Read', description: 'View queues, exchanges and messages', color: 'bg-emerald-500' },
        { value: 'write', label: 'Write', description: 'Publish and consume messages', color: 'bg-blue-500' },
        { value: 'admin', label: 'Admin', description: 'Manage vhosts, policies and other settings', color: 'bg-violet-500' },
    ]

    return (
        <Sheet open={open} onOpenChange={o => !o && onClose()}>
            <SheetContent side="right" className="w-full gap-0 border-l border-border-card bg-surface-card p-0 sm:max-w-lg">
                <Response
                    open={response.open}
                    onClose={() => {
                        const wasSuccessful = response.status === 'success'
                        setResponse(current => ({ ...current, open: false }))
                        if (wasSuccessful) {
                            onClose()
                            window.location.reload()
                        }
                    }}
                    status={response.status}
                    message={response.message}
                />
                <SheetHeader className="border-b border-border-card px-6 py-5 pr-16">
                    <div className="mb-1 flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-text-muted">
                        <ShieldCheck className="size-4 text-blue-600" /> Access control
                    </div>
                    <SheetTitle className="text-2xl font-bold tracking-tight text-text-primary">Edit ACL</SheetTitle>
                </SheetHeader>
                <div className="flex-1 overflow-y-auto px-6 py-7">
                    <div className="space-y-8">
                        <section className="space-y-3">
                            <div>
                                <h2 className="text-sm font-semibold text-text-primary">Group</h2>
                                <p className="mt-1 text-sm text-text-muted">The group that receives these access rules.</p>
                            </div>
                            <div className="flex h-12 items-center gap-3 rounded-lg border border-border-card bg-surface-page px-3.5 text-sm font-medium text-text-primary shadow-sm">
                                <UserRound className="size-5 text-text-muted" />
                                {acl.group}
                            </div>
                        </section>

                        <section className="space-y-3">
                            <div>
                                <h2 className="text-sm font-semibold text-text-primary">Vhosts</h2>
                                <p className="mt-1 text-sm text-text-muted">Select which vhosts this group should have access to.</p>
                            </div>
                            <MultiSelect
                                options={[{ value: '*', label: 'All vhosts' }, ...vhosts.map(value => ({ value }))]}
                                value={vhostIds}
                                onValueChange={(nextVhostIds) => setVhostIds(
                                    nextVhostIds.includes('*')
                                        ? vhostIds.includes('*') ? nextVhostIds.filter(vhostId => vhostId !== '*') : ['*']
                                        : nextVhostIds
                                )}
                                placeholder="Select vhosts"
                                className="[&_button]:h-12 [&_button]:rounded-lg [&_button]:px-3.5"
                            />
                        </section>

                        <section className="space-y-3">
                            <div>
                                <h2 className="text-sm font-semibold text-text-primary">Permissions</h2>
                                <p className="mt-1 text-sm text-text-muted">Choose what this group can do on the selected vhosts.</p>
                            </div>
                            <div className="divide-y divide-border-card rounded-lg border border-border-card bg-surface-card">
                                {permissionDetails.map(({ value, label, description, color }) => {
                                    const selected = permissions.includes(value)
                                    return (
                                        <button
                                            key={value}
                                            type="button"
                                            aria-pressed={selected}
                                            onClick={() => togglePermission(value)}
                                            className="flex w-full items-center gap-3.5 px-4 py-3.5 text-left transition-colors hover:bg-surface-page focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
                                        >
                                            <span className={`flex size-5 shrink-0 items-center justify-center rounded-md border ${selected ? `${color} border-transparent text-white` : 'border-input bg-surface-card'}`}>
                                                {selected && <span className="text-sm font-bold leading-none">✓</span>}
                                            </span>
                                            <span className="min-w-0">
                                                <span className="block text-sm font-semibold text-text-primary">{label}</span>
                                                <span className="mt-0.5 block text-xs leading-5 text-text-muted">{description}</span>
                                            </span>
                                        </button>
                                    )
                                })}
                            </div>
                        </section>

                        <div className="flex gap-3 rounded-lg border border-blue-100 bg-blue-50/70 px-4 py-3.5 text-sm text-blue-950">
                            <Info className="mt-0.5 size-5 shrink-0 text-blue-600" />
                            <p className="leading-6">This group will have the selected permissions on the chosen vhosts. You can update these settings at any time.</p>
                        </div>
                    </div>
                </div>
                <SheetFooter className="mt-0 flex-row justify-end gap-3 border-t border-border-card bg-surface-card px-6 py-5">
                    <Button type="button" variant="outline" className="h-11 min-w-28" onClick={onClose}>Cancel</Button>
                    <Button type="button" className="h-11 min-w-36 bg-submit-button text-white hover:bg-submit-button/90" disabled={isSaving || vhostIds.length === 0 || permissions.length === 0} onClick={handleSave}>
                        {isSaving ? 'Saving...' : 'Save changes'}
                    </Button>
                </SheetFooter>
            </SheetContent>
        </Sheet>
    )
}