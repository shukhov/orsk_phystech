// --- Auth ---

export interface UserLoginIn {
  email: string;
  password: string;
}

export interface UserRegisterIn {
  email: string;
  password: string;
  username: string;
}

export interface AuthToken {
  token: string;
}

// --- User ---

export interface UserPublicOut {
  id: number;
  created_at: string;
  updated_at: string;
  username: string;
}

export interface UserPrivateOut extends UserPublicOut {
  email: string;
  status: string;
  role_id: number;
}

export interface Role {
  id: number;
  role_name: string;
  access_level: number;
}

// --- VLESS Client ---

export interface ClientPublicOut {
  id: number;
  alias: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ClientPrivateOut extends ClientPublicOut {
  access_key: string;
  user_id: number;
  invite_id: number;
}

export interface UpdateClientAliasIn {
  new_alias: string;
}

export interface ConnectionLinkOut {
  connection_link: string;
}

// --- Invite ---

export interface InviteIn {
  invite_word: string;
  vpn_type: string;
  expires_at?: string;
}

export interface InviteActivateIn {
  invite_word: string;
  alias: string;
}

export interface InviteOut {
  id: number;
  created_at: string;
  updated_at: string;
  expires_at: string;
  status: string;
  vpn_type: string;
}

export interface InviteCheckOut {
  id: number;
  alias: string;
  vpn_type: string;
}

// --- API Error ---

export interface ErrorCallback {
  error_text: string;
}