package hysteria

var (
	NewClientQuery = `
		INSERT INTO public.hysteria_clients (invite_id, user_id, alias)
		VALUES ($1, $2, $3)
		RETURNING id, alias, status, created_at, updated_at
		`

	GetHysteriaLinkByIdQuery = `
	SELECT
	    user_id::text || '-' || id::text AS user,
		password,
		alias
	FROM public.hysteria_clients
	WHERE id = $1 AND user_id = $2
	`

	GetAllInConfigClientsQuery = `
	SELECT
	    hc.user_id::text || '-' || hc.id::text AS user,
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
)
