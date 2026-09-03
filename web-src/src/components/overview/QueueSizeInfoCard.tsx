import { Tile } from "../layout/Tile";

export function QueueSizeInfoCard() {
  return (
    <Tile className="flex-1 min-w-sm max-w-xl 2xl:max-w-fit">
      <h3 className="flex items-center gap-2 font-semibold mb-2">
        <span className="text-amber-700">&#9432;</span>Kø-størrelse
      </h3>
      <div className="space-y-2">
        <p>
          En kø kan inneholde <strong>20 000 meldinger</strong> eller <strong>1 GiB</strong>. Når denne terskelverdien nås, vil de eldste meldingene skrives over uten forvarsel!
        </p>
        <p>
          For å unngå meldingstap kan du sette overflow til{' '}
          <code className="bg-surface-page border border-border-card px-1 py-0.5 rounded text-xs font-mono whitespace-nowrap">
            reject-publish
          </code>
          , samt sette{' '}
          <code className="bg-surface-page border border-border-card px-1 py-0.5 rounded text-xs font-mono whitespace-nowrap">
            publish confirm
          </code>
          . Dette vil sørge for at publisher informeres dersom meldinger ikke kan sendes til kø.
        </p>
        <p>
          For mer informasjon og tips, se{' '}
          <a
            href="https://docs.nhn.no/k8s/brukerdokumentasjon/rabbitmq/best-practice2.html"
            target="_blank"
            rel="noopener noreferrer"
            className="text-amber-600 underline hover:text-amber-800"
          >RabbitMQ best practice documentation</a>.
        </p>
      </div>
    </Tile>
  )
}