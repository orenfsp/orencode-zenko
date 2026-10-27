export interface Permission { id: number; code: string; name?: string; description?: string }
export interface Role { id: number; name: string; permissions: Permission[] }
export interface CreateRolePayload { name: string; permission_ids: number[] }
