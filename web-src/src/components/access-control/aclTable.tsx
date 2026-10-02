import { ACL } from "@/types/acl";
import { SectionCard, SectionCardHeader } from "../ui/section-card"
import { Pencil, ShieldCheck, ShieldUser, Trash2, Users, Server } from "lucide-react"
import {Pill} from "../ui/pill"
import { Button } from "../ui/button"
import { MultiSelect } from "../ui/multi-select"
import { useState } from "react"
import { upsertACL } from "@/services/acl";
import { Response } from "../ui/response"
import { DeleteAcl } from "./DeleteAcl";
import { EditAcl } from "./EditAcl";


function AclTable({ acls, vhosts }: { acls: ACL[], vhosts: string[] }) {
    const [showDeleteDialog, setShowDeleteDialog] = useState(false)
    const [selectedAcl, setSelectedAcl] = useState<ACL | null>(null)
    const [inEditMode, setInEditMode] = useState(false)

    const permissionPillVariant = (permission: string) => {
        switch (permission) {
            case 'read':
                return 'bg-green-100 text-green-800';
            case 'write':
                return 'bg-blue-100 text-blue-800';
            case 'admin':
                return 'bg-violet-100 text-violet-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    }

    return (
         <div className="overflow-y-auto max-h-360">
            {showDeleteDialog && selectedAcl && (
                <DeleteAcl 
                    acl={selectedAcl} 
                    open={showDeleteDialog} 
                    onClose={() => setShowDeleteDialog(false)} 
                    onDeleted={() => { setShowDeleteDialog(false); window.location.reload() }} 
                />
            )}
             {selectedAcl &&
                <EditAcl acl={selectedAcl} vhosts={vhosts} open={inEditMode} onClose={() => setInEditMode(false)} />
            }
            {acls.length === 0 ? (
                <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border-card bg-surface-page px-6 py-10 text-center">
                    <ShieldCheck className="mb-3 size-8 text-blue-400" />
                    <p className="text-sm font-medium text-text-primary">No access rules yet</p>
                    <p className="mt-1 text-sm text-text-muted">Add an ACL to start managing group access.</p>
                </div>
            ) : <table className='w-full text-left border-collapse'>
                <thead>
                    <tr>
                        <th scope='col' className="border-b border-border-card py-2 px-4 text-xs font-medium text-text-muted">
                            Group
                        </th>
                        <th scope='col' className="border-b border-border-card py-2 px-4 text-xs font-medium text-text-muted">
                            Vhosts
                        </th>
                        <th scope='col' className="border-b border-border-card py-2 px-4 text-xs font-medium text-text-muted">
                            Permissions
                        </th>
                        <th scope='col' className="border-b border-border-card py-2 px-4 text-right text-xs font-medium text-text-muted">
                            Actions
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {acls.map((acl) => {
                        return (
                            <tr key={acl.group} className="cursor-pointer transition-colors hover:bg-surface-page" onClick={() => { setSelectedAcl(acl); setInEditMode(true) }}>
                                <td className='border-b border-border-card py-3 px-4 text-sm font-medium text-text-primary'>
                                    <span className="flex items-center gap-2">
                                        <span className="flex size-7 items-center justify-center rounded-md bg-blue-50 text-blue-600">
                                            <Users className="size-4" />
                                        </span>
                                        {acl.group}
                                    </span>
                                </td>
                                <td className='border-b border-border-card py-3 px-4 text-sm text-text-secondary'>
                                    <span className="flex items-center gap-2">
                                        <Server className="size-4 text-text-muted" />
                                        {acl.vhost_ids.includes('*') ? 'All vhosts' : acl.vhost_ids.join(', ')}
                                    </span>
                                </td>
                                <td className='border-b border-border-card py-3 px-4 text-sm'>
                                    {acl.permissions.map((permission) => (
                                        <Pill key={permission} className={`${permissionPillVariant(permission)} mr-1.5 border-0 px-2 text-xs`}>
                                            {permission}
                                        </Pill>
                                    ))}
                                </td>
                                <td className="border-b border-border-card py-3 px-4" onClick={event => event.stopPropagation()}>
                                    <div className="flex items-center justify-end gap-1">
                                        <Button variant="ghost" size="sm" className="size-8 p-0 text-text-muted hover:text-text-primary" aria-label={`Edit ${acl.group}`} onClick={() => { setSelectedAcl(acl); setInEditMode(true) }}>
                                            <Pencil className="size-4" />
                                        </Button>
                                        <Button variant="ghost" size="sm" className="size-8 p-0 text-text-muted hover:text-destructive" aria-label={`Delete ${acl.group}`} onClick={() => {
                                            setSelectedAcl(acl)
                                            setShowDeleteDialog(true)
                                        }}>
                                            <Trash2 className="size-4" />
                                        </Button>
                                    </div>
                                </td>
                            </tr>
                        )
                    })}
                </tbody>
            </table>}
            
        </div>
    )
}



function AddAclForm({acls, vhosts, onClose, onCancel, onError} : {acls: ACL[], vhosts: string[], onClose: () => void, onCancel: () => void, onError: (msg: string) => void}) {
    const [validationError, setValidationError] = useState<string | null>(null)
    const [vhostIds, setVhostIds] = useState<string[]>([])
    const [permissions, setPermissions] = useState<string[]>([])

    return(
        <form onSubmit={(e) => {
            e.preventDefault()
            const fd = new FormData(e.currentTarget)
            const group = fd.get('group') as string
            if (acls.some(acl => acl.group === group)) {
                setValidationError('An ACL for this group already exists.')
                return
            }
            if (vhostIds.length === 0) {
                setValidationError('At least one vhost is required.')
                return
            }
            if (permissions.length === 0) {
                setValidationError('At least one permission is required.')
                return
            }
            if (permissions.some(s => !['read', 'write', 'admin'].includes(s))) {
                setValidationError('Permissions must be valid.')
                return
            }
            setValidationError(null)
            upsertACL({group: group, vhost_ids: vhostIds, permissions: permissions}).then(() => {
                onClose()
            }).catch((err) => {
                onError(err.message ?? "Failed to add ACL")
            })

        }} className="mb-4 rounded-lg border border-blue-200 bg-surface-page p-4 shadow-sm">
            {validationError && <p className="mb-3 rounded-md bg-red-50 px-3 py-2 text-xs text-destructive">{validationError}</p>}
                <div className="mb-3 flex flex-col gap-1">
                    <label htmlFor="group" className="text-xs font-medium text-text-secondary">Group</label>
                    <input type="text" name="group" id="group" className="h-10 rounded-md border border-border-card bg-surface-card px-3 text-sm outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100" required />
            </div>
            <div className="grid grid-cols-2 gap-2 mb-2">
                <div className="flex flex-col gap-1">
                    <span className="text-xs text-text-muted">Vhosts</span>
                    <MultiSelect
                        options={[{ value: '*', label: 'All vhosts' }, ...vhosts.map((vhost) => ({ value: vhost }))]}
                        value={vhostIds}
                        onValueChange={(nextVhostIds) => {
                            setVhostIds(
                                nextVhostIds.includes('*')
                                    ? vhostIds.includes('*')
                                        ? nextVhostIds.filter((vhostId) => vhostId !== '*')
                                        : ['*']
                                    : nextVhostIds
                            )
                        }}
                        placeholder="Select vhosts"
                    />
                </div>
                <div className="flex flex-col gap-1">
                    <span className="text-xs text-text-muted">Permissions</span>
                    <MultiSelect
                        options={['read', 'write', 'admin'].map((permission) => ({ value: permission }))}
                        value={permissions}
                        onValueChange={setPermissions}
                        placeholder="Select permissions"
                    />
                </div>
            </div>
            <div className="flex justify-end gap-2">
                <Button type="button" variant="ghost" size="sm" onClick={onCancel}>Cancel</Button>
                <Button type="submit" size="sm" className="bg-submit-button text-white hover:bg-submit-button/90">Add ACL</Button>
            </div>
        </form>

    )
}

export function AclCard({ acls, vhosts, onRefresh }: {acls: ACL[], vhosts: string[], onRefresh: () => void}) {
    const [showForm, setShowForm] = useState(false)
    const [response, setResponse] = useState<{ open: boolean, status: 'success' | 'error', message: string }>({ open: false, status: 'success', message: '' })

    
    return(
        <div className='mt-2'>
           
            <Response 
                open={response.open}
                onClose={() => {
                    setResponse(r => ({ ...r, open: false }))
                    if (response.status === 'success') onRefresh()
                }}
                status={response.status}
                message={response.message}
            />
            
            <SectionCard accent="blue" className="overflow-hidden">
                <SectionCardHeader 
                    title="Access Control List"
                    icon={<ShieldUser className="h-5 w-5 text-blue-400"/>}
                    action={
                            <Button variant="outline" size="sm" onClick={() => setShowForm(true)} disabled={showForm}>
                                Add ACL
                            </Button>
                        }
                />
                
                <p className="mb-4 max-w-xl text-sm text-text-secondary">
                    Manage access control to different vhosts and what permissions they have.
                </p>
                <div className="mb-5 grid grid-cols-3 divide-x divide-border-card rounded-lg border border-border-card bg-surface-page">
                    <div className="px-4 py-3">
                        <p className="text-xl font-semibold text-text-primary">{acls.length}</p>
                        <p className="text-xs text-text-muted">Groups</p>
                    </div>
                    <div className="px-4 py-3">
                        <p className="text-xl font-semibold text-text-primary">{vhosts.length}</p>
                        <p className="text-xs text-text-muted">Available vhosts</p>
                    </div>
                    <div className="px-4 py-3">
                        <p className="text-xl font-semibold text-text-primary">{new Set(acls.flatMap(acl => acl.permissions)).size}</p>
                        <p className="text-xs text-text-muted">Permission scopes</p>
                    </div>
                </div>
                {showForm && 
                    <AddAclForm 
                        acls={acls} 
                        vhosts={vhosts} 
                        onClose={() => { setShowForm(false); setResponse({ open: true, status: 'success', message: 'ACL added successfully' }) }} 
                        onCancel={() => setShowForm(false)} 
                        onError={(msg) => { setShowForm(false); setResponse({ open: true, status: 'error', message: msg }) } }
                    />
                }
                <AclTable acls={acls} vhosts={vhosts} />
            </SectionCard>
        </div>
    )
}

