export type Scope = 'read' | 'write' | 'admin'

export interface ACL{
    Group: string
    Permissions: Scope[]
    VhostIDs: string[]
}