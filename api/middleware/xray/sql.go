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
	    access_id, 
	    user_id, 
	    invite_id, 
	    alias,
	    status, 
	    created_at, 
	    updated_at	
    FROM public.vless_clients WHERE id = $1`

	GetAllInConfigClientsQuery = `
	SELECT 
	    vc.access_id AS id, 
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
)
