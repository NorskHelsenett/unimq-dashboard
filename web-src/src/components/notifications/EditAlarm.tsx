import { AlarmProps } from "./AlarmCard"

function AlarmOverview({ alarm }: { alarm: AlarmProps }) {
    return (
        <div>
            <h2 className="text-2xl font-bold mb-4">{alarm.name}</h2>
            <p>Type: {alarm.type}</p>
            {alarm.queue_name && <p>Queue: {alarm.queue_name}</p>}
            {alarm.threshold && <p>Threshold: {alarm.threshold}</p>}
            <p>Status: {alarm.status}</p>
            {alarm.last_value && <p>Last value: {alarm.last_value}</p>}
        </div>
    )
}

export const EditAlarm = ({ alarm }: { alarm: AlarmProps }) => {
        return (
            <div>
                <AlarmOverview alarm={alarm} />
            </div>
        )

}