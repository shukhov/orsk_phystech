package hysteria

var (
	NewClientQuery = `
		INSERT INTO public.hysteria_clients (invite_id, user_id, alias)
		VALUES ($1, $2, $3)
		RETURNING id, alias, status, created_at, updated_at
		`

	GetClientByIdQuery = `
	SELECT
	    id,
		password,
	    user_id,
	    invite_id,
	    alias,
	    status,
	    created_at,
	    updated_at
	FROM public.hysteria_clients 
	WHERE id = $1`

	GetClientsByUserIdQuery = `
	SELECT
	    id,
	    alias,
	    status,
	    created_at,
	    updated_at
	FROM public.hysteria_clients
	WHERE user_id = $1 AND status != 'deleted'`

	GetHysteriaLinkByIdQuery = `
	SELECT
		password,
		alias
	FROM public.hysteria_clients
	WHERE id = $1 AND user_id = $2
	`

	GetAllInConfigClientsQuery = `
	SELECT
		password
	FROM public.hysteria_clients AS hc
	JOIN public.users AS us
		ON hc.user_id=us.id
	WHERE
	    hc.status='active'
	  AND us.status='active'`

	UpdateClientAliasQuery = `
	UPDATE public.hysteria_clients
	SET alias = $1, updated_at = now()
	WHERE id = $2
	RETURNING id, alias, status, created_at, updated_at
	`

	DeleteClientByIdQuery = `
	UPDATE public.hysteria_clients
	SET status = 'deleted', updated_at = now()
	WHERE id = $1
	RETURNING id, alias, status, created_at, updated_at
	`

	LastUpdateQuery = `
	SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz) AS updated_at
	FROM public.hysteria_clients
	`
)
