import { useState } from "react"
import { Button } from "../ui/button"
import { SelectLabel, Selector, SelectorContent, SelectorItem, SelectorTrigger, SelectorValue } from "../ui/selector"
import { Input } from "../ui/input"
import { DeleteItem } from "./DeleteItem"


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
    //Rewrite the deletealarm component to handle both recipients and alarms
    
    return (
        <div className="mt-4">
            <DeleteItem recipient={deletingRecipient} vhost={vhost} open={deletingId !== null} onClose={() => setDeletingId(null)} />
            <div className="flex flex-col divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
                {recipients.length != null ? (
                    recipients.map(recipient => (
                        <div key={recipient.id} className="flex items-center justify-between p-4 hover:bg-gray-50">
                            <div className="flex items-center gap-3">
                                <span className="relative flex size-3 shrink-0 items-center justify-center">
                                    <span className="relative inline-flex size-2 rounded-full bg-blue-500" />
                                </span>
                                <div>
                                    <p className="text-sm font-medium text-gray-900">{recipient.name}</p>
                                    <p className="text-xs text-gray-500">{recipient.type} · {recipient.url}</p>
                                </div>
                            </div>
                            <div className="flex items-center gap-2">
                                <Button variant="destructive" size="xs" onClick={() => recipient.id && setDeletingId(recipient.id)}>Delete</Button>
                            </div>
                        </div>
                    ))
                ) : (
                    <p className="p-4 text-sm text-gray-500">No recipients configured for this vhost.</p>
                )}

            </div>
        </div>
    )
}

function AddRecipientForm({selectedType, vhost, onClose}: {selectedType: string, vhost: string, onClose: () => void}) {
    if (!selectedType) return null

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        const form = e.currentTarget
        const data = new FormData(form)
        data.set('vhost', vhost)
        data.set('type', selectedType)
        fetch('/notifications/recipients/add', { method: 'POST', body: data }).then(() => {
            window.location.reload()
        })
    }

    return (
        <form onSubmit={handleSubmit} className="bg-gray-50 border border-gray-300 rounded-lg p-4">
            <h2 className="text-sm font-semibold mb-3">New recipient: {recipientTypeOptions.find(o => o.value === selectedType)?.label}</h2>
            {selectedType && (
                <div className="grid grid-cols-2 gap-4 mb-3">
                    <div>
                        <p className="text-sm mb-1">Name</p>
                        <Input name="name" placeholder="E.g #alerts-kanalen" className="bg-white" required />
                    </div>
                    <div>
                        <p className="text-sm mb-1">Webhook URL</p>
                        <Input name="url" type='url' placeholder="testtest" className="bg-white" required />
                    </div>
                </div>
            )}
            <div className="flex justify-between mt-2">
                <Button type="submit">Add recipient</Button>
                <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            </div>
        </form>
    )
}

export function RecipientCard({existingRecipients, vhost}: {existingRecipients: RecipientsProps[], vhost: string}) {
    const [selectedType, setSelectedType] = useState('')

    return (
        <div className="bg-white shadow rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">

                <h2 className="text-lg font-semibold">Existing Recipients</h2>
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
            </div>

            <AddRecipientForm selectedType={selectedType} vhost={vhost} onClose={() => setSelectedType('')} />
            <ExisitingRecipients existingRecipients={existingRecipients} vhost={vhost} />
        </div>
    )
}