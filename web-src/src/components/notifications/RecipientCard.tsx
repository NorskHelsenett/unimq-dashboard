import { useState } from "react"
import { Button } from "../ui/button"
import { SelectLabel, Selector, SelectorContent, SelectorItem, SelectorTrigger, SelectorValue } from "../ui/selector"
import { Input } from "../ui/input"
import { DeleteItem } from "./DeleteItem"
import { addRecipient } from '@/services/notifications'
import { StatusDot } from "../ui/status-dot"
import { SectionCard, SectionCardHeader } from "../ui/section-card"
import { Users } from "lucide-react"


export interface RecipientsProps {
    id: string
    name: string
    url: string
    type: string // slack, teams, webhook
}

const recipientTypeOptions = [
    { value: 'teams', label: 'Microsoft Teams' },
    { value: 'slack', label: 'Slack' },
    { value: 'webhook', label: 'Webhook' },
]

function ExisitingRecipients({existingRecipients, vhost}: {existingRecipients: RecipientsProps[], vhost: string}) {
    const recipients = existingRecipients || []
    const [deletingId, setDeletingId] = useState<string | null>(null)
    const deletingRecipient = recipients.find(r => r.id === deletingId)

    return (
        <div className="mt-2">
            <DeleteItem recipient={deletingRecipient} vhost={vhost} open={deletingId !== null} onClose={() => setDeletingId(null)} />
            {recipients.length === 0 ? (
                <p className="text-sm text-text-muted py-2">No recipients configured for this vhost.</p>
            ) : (
                <div className="overflow-y-auto max-h-64">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Name</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Type</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted">Endpoint</th>
                                <th className="border-b border-border-card py-2 px-4 text-xs text-text-muted text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {recipients.map(recipient => (
                                <tr key={recipient.id} className="hover:bg-surface-page">
                                    <td className="border-b border-border-card py-2 px-4 text-sm font-medium">
                                        <span className="flex items-center gap-2">
                                            <StatusDot color="blue" pulse={false} />
                                            {recipient.name}
                                        </span>
                                    </td>
                                    <td className="border-b border-border-card py-2 px-4 text-sm text-text-muted">
                                        {recipient.type}
                                    </td>
                                    <td className="border-b border-border-card py-2 px-4 text-sm text-text-muted truncate">
                                        {recipient.url}
                                    </td>
                                    <td className="border-b border-border-card py-2 px-4">
                                        <div className="flex items-center justify-end">
                                            <Button variant="destructive" size="xs" onClick={() => recipient.id && setDeletingId(recipient.id)}>Delete</Button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}

function AddRecipientForm({selectedType, vhost, onClose}: {selectedType: string, vhost: string, onClose: () => void}) {
    if (!selectedType) return null

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        const fd = new FormData(e.currentTarget)
        addRecipient(vhost, {
            name: fd.get('name') as string,
            type: selectedType,
            url: (fd.get('url') as string) || undefined,
            email: (fd.get('email') as string) || undefined,
        }).then(() => window.location.reload())
    }

    return (
        <form onSubmit={handleSubmit} className="bg-surface-card border border-blue-200 rounded-lg p-4 mb-4 shadow">
            <div className="flex flex-wrap items-end gap-2">
                <div className="flex flex-col gap-1 flex-1 min-w-[160px]">
                    <label className="text-sm font-medium">{recipientTypeOptions.find(o => o.value === selectedType)?.label}</label>
                    <Input name="name" placeholder="E.g. #alerts-channel" className="bg-surface-card" required />
                </div>
                <div className="flex flex-col gap-1 flex-1 min-w-[200px]">
                    <label className="text-xs text-text-muted">Webhook URL</label>
                    <Input name="url" type="url" placeholder="https://hooks.example.com/..." className="bg-surface-card" required />
                </div>
                <Button type="submit" variant="orange" size="sm">Save</Button>
                <Button type="button" variant="ghost" size="sm" onClick={onClose}>Cancel</Button>
            </div>
        </form>
    )
}

export function RecipientCard({existingRecipients, vhost}: {existingRecipients: RecipientsProps[], vhost: string}) {
    const [selectedType, setSelectedType] = useState('')

    return (
        <SectionCard accent="blue">
            <SectionCardHeader
                title="Recipients"
                icon={<Users className="w-5 h-5 text-blue-400" />}
                action={
                    <Selector value={selectedType} onValueChange={(val) => setSelectedType(val)}>
                        <SelectorTrigger className="w-60 ml-3">
                            <SelectorValue placeholder="Add new recipient" />
                        </SelectorTrigger>
                        <SelectorContent>
                            <SelectLabel>Select type</SelectLabel>
                            {recipientTypeOptions.map(option => (
                                <SelectorItem key={option.value} value={option.value}>
                                    {option.label}
                                </SelectorItem>
                            ))}
                        </SelectorContent>
                    </Selector>
                }
            />
            <AddRecipientForm selectedType={selectedType} vhost={vhost} onClose={() => setSelectedType('')} />
            {selectedType && <hr className="border-border-card mb-4" />}
            <ExisitingRecipients existingRecipients={existingRecipients} vhost={vhost} />
        </SectionCard>
    )
}