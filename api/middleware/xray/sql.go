package xray

var (
	GetClientsByUserIdQuery = `
	SELECT 
	    id, 
	    alias,
	    status, 
	    created_at, 
	    updated_at
	FROM public.vless_clients 
	WHERE user_id = $1`

	GetClientByIdQuery = `
	SELECT 
	    id, 
	    access_key, 
	    user_id, 
	    invite_id, 
	    alias,
	    status, 
	    created_at, 
	    updated_at	
    FROM public.vless_clients WHERE id = $1`

	GetAllInConfigClientsQuery = `
	SELECT
	    vc.access_key AS id,
	    'xtls-rprx-vision' AS flow
	FROM public.vless_clients AS vc
	JOIN public.users AS us
		ON vc.user_id=us.id
	WHERE 
	    vc.status='active' 
	  AND us.status='active'`

	NewClientQuery = `
	INSERT INTO public.vless_clients (invite_id, user_id, alias)
	VALUES ($1, $2, $3)
	RETURNING id, alias, status, created_at, updated_at
	`

	GetXrayLinkByIdQuery = `
	SELECT access_key, alias
	FROM public.vless_clients
	WHERE id = $1 AND user_id = $2
	`

	UpdateClientAliasQuery = `
	UPDATE public.vless_clients
	SET alias = $1, updated_at = now()
	WHERE id = $2
	RETURNING id, alias, status, created_at, updated_at
	`

	LastUpdateQuery = `
	SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz) AS updated_at
	FROM public.vless_clients
	`
)
