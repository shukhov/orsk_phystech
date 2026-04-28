
CREATE TABLE IF NOT EXISTS public.roles (
    id BIGSERIAL PRIMARY KEY,
    role_name TEXT UNIQUE NOT NULL,
    access_level INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS public.users (
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    username      TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    email         TEXT UNIQUE NOT NULL,

    role_id       BIGINT NOT NULL DEFAULT 1 REFERENCES public.roles(id),

    status        TEXT NOT NULL DEFAULT 'active'
)
;

CREATE TABLE IF NOT EXISTS public.invites (
    id BIGSERIAL PRIMARY KEY,

    invite_hash TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + interval '1 month',

    status TEXT NOT NULL DEFAULT 'active' CHECK ( status in ('active', 'activated', 'expired', 'revoked') ),

    vpn_type TEXT NOT NULL DEFAULT 'xray'
);

CREATE TABLE IF NOT EXISTS public.vless_clients (
    id         BIGSERIAL PRIMARY KEY,
    access_id  UUID DEFAULT gen_random_uuid(),
    user_id    BIGINT NOT NULL REFERENCES public.users(id),
    invite_id  BIGINT NOT NULL UNIQUE REFERENCES public.invites(id),

    alias      TEXT NOT NULL,

    status     TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);



-------

INSERT INTO public.roles (id, role_name, access_level)
VALUES (1, 'user', 1),
       (2, 'vpn_publisher', 2),
       (3, 'network_admin', 3),
       (4, 'service_admin', 4),
       (5, 'superuser', 5)
ON CONFLICT
DO NOTHING;

