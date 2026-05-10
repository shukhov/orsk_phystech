package security

var (
	GetUserByIdQuery = `
	SELECT id, created_at, updated_at, status 
	FROM public.users 
	WHERE id = $1`

	MeQuery = `
	SELECT id, created_at, updated_at, username, email, status, role_id 
	FROM public.users WHERE id = $1
	`

	AllowForRoleQuery = `
	SELECT user_role.access_level >= role_req.access_level AS has_access
	FROM public.users AS usr
	JOIN public.roles AS user_role ON usr.role_id = user_role.id
	JOIN public.roles AS role_req ON role_req.id = $1
	WHERE usr.id = $2;`

	SetRoleForUserQuery = `
	UPDATE public.users SET role_id=$1, updated_at=now() WHERE id = $2 
	RETURNING id, created_at, updated_at, username, email, status, role_id;
	`

	GetRoleByIdQuery = `
    SELECT id, role_name, access_level FROM public.roles WHERE id = $1
 	`

	CreateUserQuery = `
	INSERT INTO public.users (username, password_hash, email)
	VALUES ($1, $2, $3)
	RETURNING id, username, created_at, updated_at;`

	LoginQuery = `SELECT id, password_hash FROM public.users WHERE email = $1`
)
