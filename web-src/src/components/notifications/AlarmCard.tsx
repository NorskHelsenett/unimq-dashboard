import { Selector, SelectorTrigger, SelectorContent, SelectorItem, SelectorValue, SelectLabel } from "../ui/selector"

interface AlarmCardProps {
    name: string
    status: string //"firing", "ok"
    threshold: number
    currentValue: number
}

const alarmDropdownOptions = [
    { value: 'connections', label: 'Connections' },
    { value: 'channels', label: 'Channels' },
    { value: 'queues', label: 'Queues' },
    { value: 'unacked', label: 'Unacknowledged Messages' },
    { value: 'queue_messages', label: 'Messages in Queue' },
    { value: 'queue_size', label: 'Queue Size' },
    { value: 'no_consumers', label: 'No Consumers' },
    { value: 'maintanance', label: 'Maintenance Message' },
]

function AddAlarmForm(alarm: String) {

    
}


export function AlarmCard() {
    return (
        <div className="bg-white rounded-lg shadow p-4 min-w-[300px] border border-border-card">
            <div className="flex items-center justify-between mb-2">
                <h3 className="text-lg font-semibold">Alarmer</h3>
                <Selector>
                    <SelectorTrigger className="w-60 ml-3 placeholder:text-gray-400">
                        <SelectorValue placeholder="Legg til ny alarm" />
                    </SelectorTrigger>
                    <SelectorContent>
                    <SelectLabel>Velg metrikk</SelectLabel>
                        {alarmDropdownOptions.map(option => (
                            <SelectorItem key={option.value} value={option.value}>
                                {option.label}
                            </SelectorItem>
                        ))}
                    </SelectorContent>
                </Selector>
            </div>
            <p className="text-sm text-gray-600 mb-4 max-w-xl">
                Sett opp alarmer for å varsle teamet når grenser nås. 
                Varsler sendes kun én gang per utløsning og nullstilles automatisk når verdien er tilbake under grensen.
            </p>
            
        </div>
    )
}