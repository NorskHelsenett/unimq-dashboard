export const LogoLink = ({ collapsed }: { collapsed: boolean }) => (
    <a href="/" className="flex items-center gap-6 mx-4">
        <img src="./../static/logo.png" width={64} className="shrink-0" />
        {!collapsed && <h1 className="text-4xl">UniMQ</h1>}
    </a>
)