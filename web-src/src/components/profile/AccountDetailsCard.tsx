import { useState } from 'react'
import { Eye, EyeOff, Copy, Check, User, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDate } from './profileUtils'
import { AccountDetailsCardProps } from '@/types/profile'

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between px-5 py-3 gap-4">
      <span className="text-xs text-text-muted uppercase tracking-wide shrink-0">{label}</span>
      <div className="min-w-0">{children}</div>
    </div>
  )
}

export function AccountDetailsCard({
  name, email, verified, sub, username, iss, idProvider, iat, exp, rawProfile,
}: AccountDetailsCardProps) {
  const [showSub, setShowSub] = useState(false)
  const [copied, setCopied] = useState(false)
  const [showClaims, setShowClaims] = useState(false)

  const copyIss = () => {
    if (!iss) return
    navigator.clipboard.writeText(iss).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div className="border border-border-card rounded-xl bg-surface-card flex flex-col">
      <div className="flex items-center gap-2 px-5 py-4 border-b border-border-card">
        <User size={16} className="text-text-muted" />
        <div>
          <h3 className="font-semibold text-text-primary text-sm">Account details</h3>
          <p className="text-xs text-text-muted">Your account and authentication information.</p>
        </div>
      </div>
      <div className="divide-y divide-border-card flex-1">
        {name && (
          <Row label="Full name"><span className="text-sm text-text-primary">{name}</span></Row>
        )}
        {email && (
          <Row label="Email">
            <span className="flex items-center gap-2 text-sm text-text-primary">
              {email}
              {verified && <span className="px-1.5 py-0.5 rounded-full bg-status-ok/10 text-status-ok text-xs">Verified</span>}
            </span>
          </Row>
        )}
        {username && (
          <Row label="Username"><span className="text-sm text-text-primary">{username}</span></Row>
        )}
        {sub && (
          <Row label="User ID (sub)">
            <div className="flex items-center gap-2 min-w-0">
              <span className="text-xs text-text-muted font-mono truncate">
                {showSub ? sub : `••••••••`}
              </span>
              <button onClick={() => setShowSub(v => !v)} className="text-text-muted hover:text-text-primary transition-colors shrink-0">
                {showSub ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
            </div>
          </Row>
        )}
        {idProvider && (
          <Row label="Identity provider"><span className="text-sm text-text-primary">{idProvider}</span></Row>
        )}
        {iss && (
          <Row label="Issuer (iss)">
            <div className="flex items-center gap-2 min-w-0">
              <span className="text-xs text-text-muted font-mono truncate">{iss}</span>
              <button onClick={copyIss} className="text-text-muted hover:text-text-primary transition-colors shrink-0">
                {copied ? <Check size={14} className="text-status-ok" /> : <Copy size={14} />}
              </button>
            </div>
          </Row>
        )}
        {iat && <Row label="Signed in"><span className="text-sm text-text-primary">{formatDate(iat)}</span></Row>}
        {exp && <Row label="Session expires"><span className="text-sm text-text-primary">{formatDate(exp)}</span></Row>}
      </div>
      <button
        onClick={() => setShowClaims(v => !v)}
        className="flex items-center justify-between px-5 py-3 border-t border-border-card text-sm text-text-muted hover:bg-surface-page transition-colors rounded-b-xl"
      >
        <span>Technical details (token claims)</span>
        {showClaims ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
      </button>
      {showClaims && (
        <div className="px-5 pb-4 border-t border-border-card">
          <pre className="text-xs text-text-muted bg-surface-page rounded-lg p-3 overflow-auto max-h-64 mt-3">
            {JSON.stringify(rawProfile, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
