export interface AccessPermissionsCardProps {
  displayRoles: string[]
  resourceAccess: Record<string, { roles: string[] }>
  groups: string[]
}

export interface AccountDetailsCardProps {
  name?: string
  email?: string
  verified?: boolean
  sub?: string
  username?: string
  iss?: string
  idProvider?: string
  iat?: number
  exp?: number
  rawProfile: unknown
}

export interface ProfileHeroCardProps {
  name?: string
  email?: string
  verified?: boolean
  primaryRole?: string
  iat?: number
  exp?: number
}