let _token: string | null = null

export function setAuthToken(token: string | null): void {
  _token = token
}

export function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers)
  if (_token) {
    headers.set('Authorization', `Bearer ${_token}`)
  }
  return fetch(input, { ...init, headers })
}
