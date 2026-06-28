import type {
  AuthToken,
  UserLoginIn,
  UserRegisterIn,
  UserPrivateOut,
  UserPublicOut,
  ClientPublicOut,
  ClientPrivateOut,
  HysteriaClientPrivateOut,
  ConnectionLinkOut,
  InviteIn,
  UpdateClientAliasIn,
  InviteActivateIn,
  InviteOut,
  InviteCheckOut,
  Role,
} from '@/types';

const BASE_URL = '/api/v1';

function getToken(): string | null {
  return localStorage.getItem('token');
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error_text: 'Неизвестная ошибка' }));
    throw new Error(body.error_text || `Ошибка ${response.status}`);
  }

  return response.json();
}

// --- Security ---

export async function register(data: UserRegisterIn): Promise<UserPublicOut> {
  return request<UserPublicOut>('/security/register', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function login(data: UserLoginIn): Promise<AuthToken> {
  return request<AuthToken>('/security/login', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getMe(): Promise<UserPrivateOut> {
  return request<UserPrivateOut>('/security/me');
}

export async function getUserById(userId: number): Promise<UserPublicOut> {
  return request<UserPublicOut>(`/security/users/${userId}`);
}

export async function setRoleForUser(userId: number, roleId: number): Promise<UserPrivateOut> {
  return request<UserPrivateOut>(`/security/users/${userId}/set_role/${roleId}`, {
    method: 'POST',
  });
}

export async function getRoleById(roleId: number): Promise<Role> {
  return request<Role>(`/security/roles/${roleId}`);
}

// --- XRay / VLESS Clients ---

export async function getVlessClientsByUserId(userId: number): Promise<ClientPublicOut[]> {
  return request<ClientPublicOut[]>(`/xray/clients/by_user_id/${userId}`);
}

export async function getVlessClientById(clientId: string): Promise<ClientPrivateOut> {
  return request<ClientPrivateOut>(`/xray/clients/${clientId}`);
}

export async function getXrayLink(clientId: string): Promise<ConnectionLinkOut> {
  return request<ConnectionLinkOut>(`/xray/clients/link/${clientId}`);
}

export async function updateVlessClientAlias(clientId: number, data: UpdateClientAliasIn): Promise<ClientPublicOut> {
  return request<ClientPublicOut>(`/xray/clients/${clientId}/alias`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}

export async function deleteVlessClient(clientId: number): Promise<ClientPublicOut> {
  return request<ClientPublicOut>(`/xray/clients/${clientId}`, {
    method: 'DELETE',
  });
}

// --- Hysteria Clients ---

export async function getHysteriaClientsByUserId(userId: number): Promise<ClientPublicOut[]> {
  return request<ClientPublicOut[]>(`/hysteria/clients/by_user_id/${userId}`);
}

export async function getHysteriaClientById(clientId: number): Promise<HysteriaClientPrivateOut> {
  return request<HysteriaClientPrivateOut>(`/hysteria/clients/${clientId}`);
}
export async function getHysteriaLink(clientId: number): Promise<ConnectionLinkOut> {
  return request<ConnectionLinkOut>(`/hysteria/clients/link/${clientId}`);
}

export async function updateHysteriaClientAlias(clientId: number, data: UpdateClientAliasIn): Promise<ClientPublicOut> {
  return request<ClientPublicOut>(`/hysteria/clients/${clientId}/alias`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}

export async function deleteHysteriaClient(clientId: number): Promise<ClientPublicOut> {
  return request<ClientPublicOut>(`/hysteria/clients/${clientId}`, {
    method: 'DELETE',
  });
}

// --- Invites ---

export async function newInvite(data: InviteIn): Promise<InviteOut> {
  return request<InviteOut>('/invite/new', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function activateInvite(data: InviteActivateIn): Promise<InviteCheckOut> {
  return request<InviteCheckOut>('/invite/activate', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}
