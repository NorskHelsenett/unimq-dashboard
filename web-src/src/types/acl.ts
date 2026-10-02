export type Scope = 'read' | 'write' | 'admin'

export interface ACL{
    group: string
    permissions: Scope[]
    vhost_ids: string[]
}