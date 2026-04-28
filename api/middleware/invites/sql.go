package invites

var (
	NewInviteQuery = `
	INSERT INTO public.invnites (invite_hash, vpn_type, expires_at) 
	SELECT 
	    $1 AS invite_hash,
	    $2 AS vpn_type,
	    IF(
	    	$3::timestampz > now(), 
	    	$3::timestampz, 
	    	now() + interval '1 month'
	    ) AS expires_at
	RETURNING id, created_at, updated_at, expires_at, status, vpn_type
	`

	GetInviteInfoQuery = `
	SELECT id, vpn_type
	FROM public.invites
	WHERE invite_hash = $1 AND status = 'active'
	ORDER BY updated_at ASC
	LIMIT 1
	`

	ActivateInviteQuery = `
	UPDATE public.invites
	SET status = 'activated', updated_at = now()
	WHERE id = $1
	`
)
